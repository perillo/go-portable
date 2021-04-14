// Copyright 2021 Manlio Perillo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/perillo/go-portable/internal/invoke"
)

var (
	// gocmd is the go command to use.  It can be overridden using the GOCMD
	// environment variable.
	gocmd = "go"

	// gocmdpath is the resolved path to the go command.
	gocmdpath string
)

type platform struct {
	os   string
	arch string
}

func init() {
	if value, ok := os.LookupEnv("GOCMD"); ok {
		gocmd = value
	}
	gocmdpath = gocmd // set default value

	// Don't report the error now.
	if path, err := exec.LookPath(gocmd); err == nil {
		gocmdpath = path
	}
}

func main() {
	// Setup log.
	log.SetFlags(0)

	// Parse command line.
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "Usage: go-portable [flags] packages")
		fmt.Fprintf(w, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()

	// Call godistlist outside the syntax function, so that we can detect a
	// problem with the go tool early.
	platforms, err := godistlist()
	if err != nil {
		log.Fatal(err)
	}

	if err := run(platforms, args); err != nil {
		log.Fatal(err)
	}
}

// run invokes go vet for all the specified platforms.
func run(platforms []platform, patterns []string) error {
	nl := []byte("\n")
	index := 0 // current failed platform

	for _, sys := range platforms {
		msg, err := govet(sys, patterns)
		if err != nil {
			return err
		}
		if msg == nil {
			continue
		}

		// Print go vet diagnostic message.
		if index > 0 {
			os.Stderr.Write(nl)
		}
		fmt.Fprintf(os.Stderr, "%s/%s using %s\n", sys.os, sys.arch, gocmd)
		os.Stderr.Write(msg)
		os.Stderr.Write(nl)

		index++
	}

	return nil
}

// godistlist invokes go tool dist list to get a list of supported platforms.
func godistlist() ([]platform, error) {
	tool := gocmd + " tool dist list"

	cmd := invoke.Command(gocmdpath, "tool", "dist", "list")
	stdout, err := invoke.Output(cmd)
	if err != nil {
		return nil, err
	}

	// Parse the list of os/arch pairs.
	list := make([]platform, 0, 128) // preallocate memory
	sc := bufio.NewScanner(bytes.NewReader(stdout))
	for sc.Scan() {
		line := sc.Text()
		fields := strings.Split(line, "/")
		if len(fields) != 2 {
			return nil, fmt.Errorf("%s: invalid output: %q", tool, line)
		}

		ent := platform{
			os:   fields[0],
			arch: fields[1],
		}
		list = append(list, ent)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("%s, internal error: %v", tool, err)
	}

	return list, nil
}

// govet invokes go vet on the packages named by the given patterns, for the
// specified platform.  It returns the diagnostic message and a non nil error,
// in case of a fatal error like go command not found or incorrect command line
// arguments.
func govet(sys platform, patterns []string) ([]byte, error) {
	args := append([]string{"vet"}, patterns...)
	cmd := invoke.Command(gocmdpath, args...)
	cmd.Env = append(os.Environ(), "GOOS="+sys.os, "GOARCH="+sys.arch)

	if err := invoke.Run(cmd); err != nil {
		cmderr := err.(*invoke.Error)

		// Determine the error type to decide if there was a fatal problem
		// with the invocation of go vet that requires the termination of
		// the program.
		switch cmderr.Err.(type) {
		case *exec.Error:
			return nil, err
		case *exec.ExitError:
			if isFatal(cmderr) {
				return nil, err
			}

			return cmderr.Stderr, nil
		}
	}

	return nil, nil
}

// isFatal returns true if the error returned by go vet is fatal.
func isFatal(err *invoke.Error) bool {
	// In case of build constraints excluding all Go files, go vet returns
	// exit status 1 and the error message starts with "package".
	//
	// TODO(mperillo): all Go files excluded due to build constraints is
	// probably a fatal error.
	if bytes.HasPrefix(err.Stderr, []byte("package")) {
		return false
	}

	// In case of syntax errors, go vet returns exit status 2 and
	// the error message starts with # and the package name.
	if err.Stderr[0] == '#' {
		return false
	}

	return false
}
