VERSION := $(shell git describe --tags --abbrev=0)
REVISION := $(shell git rev-parse --short HEAD)
SRC_DIR := ./
BIN_NAME := boltdb-exporter
BINARY := bin/$(BIN_NAME)

GOLANGCI_LINT_VERSION := v1.30.0
export GO111MODULE=on

## Build binaries on your environment
build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(SRC_DIR)

## Format source codes
fmt:
	@(if ! type goimports >/dev/null 2>&1; then go get -u golang.org/x/tools/cmd/goimports ;fi)
	goimports -w $$(go list -f {{.Dir}} ./... | grep -v /vendor/)

## Lint
lint:
	@(if ! type golangci-lint >/dev/null 2>&1; then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION) ;fi)
	golangci-lint run ./...

## Run tests for my project
test:
	go test -v ./...

## Clean Binary
clean:
	rm -f $(BIN_NAME)
