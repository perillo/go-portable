// Copyright 2021 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package invoke

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

// TestRun tests the Run function by executing a temporary shell script.
func TestRun(t *testing.T) {
	const stderr = "hello stderr"

	name := tempScript(t)
	argv := []string{"-a", "b"}
	cmd := exec.Command(name, argv...)

	err := Run(cmd)
	if err == nil {
		t.Fatal("expected err != nil")
	}
	validate(t, err, name, argv, stderr)
}

// TestOutput tests the Output function by executing a temporary shell script.
func TestOutput(t *testing.T) {
	const stdout = "hello stdout"
	const stderr = "hello stderr"

	name := tempScript(t)
	argv := []string{"-a", "b"}
	cmd := exec.Command(name, argv...)

	data, err := Output(cmd)
	if err == nil {
		t.Fatal("expected err != nil")
	}
	if string(data) != stdout {
		t.Errorf("want data = %s, got %s", stdout, data)
	}
	validate(t, err, name, argv, stderr)
}

// validate validates the error returned by Run or Output.
func validate(t *testing.T, err error, name string, argv []string, stderr string) {
	var eerr *exec.ExitError

	e := err.(*Error)
	if !errors.As(e.Err, &eerr) {
		t.Fatalf("expected e.Err as %T, got %T", eerr, e.Err)
	}
	if e.Cmd != name {
		t.Errorf("want e.Cmd = %s, got %s", name, e.Cmd)
	}
	if !reflect.DeepEqual(e.Argv, argv) {
		t.Errorf("want e.Argv = %q, got %q", argv, e.Argv)
	}
	if string(e.Stderr) != stderr {
		t.Errorf("want e.Stderr = %s, got %s", stderr, e.Stderr)
	}
}

// tempScript creates a temporary shell script that writes "hello stdout" on
// stdout and "hello stderr" on stderr with additional whitespace, and exits
// with exit status 1.
//
// tempScript currently only support UNIX systems.
func tempScript(t *testing.T) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sh")

	code := `#!/bin/sh
printf "\thello stdout\n" >&1
printf "\thello stderr\n" >&2
exit 1
`
	if err := os.WriteFile(path, []byte(code), 0o700); err != nil {
		t.Fatalf("tempscript: %v", err)
	}

	return path
}
