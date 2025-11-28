.PHONY: build test clean install run

build:
	go build -o clem ./cmd/clem

test:
	go test -v -race ./...

test-short:
	go test -v -short ./...

clean:
	rm -f clem
	go clean

install:
	go install ./cmd/clem

run:
	go run ./cmd/clem

lint:
	golangci-lint run

.DEFAULT_GOAL := build
