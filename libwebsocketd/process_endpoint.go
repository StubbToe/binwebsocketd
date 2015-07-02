// Copyright 2013 Joe Walnes and the websocketd team.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libwebsocketd

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"syscall"
	"time"
)

type ProcessEndpoint struct {
	process    *LaunchedProcess
	bufferedIn *bufio.Writer
	binBufIn   *bufio.Reader
	binBufOut  *bufio.Writer
	output     chan string
	binOut     chan []byte
	log        *LogScope
}

func NewProcessEndpoint(process *LaunchedProcess, log *LogScope) *ProcessEndpoint {
	return &ProcessEndpoint{
		process:    process,
		bufferedIn: bufio.NewWriter(process.stdin),
		binBufIn:   bufio.NewReader(process.binin),
		binBufOut:  bufio.NewWriter(process.binout),
		output:     make(chan string),
		binOut:     make(chan []byte),
		log:        log}
}

func (pe *ProcessEndpoint) Terminate() {

	terminated := make(chan struct{})
	go func() { pe.process.cmd.Wait(); terminated <- struct{}{} }()

	// for some processes this is enough to finish them...
	pe.process.stdin.Close()
	pe.process.binin.Close()
	pe.process.binout.Close()

	// a bit verbose to create good debugging trail
	select {
	case <-terminated:
		pe.log.Debug("process", "Process %v terminated after stdin was closed", pe.process.cmd.Process.Pid)
		return // means process finished
	case <-time.After(100 * time.Millisecond):
	}

	err := pe.process.cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		// process is done without this, great!
		pe.log.Error("process", "SIGINT unsuccessful to %v: %s", pe.process.cmd.Process.Pid, err)
	}

	select {
	case <-terminated:
		pe.log.Debug("process", "Process %v terminated after SIGINT", pe.process.cmd.Process.Pid)
		return // means process finished
	case <-time.After(250 * time.Millisecond):
	}

	err = pe.process.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		// process is done without this, great!
		pe.log.Error("process", "SIGTERM unsuccessful to %v: %s", pe.process.cmd.Process.Pid, err)
	}

	select {
	case <-terminated:
		pe.log.Debug("process", "Process %v terminated after SIGTERM", pe.process.cmd.Process.Pid)
		return // means process finished
	case <-time.After(500 * time.Millisecond):
	}

	err = pe.process.cmd.Process.Kill()
	if err != nil {
		pe.log.Error("process", "SIGKILL unsuccessful to %v: %s", pe.process.cmd.Process.Pid, err)
		return
	}

	select {
	case <-terminated:
		pe.log.Debug("process", "Process %v terminated after SIGKILL", pe.process.cmd.Process.Pid)
		return // means process finished
	case <-time.After(1000 * time.Millisecond):
	}

	pe.log.Error("process", "SIGKILL did not terminate %v!", pe.process.cmd.Process.Pid)
}

func (pe *ProcessEndpoint) BinOutput() chan []byte {
	return pe.binOut
}

func (pe *ProcessEndpoint) Output() chan string {
	return pe.output
}

func (pe *ProcessEndpoint) SendBinary(data []byte) bool {
	pe.binBufOut.Write(data)
	pe.binBufOut.Flush()
	return true
}

func (pe *ProcessEndpoint) Send(msg string) bool {
	pe.bufferedIn.WriteString(msg)
	pe.bufferedIn.WriteString("\n")
	pe.bufferedIn.Flush()
	return true
}

func (pe *ProcessEndpoint) StartReading() {
	go pe.log_stderr()
	go pe.process_stdout()
	go pe.process_binout()
}

func (pe *ProcessEndpoint) process_stdout() {
	bufin := bufio.NewReader(pe.process.stdout)
	for {
		str, err := bufin.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				pe.log.Error("process", "Unexpected error while reading STDOUT from process: %s", err)
			} else {
				pe.log.Debug("process", "Process STDOUT closed")
			}
			break
		}
		pe.output <- trimEOL("\x00" + str)
	}
	close(pe.output)
}

func (pe *ProcessEndpoint) process_binout() {
	lenb := make([]byte, 4, 4)

	for {
		n, err := pe.binBufIn.Read(lenb)
		buf := bytes.NewBuffer(lenb)
		if err != nil {
			pe.log.Error("process", "Unexpected error while reading BINOUT from process: %s", err)
			return
		}

		if n > 0 {
			var dataLength uint32
			binary.Read(buf, binary.LittleEndian, &dataLength)
			dataOut := make([]byte, dataLength+1, dataLength+1)
			n2, err2 := pe.binBufIn.Read(dataOut)
			if err2 != nil {
				pe.log.Error("process", "Unexpected error while reading BINOUT from process: %s", err2)
				return
			}

			if n2 > 0 {
				pe.binOut <- dataOut
			}
		}
	}
	close(pe.binOut)
}

func (pe *ProcessEndpoint) log_stderr() {
	bufstderr := bufio.NewReader(pe.process.stderr)
	for {
		str, err := bufstderr.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				pe.log.Error("process", "Unexpected error while reading STDERR from process: %s", err)
			} else {
				pe.log.Debug("process", "Process STDERR closed")
			}
			break
		}
		pe.log.Error("stderr", "%s", trimEOL(str))
	}
}

// trimEOL cuts unixy style \n and windowsy style \r\n suffix from the string
func trimEOL(s string) string {
	lns := len(s)
	if lns > 0 && s[lns-1] == '\n' {
		lns--
		if lns > 0 && s[lns-1] == '\r' {
			lns--
		}
	}
	return s[0:lns]
}
