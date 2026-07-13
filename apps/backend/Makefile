.PHONY: install-air start build run test test-unit test-unit-nocache test-fast lint lint-clock format scaffold-domain clean dev-bootstrap

export DATABASE_URL ?= postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable
export COMPANY_NAME ?= Demo Company
export SESSION_SECRET ?= tokenjoy-dev-session-secret
TEST_PARALLEL ?= 8
TEST_PKG_PARALLEL ?= 2

AIR_VERSION ?= v1.65.3
AIR := $(shell go env GOPATH)/bin/air
SERVER_BIN := tmp/server

install-air:
	go install github.com/air-verse/air@$(AIR_VERSION)

start:
	@test -x "$(AIR)" || (echo "air not found; run: pnpm install"; exit 1)
	@set -a \
		&& [ -f .env.development ] && . ./.env.development \
		&& [ -f .env ] && . ./.env \
		; set +a; "$(AIR)"

build:
	@mkdir -p tmp
	go build -o $(SERVER_BIN) ./cmd/server

run: build
	exec ./$(SERVER_BIN)

dev-bootstrap:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL required; run: pnpm start:postgres"; exit 1)
	@set -a \
		&& [ -f .env.development ] && . ./.env.development \
		&& [ -f .env ] && . ./.env \
		; set +a; go run ./cmd/dev-bootstrap

test: test-unit

test-unit:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL required; run: pnpm start:postgres"; exit 1)
	go test -tags=testhook -p $(TEST_PKG_PARALLEL) -parallel $(TEST_PARALLEL) $$(go list -tags=testhook ./tests/... | grep -v '/tests/testutil')

test-unit-nocache:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL required; run: pnpm start:postgres"; exit 1)
	go test -tags=testhook -p $(TEST_PKG_PARALLEL) -parallel $(TEST_PARALLEL) -count=1 $$(go list -tags=testhook ./tests/... | grep -v '/tests/testutil')

test-fast:
	go test -tags=testhook ./tests/pkg/...

test-db-clean:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL required; run: pnpm start:postgres"; exit 1)
	@go run -tags=testhook ./cmd/testdbclean

clean:
	rm -rf tmp

lint: lint-clock
	go vet -tags=testhook ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt required on:"; gofmt -l .; exit 1)

lint-clock:
	@chmod +x ./scripts/lint-clock.sh
	./scripts/lint-clock.sh

format:
	gofmt -w .

scaffold-domain:
	@test -n "$(DOMAIN)" || (echo "usage: make scaffold-domain DOMAIN=foo"; exit 1)
	@chmod +x ./scripts/scaffold-domain.sh
	./scripts/scaffold-domain.sh $(DOMAIN)
