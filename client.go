package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func parseConnectUrl(urlStr string) *url.URL {
	connectUrl, err := url.Parse(urlStr)
	if err != nil {
		wsdogLogger.Fatalf("\"%s\" is not a valid url", urlStr)
	}

	if connectUrl.Scheme == "" {
		wsdogLogger.Fatalf("missing scheme in url: \"%s\" to connect", urlStr)
	}

	if connectUrl.Host == "" {
		wsdogLogger.Fatalf("missing host in url: \"%s\" to connect", urlStr)
	}

	if strings.HasPrefix(connectUrl.Scheme, "http") {
		connectUrl.Scheme = strings.Replace(connectUrl.Scheme, "http", "ws", 1)
	}

	if connectUrl.Scheme != "wss" && connectUrl.Scheme != "ws" {
		wsdogLogger.Fatalf("malformed scheme in url: \"%s\" to connect", urlStr)
	}

	return connectUrl
}

func newDialer(cliOpts CommandLineOptions) websocket.Dialer {
	var tlsConfig = tls.Config{}
	if cliOpts.NoTlsCheck {
		tlsConfig.InsecureSkipVerify = true
	}
	return websocket.Dialer{
		TLSClientConfig:  &tlsConfig,
		Subprotocols:     []string{cliOpts.Subprotocol},
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: defaultHandshakeTimeout,
	}
}

func buildConnectHeaders(cliOpts CommandLineOptions) http.Header {
	headers := http.Header{}
	if len(cliOpts.Origin) > 0 {
		headers["Origin"] = []string{cliOpts.Origin}
	}

	if len(cliOpts.Host) > 0 {
		headers["Host"] = []string{cliOpts.Host}
	}

	if len(cliOpts.Headers) > 0 {
		for k, v := range cliOpts.Headers {
			headers[k] = []string{v}
		}
	}

	if len(cliOpts.Auth) > 0 {
		auth := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(cliOpts.Auth)))
		headers["Authorization"] = []string{auth}
	}
	return headers
}

type Client struct {
	conn           *websocket.Conn
	readWsChan     chan WebSocketMessage
	readWsDoneChan chan struct{}
	enableSlash    bool
	closed         uint32
}

type CommandType string

const (
	PingCommand   CommandType = "ping"
	PongCommand               = "pong"
	BinaryCommand             = "binary"
	TextCommand               = "text"
	CloseCommand              = "close"
)

type ConsoleCommand struct {
	command   CommandType
	parameter string
}

func (client *Client) doWriteMessage(messageType int, message []byte) {
	if err := client.conn.SetWriteDeadline(time.Now().Add(defaultWriteWaitDuration)); err != nil {
		panic(err)
	}

	err := client.conn.WriteMessage(messageType, message)
	if err != nil {
		panic(err)
	}
}

func parseConsoleCommand(input string, enableSlash bool) (*ConsoleCommand, error) {
	if !enableSlash || input[0:1] != "/" {
		return &ConsoleCommand{command: "text", parameter: input}, nil
	}

	slashInput := input[1:]
	if len(slashInput) <= 0 {
		return nil, fmt.Errorf("invalid slash command: %s", slashInput)
	}
	parsed := strings.SplitN(slashInput, " ", 2)
	if len(parsed) > 1 {
		return &ConsoleCommand{command: CommandType(parsed[0]), parameter: parsed[1]}, nil
	}
	return &ConsoleCommand{command: CommandType(parsed[0]), parameter: ""}, nil
}

func (client *Client) writeMessage(input string) bool {
	slashCmd, err := parseConsoleCommand(input, client.enableSlash)
	if err != nil {
		wsdogLogger.Errorf("invalid slash command. %s", err.Error())
		return false
	}
	switch slashCmd.command {
	case PingCommand:
		client.doWriteMessage(websocket.PingMessage, nil)
	case PongCommand:
		client.doWriteMessage(websocket.PongMessage, nil)
	case TextCommand:
		client.doWriteMessage(websocket.TextMessage, []byte(slashCmd.parameter))
	case BinaryCommand:
		if len(slashCmd.parameter) == 0 {
			break
		}
		sDec, err := base64.StdEncoding.DecodeString(slashCmd.parameter)
		if err != nil {
			wsdogLogger.Errorf("invalid string in Base64: \"%s\"", slashCmd.parameter)
			break
		}
		client.doWriteMessage(websocket.BinaryMessage, sDec)
	case CloseCommand:
		statusCode := defaultCloseStatusCode
		reason := defaultCloseReason
		re := regexp.MustCompile("\\s+")
		toks := re.Split(input, -1)
		if len(toks) >= 2 {
			var err error
			if statusCode, err = strconv.Atoi(toks[1]); err != nil {
				wsdogLogger.Errorf("invalid close status code: \"%s\"", toks[1])
				break
			}
		}

		if len(toks) >= 3 {
			reason = strings.Join(toks[2:], " ")
		}

		message := websocket.FormatCloseMessage(statusCode, reason)
		client.doWriteMessage(websocket.CloseMessage, message)
		client.close()
		return true
	default:
		wsdogLogger.Errorf("unknown slash command: \"%s\"", slashCmd.command)
	}
	return false
}

func (client *Client) executeCommandThenShutdown(cliOpts CommandLineOptions) {
	client.writeMessage(cliOpts.ExecuteCommand)

	timout := time.Second * time.Duration(cliOpts.Wait)
	ticker := time.NewTicker(timout)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		case <-ticker.C:
			return
		case message, ok := <-client.readWsChan:
			if !ok {
				return
			}
			PrintReceivedMessage(&message)
		case <-interrupt:
			return
		}
	}
}

func (client *Client) loopExecuteCommandFromConsole(cliOpts CommandLineOptions) {
	consoleReader := NewConsoleInputReader()
	defer consoleReader.Close()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		case <-consoleReader.done:
			return
		case output := <-consoleReader.outputChan:
			if len(output) > 0 {
				if client.writeMessage(output) {
					return
				}
			}
		case message, ok := <-client.readWsChan:
			if !ok {
				return
			}
			consoleReader.Clean()
			PrintReceivedMessage(&message)
			consoleReader.Refresh()
		case <-interrupt:
			return
		}
	}
}

func (client *Client) run(cliOpts CommandLineOptions) {
	if len(cliOpts.ExecuteCommand) > 0 {
		client.executeCommandThenShutdown(cliOpts)
	} else {
		client.loopExecuteCommandFromConsole(cliOpts)
	}
}

func (client *Client) close() {
	if !atomic.CompareAndSwapUint32(&client.closed, 0, 1) {
		return
	}
	close(client.readWsDoneChan)
	if err := client.conn.Close(); err != nil {
		wsdogLogger.Debugf("close client failed: %s", err.Error())
	}
}

func (client *Client) gracefulClose() {
	if !atomic.CompareAndSwapUint32(&client.closed, 0, 1) {
		return
	}
	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection.
	client.doWriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	close(client.readWsDoneChan)
	if err := client.conn.Close(); err != nil {
		wsdogLogger.Debugf("close client failed: %s", err.Error())
	}
}

func checkResponseSubprotocol(requiredProtocol string, resp *http.Response) {
	subProtoHeader := resp.Header[http.CanonicalHeaderKey(subprotocolHeader)]
	if len(subProtoHeader) < 1 || subProtoHeader[0] != requiredProtocol {
		wsdogLogger.Fatal("error: Server sent no subprotocol")
	}
}

func RunAsClient(url string, cliOpts CommandLineOptions) {
	connectUrl := parseConnectUrl(url)

	dialer := newDialer(cliOpts)
	headers := buildConnectHeaders(cliOpts)

	conn, resp, err := dialer.Dial(connectUrl.String(), headers)
	if err != nil {
		wsdogLogger.Fatalf("connect to \"%s\" failed with error: \"%s\"", connectUrl, err)
	}

	if len(cliOpts.Subprotocol) > 0 {
		checkResponseSubprotocol(cliOpts.Subprotocol, resp)
	}

	wsdogLogger.Ok("Connected (press CTRL+C to quit)")

	readWsChan, readWsDoneChan := SetupReadFromConn(conn, cliOpts.ShowPingPong)
	client := Client{conn, readWsChan, readWsDoneChan, cliOpts.EnableSlash, 0}
	defer client.gracefulClose()
	client.run(cliOpts)
}
