.PHONY: run test test-verbose build

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test -race ./...

test-verbose:
	go test -race ./... -v
