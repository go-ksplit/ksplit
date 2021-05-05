.PHONY: test citest build

SHELL := /bin/bash -o pipefail
SRC = $(shell find . -name "*.go")

VERSION ?=`git describe --tags`
DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
VERSION_PACKAGE = github.com/go-ksplit/ksplit/pkg/version
GIT_TREE = $(shell git rev-parse --is-inside-work-tree 2>/dev/null)
ifneq "$(GIT_TREE)" ""
define GIT_UPDATE_INDEX_CMD
git update-index --assume-unchanged
endef
define GIT_SHA
`git rev-parse HEAD`
endef
else
define GIT_UPDATE_INDEX_CMD
echo "Not a git repo, skipping git update-index"
endef
define GIT_SHA
""
endef
endif

define LDFLAGS
-ldflags "\
	-X ${VERSION_PACKAGE}.version=${VERSION} \
	-X ${VERSION_PACKAGE}.gitSHA=${GIT_SHA} \
	-X ${VERSION_PACKAGE}.buildTime=${DATE} \
"
endef

test:
	GO111MODULE=on go test --race -v ./...

citest: .state/coverage.out

.state/coverage.out: $(SRC) go.mod go.sum
	@mkdir -p .state/
	GO111MODULE=on go test -coverprofile=.state/coverage.out --race -v ./...

build: bin/ksplit

bin/ksplit: $(SRC) go.mod go.sum
	@mkdir -p bin/
	go build \
		-o bin/ksplit \
		${LDFLAGS} \
	    ./ksplit
