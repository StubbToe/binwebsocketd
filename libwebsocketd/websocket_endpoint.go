// Copyright 2013 Joe Walnes and the websocketd team.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libwebsocketd

import (
	"encoding/binary"
	"golang.org/x/net/websocket"
	"io"
)

type WebSocketEndpoint struct {
	ws     *websocket.Conn
	output chan string
	binOut chan []byte
	log    *LogScope
}

func NewWebSocketEndpoint(ws *websocket.Conn, log *LogScope) *WebSocketEndpoint {
	return &WebSocketEndpoint{
		ws:     ws,
		output: make(chan string),
		binOut: make(chan []byte),
		log:    log}
}

func (we *WebSocketEndpoint) Terminate() {
}

func (we *WebSocketEndpoint) BinOutput() chan []byte {
	return we.binOut
}

func (we *WebSocketEndpoint) Output() chan string {
	return we.output
}

func (we *WebSocketEndpoint) Send(msg string) bool {
	err := websocket.Message.Send(we.ws, msg)
	if err != nil {
		we.log.Trace("websocket", "Cannot send: %s", err)
		return false
	}
	return true
}

func (we *WebSocketEndpoint) SendBinary(data []byte) bool {
	we.log.Error("websocket", "Sending %s bytes", len(data))
	err := websocket.Message.Send(we.ws, data)
	if err != nil {
		we.log.Trace("websocket", "Cannot send: %s", err)
		return false
	}
	return true
}

func (we *WebSocketEndpoint) StartReading() {
	go we.read_client()
}

func (we *WebSocketEndpoint) read_client() {
	for {
		var msg string
		err := websocket.Message.Receive(we.ws, &msg)
		if err != nil {
			if err != io.EOF {
				we.log.Debug("websocket", "Cannot receive: %s", err)
			}
			break
		}

		if len(msg) > 0 {
			byteArray := []byte(msg)
			btype := byteArray[0]

			if btype > 0 {
				if len(byteArray) > 1 {
					byteArray = byteArray[1:len(byteArray)]
				}

				dataLength := uint32(len(byteArray))
				dataOut := make([]byte, 4)
				binary.BigEndian.PutUint32(dataOut, dataLength)
				dataOut = append(dataOut, btype)
				dataOut = append(dataOut, byteArray...)
				we.log.Error("websocket", "Read binary data type: %d, length: %d, actual: %d", btype, dataLength, len(dataOut))
				we.binOut <- dataOut
			} else {
				we.output <- msg[1:len(msg)]
			}

		}
	}
	close(we.binOut)
	close(we.output)
}
