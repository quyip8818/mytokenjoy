.PHONY: dev test build lint format

dev:
	go run ./cmd/server

test:
	go test ./...

build:
	go build -o bin/server ./cmd/server

lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt required on:"; gofmt -l .; exit 1)

format:
	gofmt -w .
