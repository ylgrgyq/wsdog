package main

import "time"

const defaultHandshakeTimeout = 5 * time.Second
const defaultWriteWaitDuration = 5 * time.Second
const defaultReadWaitDuration = 5 * time.Second
const defaultCloseStatusCode = 1000
const defaultCloseReason = ""
const subprotocolHeader = "Sec-WebSocket-Protocol"
