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
	"strings"

	"github.com/perillo/go-syntax/internal/invoke"
)

// gocmd is the go command to use.  It can be overridden using the GOCMD
// environment variable.
var gocmd = "go"

type platform struct {
	os   string
	arch string
}

func init() {
	value, ok := os.LookupEnv("GOCMD")
	if ok {
		gocmd = value
	}
}

func main() {
	// Setup log.
	log.SetFlags(0)

	// Parse command line.
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "Usage: go-syntax [flags] packages")
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

	status := syntax(platforms, args)
	os.Exit(status)
}

func syntax(platforms []platform, patterns []string) int {
	nl := []byte("\n")
	index := 0  // current failed platform
	status := 0 // process exit status

	for _, sys := range platforms {
		if err := govet(sys, patterns); err != nil {
			status = 1
			err := err.(*invoke.Error)

			if index > 0 {
				os.Stderr.Write(nl)
			}
			fmt.Fprintf(os.Stderr, "%s/%s using %s\n", sys.os, sys.arch, gocmd)
			os.Stderr.Write(err.Stderr)
			os.Stderr.Write(nl)

			index++
		}
	}

	return status
}

// godistlist invokes go tool dist list to get a list of supported platforms.
func godistlist() ([]platform, error) {
	tool := gocmd + " tool dist list"

	cmd := invoke.Command(gocmd, "tool", "dist", "list")
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
// specified platform.
func govet(sys platform, patterns []string) error {
	args := append([]string{"vet"}, patterns...)
	cmd := invoke.Command(gocmd, args...)
	cmd.Env = append(os.Environ(), "GOOS="+sys.os, "GOARCH="+sys.arch)

	return invoke.Run(cmd)
}
