.PHONY: test citest build

SHELL := /bin/bash -o pipefail
SRC = $(shell find . -name "*.go")

test:
	GO111MODULE=on go test --race -v ./...

citest: .state/coverage.out

.state/coverage.out: $(SRC) go.mod go.sum
	@mkdir -p .state/
	GO111MODULE=on go test -coverprofile=.state/coverage.out --race -v ./...

build: bin/ksplit

bin/ksplit: $(SRC) go.mod go.sum
	@mkdir -p bin/
	go build -o bin/ksplit ./ksplit
