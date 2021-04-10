// Copyright 2021 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package invoke provides support for invoking a command.  It wraps the
// standard os/exec package, returning a custom error type that will
// additionally report the command arguments and the entire content of the
// invoked command stderr.
package invoke

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var Command = exec.Command

// Error is the error returned when a command returns an error.
type Error struct {
	Cmd    string   // the command invoked
	Argv   []string // arguments to the command
	Stderr []byte   // the entire content of the command stderr
	Err    error    // the original error from os/exec.Command.Run
}

// Error implements the error interface.
func (e *Error) Error() string {
	argv := strings.Trim(fmt.Sprint(e.Argv), "[]")
	stderr := string(e.Stderr)
	msg := e.Cmd
	if argv != "" {
		msg += " " + argv
	}
	msg += ": " + e.Err.Error()

	if stderr == "" {
		return msg
	}

	return msg + ": " + stderr
}

// Unwrap implements the Wrapper interface.
func (e *Error) Unwrap() error {
	return e.Err
}

// Run runs cmd.
//
// In case the command exits with a non 0 exit status, the error will contain
// the entire content of the command stderr, with whitespace trimmed.
func Run(cmd *exec.Cmd) error {
	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		err := &Error{
			Cmd:    cmd.Path,
			Argv:   cmd.Args[1:],
			Stderr: normalize(stderr),
			Err:    err,
		}

		return err
	}

	return nil
}

// Output invokes cmd and returns the stdout content, with whitespace trimmed.
//
// In case the command exits with a non 0 exit status, the error will contain
// the entire content of the command stderr, with whitespace trimmed.
func Output(cmd *exec.Cmd) ([]byte, error) {
	if cmd.Stdout != nil {
		return nil, errors.New("invoke: Stdout already set")
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		err := &Error{
			Cmd:    cmd.Path,
			Argv:   cmd.Args[1:],
			Stderr: normalize(stderr),
			Err:    err,
		}

		return normalize(stdout), err
	}

	return normalize(stdout), nil
}

// normalize returns the data buffered in b with leading and trailing white
// space removed.
func normalize(b *bytes.Buffer) []byte {
	return bytes.TrimSpace(b.Bytes())
}
