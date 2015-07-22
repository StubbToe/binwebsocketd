// Copyright 2013 Joe Walnes and the websocketd team.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libwebsocketd

import (
	"io"
	"os"
	"os/exec"
	"sync"
)

type closeOnce struct {
	*os.File

	once sync.Once
	err  error
}

type LaunchedProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	binin  *os.File
	binout *os.File
}

func launchCmd(commandName string, commandArgs []string, env []string) (*LaunchedProcess, error) {
	cmd := exec.Command(commandName, commandArgs...)
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	binin, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	pr, binout, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	cmd.ExtraFiles = []*os.File{pr, pw}

	err = cmd.Start()

	if err != nil {
		return nil, err
	}

	// get rid of extra handles
	pr.Close()
	pw.Close()

	return &LaunchedProcess{cmd, stdin, stdout, stderr, binin, binout}, err
}
