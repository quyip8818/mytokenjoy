.PHONY: start test test-unit build lint format scaffold-domain

export DATABASE_URL ?= postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable
export COMPANY_NAME ?= Demo Company
export SESSION_SECRET ?= tokenjoy-dev-session-secret

start:
	@set -a && [ -f .env ] && . ./.env; set +a; go run ./cmd/server

test: test-unit

test-unit:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL required; run: pnpm start:postgres"; exit 1)
	go test -tags=testhook $$(go list -tags=testhook ./tests/... | grep -v '/tests/testutil')

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
