VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
SOURCE_DIR := $(shell pwd)
LDFLAGS := -ldflags "-X github.com/drdanmaggs/rocket-fuel/cmd.Version=$(VERSION) -X github.com/drdanmaggs/rocket-fuel/cmd.SourceDir=$(SOURCE_DIR)"

.PHONY: build install test test-unit test-integration lint fmt fmt-check clean all setup

build:
	go build $(LDFLAGS) -o bin/rf .

install:
	go build $(LDFLAGS) -o $(HOME)/go/bin/rf .
	@echo "Installed rf to $(HOME)/go/bin/rf"

test:
	go test -race ./...

test-unit:
	go test -race -count=1 ./...

test-integration:
	go test -race -tags=integration ./...

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .
	goimports -w .

fmt-check:
	@test -z "$$(gofumpt -l .)" || (echo "Files need formatting:" && gofumpt -l . && exit 1)

clean:
	rm -rf bin/

setup:
	git config core.hooksPath .githooks
	@echo "Git hooks configured to use .githooks/"

all: fmt lint test build
