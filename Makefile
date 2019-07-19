.PHONY: test citest

SHELL := /bin/bash -o pipefail
SRC = $(shell find . -name "*.go")

test:
	GO111MODULE=on go test --race -v ./...

citest: .state/coverage.out

.state/coverage.out: $(SRC)
	@mkdir -p .state/
	GO111MODULE=on go test -coverprofile=.state/coverage.out --race -v ./...
