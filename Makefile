VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/drdanmaggs/rocket-fuel/cmd.Version=$(VERSION)"

.PHONY: build install test lint fmt fmt-check clean all

build:
	go build $(LDFLAGS) -o bin/rocket-fuel .

install:
	go install $(LDFLAGS) .

test:
	go test -race ./...

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .
	goimports -w .

fmt-check:
	@test -z "$$(gofumpt -l .)" || (echo "Files need formatting:" && gofumpt -l . && exit 1)

clean:
	rm -rf bin/

all: fmt lint test build
