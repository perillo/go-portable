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
	"path/filepath"
	"strings"

	"github.com/perillo/go-portable/internal/invoke"
)

var (
	// gocmd is the go command to use.  It can be overridden using the GOCMD
	// environment variable.
	gocmd = "go"

	// gocmdshort is the base name the go command, used in error messages.
	gocmdshort string
)

// First class ports, taken from
// https://github.com/golang/go/wiki/PortingPolicy#first-class-ports
var firstClass = map[string]bool{
	"linux/amd64":   true,
	"linux/386":     true,
	"linux/arm":     true,
	"linux/arm64":   true,
	"darwin/amd64":  true,
	"windows/amd64": true,
	"windows/386":   true,
}

// Flags.
var primary = flag.Bool("first-class", false, "use only first class ports")

type platform struct {
	os   string
	arch string
}

func init() {
	if value, ok := os.LookupEnv("GOCMD"); ok {
		gocmd = value
	}

	// Don't report the error now.
	if path, err := exec.LookPath(gocmd); err == nil {
		gocmd = path
	}

	gocmdshort = filepath.Base(gocmd)
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
	platforms, err := godistlist(*primary)
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
		fmt.Fprintf(os.Stderr, "%s/%s using %s\n", sys.os, sys.arch, gocmdshort)
		os.Stderr.Write(msg)
		os.Stderr.Write(nl)

		index++
	}

	return nil
}

// godistlist invokes go tool dist list to get a list of supported platforms.
// When primary is true, only first class ports are included.
func godistlist(primary bool) ([]platform, error) {
	tool := gocmdshort + " tool dist list"

	cmd := exec.Command(gocmd, "tool", "dist", "list")
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

		if primary && !firstClass[line] {
			continue
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
// in case of a fatal error like go command not found.
func govet(sys platform, patterns []string) ([]byte, error) {
	args := append([]string{"vet"}, patterns...)
	cmd := exec.Command(gocmd, args...)
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
			return cmderr.Stderr, nil
		}

		return nil, err // should not be reached
	}

	return nil, nil
}
