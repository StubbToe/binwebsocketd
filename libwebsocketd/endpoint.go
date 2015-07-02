// Copyright 2013 Joe Walnes and the websocketd team.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libwebsocketd

type Endpoint interface {
	StartReading()
	Terminate()
	Output() chan string
	BinOutput() chan []byte
	Send(string) bool
	SendBinary([]byte) bool
}

func PipeEndpoints(e1, e2 Endpoint) {
	e1.StartReading()
	e2.StartReading()

	defer e1.Terminate()
	defer e2.Terminate()
	for {
		select {
		case msgOneB, ok := <-e1.BinOutput():
			if len(msgOneB) == 0 {
				return
			}

			if !ok || !e2.SendBinary(msgOneB) {
				return
			}
		case msgOne, ok := <-e1.Output():
			if len(msgOne) == 0 {
				return
			}

			if !ok || !e2.Send(msgOne) {
				return
			}
		case msgTwoB, ok := <-e2.BinOutput():
			if len(msgTwoB) == 0 {
				return
			}

			if !ok || !e1.SendBinary(msgTwoB) {
				return
			}
		case msgTwo, ok := <-e2.Output():
			if len(msgTwo) == 0 {
				return
			}

			if !ok || !e1.Send(msgTwo) {
				return
			}
		}
	}
}
