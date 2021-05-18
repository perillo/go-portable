# Go Portability Checker

[![Go Reference](https://pkg.go.dev/badge/github.com/perillo/go-portable.svg)](https://pkg.go.dev/github.com/perillo/go-portable)

## Installation

go-portable requires [Go 1.16](https://golang.org/doc/devel/release.html#go1.16).

    go install github.com/perillo/go-portable@latest

## Purpose

go-portable checks if a package is compatible with other platforms.
Internally, it invokes `go vet` or `go build` on all the officially supported
ports, as reported by `go tool dist list`.

The output of this tool reports problems for each platform that a package does
not support.

## Usage

    go-portable [-first-class] [-mode mode] [packages]

Invoke `go-portable` with one or more import paths.  go-portable uses the
same [import path syntax](https://golang.org/cmd/go/#hdr-Import_path_syntax) as
the `go` command and therefore also supports relative import paths like
`./...`. Additionally the `...` wildcard can be used as suffix on relative and
absolute file paths to recurse into them.

When the `-first-class` option is set, only
[first class ports](https://github.com/golang/go/wiki/PortingPolicy#first-class-ports) are used.

The `-mode` option allows the user to specify how to verify portability.  It
can be set to `vet` or `build`, with `vet` being the default.

By default, `go-portable` uses the `go` command installed on the system, but it
is possible to specify a different version using the `GOCMD` environment
variable.  When using `-mode build`, `GOCMD` should be `go1.8 or later`.
