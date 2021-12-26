package main

import (
	"encoding/base64"
	"github.com/gorilla/websocket"
	"net"
	"time"
)

type WebSocketMessage struct {
	messageType int
	payload     []byte
}

func setupPingPongHandler(conn *websocket.Conn, output chan WebSocketMessage) {
	pingHandler := func(message string) error {
		wsdogLogger.Ok("Receive Ping frame")
		err := conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(defaultWriteWaitDuration))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}
		return err
	}

	pongHandler := func(message string) error {
		wsdogLogger.Ok("Receive Pong frame")
		return nil
	}

	conn.SetPingHandler(pingHandler)
	conn.SetPongHandler(pongHandler)
}

func setupCloseHandler(conn *websocket.Conn) {
	conn.SetCloseHandler(func(code int, text string) error {
		wsdogLogger.Okf("Receive close frame (code: %d, reason %s)", code, text)
		return &websocket.CloseError{Code: code, Text: text}
	})
}

func SetupReadFromConn(conn *websocket.Conn, showPingPong bool) (chan WebSocketMessage, chan struct{}) {
	done := make(chan struct{})
	output := make(chan WebSocketMessage)
	if showPingPong {
		setupPingPongHandler(conn, output)
	}
	setupCloseHandler(conn)
	go func() {
		defer close(output)
		for {
			select {
			case <- done:
				wsdogLogger.Okf("Disconnected")
				return
			default:
				mt, message, err := conn.ReadMessage()
				if err != nil {
					closeErr, ok := err.(*websocket.CloseError)
					if ok {
						wsdogLogger.Okf("Disconnected (code: %d, reason: \"%s\")", closeErr.Code, closeErr.Text)
						return
					} else {
						wsdogLogger.Debugf("error: %s", err.Error())
						continue
					}
				}
				output <- WebSocketMessage{mt, message}
			}

		}
	}()
	return output, done
}

func PrintReceivedMessage(message *WebSocketMessage) {
	switch message.messageType {
	case websocket.TextMessage:
		wsdogLogger.ReceiveMessagef("< %s", message.payload)
	case websocket.BinaryMessage:
		sEnc := base64.StdEncoding.EncodeToString(message.payload)
		wsdogLogger.ReceiveMessagef("<< %s", sEnc)
	}
}
