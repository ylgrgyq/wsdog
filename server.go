package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{} // use default options

func closeConn(conn *websocket.Conn) {
	// failed to send the last close message is tolerable due to the connection may broken
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		wsdogLogger.Debugf("write close message failed: %s", err.Error())
	}

	if err := conn.Close(); err != nil {
		wsdogLogger.Debugf("close websocket connection failed: %s", err.Error())
	}
}

func generateWsHandler(opts CommandLineOptions) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			wsdogLogger.Errorf("websocket upgrade failed: %s", err.Error())
			return
		}

		readWsChan, readFromConnDone := SetupReadFromConn(conn, opts.ShowPingPong)
		defer closeConn(conn)
		for {
			select {
			case <-readFromConnDone:
				return
			case message := <-readWsChan:
				wsdogLogger.ReceiveMessagef("< %s", message.payload)
				if message.messageType == websocket.TextMessage {
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

func runAsServer(listenPort uint16, opts CommandLineOptions) {
	http.HandleFunc("/", generateWsHandler(opts))

	wsdogLogger.Okf("listening on port %d (press CTRL+C to quit)", listenPort)
	wsdogLogger.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", listenPort), nil))
}
