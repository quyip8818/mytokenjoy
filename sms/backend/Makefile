.PHONY: build run seed reset dev lint format test

build:
	go build -o bin/server ./cmd/server
	go build -o bin/seed ./cmd/seed

run:
	go run ./cmd/server

seed:
	go run ./cmd/seed

reset: seed

dev:
	@which air > /dev/null 2>&1 && air || go run ./cmd/server

lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt required on:"; gofmt -l .; exit 1)

format:
	gofmt -w .

test:
	go test ./tests/...
