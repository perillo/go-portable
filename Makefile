# Copyright 2019 Manlio Perillo. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# A Makefile template for Go projects.

# Exported variable definitions.
export GO111MODULE := on

# Imported variables.
# GOPKG - used to select the target package

# Variable definitions.
BENCHFLAGS := -v
COVERMODE := atomic # atomic is necessary if the -race flag is enabled
TESTFLAGS := -race -v

# Standard rules.
.POSIX:

# Default rule.
.PHONY: build
build:
	go build -o build ./...

# Custom rules.
.PHONY: bench
bench:
	go test ${BENCHFLAGS} -bench=. -benchmem ./...

.PHONY: clean
clean:
	go mod tidy
	go clean
	go clean -i
	rm -f build/*

.PHONY: cover
cover:
	go tool cover -html=build/coverage.out -o=build/coverage.html

.PHONY: github
github:
	git push --follow-tags -u github master

.PHONY: install
install:
	go install ./...

.PHONY: lint
lint:
	golint ./...

.PHONY: print
print:
	goprint -font='"Inconsolata" 10pt/12pt' ${GOPKG} > build/pkg.html
	prince -o build/pkg.pdf build/pkg.html

.PHONY: test
test:
	go test ${TESTFLAGS} -covermode=${COVERMODE} \
		-coverprofile=build/coverage.out ./...

.PHONY: test-trace
test-trace:
	go test ${TESTFLAGS} -trace=build/trace.out ${GOPKG}

.PHONY: trace
trace:
	go tool trace build/trace.out

.PHONY: vet
vet:
	go vet ./...
