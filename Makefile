.PHONY: build test lint clean run install

BINARY := watchdog
MODULE := github.com/devaloi/watchdog

build:
	go build -o $(BINARY) ./cmd/watchdog

test:
	go test -race ./...

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)
	go clean

run: build
	./$(BINARY)

install:
	go install ./cmd/watchdog

all: lint test build
