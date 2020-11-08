package main

import (
	"github.com/gorilla/websocket"
	"net"
	"time"
)

type WebSocketMessage struct {
	messageType int
	payload     []byte
}

func SetupPingPongHandler(conn *websocket.Conn, output chan WebSocketMessage) {
	pingHandler := func(message string) error {
		output <- WebSocketMessage{websocket.PingMessage, []byte("Received ping")}
		err := conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(defaultWriteWaitDuration))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}
		return err
	}

	pongHandler := func(message string) error {
		output <- WebSocketMessage{websocket.PongMessage, []byte("Received pong")}
		return nil
	}

	conn.SetPingHandler(pingHandler)
	conn.SetPongHandler(pongHandler)

}

func SetupReadFromConn(conn *websocket.Conn, showPingPong bool) (chan WebSocketMessage, chan struct{}) {
	done := make(chan struct{})
	output := make(chan WebSocketMessage)
	if showPingPong {
		SetupPingPongHandler(conn, output)
	}

	go func() {
		defer close(done)
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				closeErr, ok := err.(*websocket.CloseError)
				if ok {
					wsdogLogger.Okf("Disconnected (code: %d, reason: \"%s\")", closeErr.Code, closeErr.Text)
				} else {
					wsdogLogger.Errorf("error: %s", err.Error())
				}
				return
			}

			output <- WebSocketMessage{mt, message}
		}
	}()
	return output, done
}
