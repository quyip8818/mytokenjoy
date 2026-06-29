.PHONY: dev test test-unit test-integration build lint format scaffold-domain

dev:
	go run ./cmd/server

test: test-unit

test-unit:
	go test -tags=testhook ./tests/...

test-integration:
	go test -tags=integration -count=1 ./tests/store/postgres/...

build:
	go build -o bin/server ./cmd/server

lint:
	go vet -tags=testhook ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt required on:"; gofmt -l .; exit 1)

format:
	gofmt -w .

scaffold-domain:
	@test -n "$(DOMAIN)" || (echo "usage: make scaffold-domain DOMAIN=foo"; exit 1)
	@chmod +x ./scripts/scaffold-domain.sh
	./scripts/scaffold-domain.sh $(DOMAIN)
