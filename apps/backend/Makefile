.PHONY: install-air start build run test test-unit test-unit-nocache lint format scaffold-domain clean

export DATABASE_URL ?= postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable
export COMPANY_NAME ?= Demo Company
export SESSION_SECRET ?= tokenjoy-dev-session-secret

AIR_VERSION ?= v1.65.3
AIR := $(shell go env GOPATH)/bin/air
SERVER_BIN := tmp/server

install-air:
	go install github.com/air-verse/air@$(AIR_VERSION)

start:
	@test -x "$(AIR)" || (echo "air not found; run: pnpm install"; exit 1)
	@set -a && [ -f .env ] && . ./.env; set +a; "$(AIR)"

build:
	@mkdir -p tmp
	go build -o $(SERVER_BIN) ./cmd/server

run: build
	exec ./$(SERVER_BIN)

test: test-unit

test-unit:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL required; run: pnpm start:postgres"; exit 1)
	go test -tags=testhook -parallel 4 $$(go list -tags=testhook ./tests/... | grep -v '/tests/testutil')

test-unit-nocache:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL required; run: pnpm start:postgres"; exit 1)
	go test -tags=testhook -parallel 4 -count=1 $$(go list -tags=testhook ./tests/... | grep -v '/tests/testutil')

clean:
	rm -rf tmp

lint:
	go vet -tags=testhook ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt required on:"; gofmt -l .; exit 1)

format:
	gofmt -w .

scaffold-domain:
	@test -n "$(DOMAIN)" || (echo "usage: make scaffold-domain DOMAIN=foo"; exit 1)
	@chmod +x ./scripts/scaffold-domain.sh
	./scripts/scaffold-domain.sh $(DOMAIN)
