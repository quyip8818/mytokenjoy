.PHONY: dev test test-unit test-integration build lint format

dev:
	go run ./cmd/server

test: test-unit

test-unit:
	go test ./tests/...

test-integration:
	go test -tags=integration -count=1 ./tests/store/postgres/...

build:
	go build -o bin/server ./cmd/server

lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt required on:"; gofmt -l .; exit 1)

format:
	gofmt -w .
