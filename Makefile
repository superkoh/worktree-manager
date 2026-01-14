.PHONY: build install clean test run

BINARY_NAME=wt
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/wt

install: build
	cp bin/$(BINARY_NAME) $(HOME)/.local/bin/

clean:
	rm -rf bin/

test:
	go test -v ./...

run: build
	./bin/$(BINARY_NAME)

tidy:
	go mod tidy

.DEFAULT_GOAL := build
