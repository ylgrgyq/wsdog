package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

func closeConn(conn *websocket.Conn) {
	if err := conn.Close(); err != nil {
		wsdogLogger.Debugf("close websocket connection failed: %s", err.Error())
	}
}

func generateWsHandler(opts CommandLineOptions) func(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{Subprotocols: []string{opts.Subprotocol}, HandshakeTimeout: defaultHandshakeTimeout}
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			wsdogLogger.Errorf("websocket upgrade failed: %s", err.Error())
			return
		}

		wsdogLogger.Ok("Client connected")

		readWsChan, readFromConnDone := SetupReadFromConn(conn, opts.ShowPingPong)
		defer closeConn(conn)
		for {
			select {
			case <-readFromConnDone:
				return
			case message := <-readWsChan:
				PrintReceivedMessage(&message)

				if opts.Echo {
					err = conn.WriteMessage(message.messageType, message.payload)
					if err != nil {
						wsdogLogger.Errorf("error: %s", err)
						return
					}
				}
			}
		}
	}
}

func RunAsServer(listenPort uint16, opts CommandLineOptions) {
	http.HandleFunc("/", generateWsHandler(opts))

	wsdogLogger.Okf("Listening on port %d (press CTRL+C to quit)", listenPort)
	wsdogLogger.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", opts.ListenHost, listenPort), nil))
}
