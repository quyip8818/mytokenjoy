.PHONY: install-air start build run test test-unit test-unit-nocache test-fast test-changed test-integration lint lint-clock format scaffold-domain clean dev-bootstrap

export DATABASE_URL ?= postgres://tokenjoy:tokenjoy@127.0.0.1:5432/tokenjoy?sslmode=disable
export COMPANY_NAME ?= Demo Company
export SESSION_SECRET ?= tokenjoy-dev-session-secret
NCPU := $(shell sysctl -n hw.ncpu 2>/dev/null || nproc 2>/dev/null || echo 8)
TEST_PARALLEL ?= $(shell echo $$(( $(NCPU) * 2 )))
TEST_PKG_PARALLEL ?= $(NCPU)

AIR_VERSION ?= v1.65.3
AIR := $(shell go env GOPATH)/bin/air
SERVER_BIN := tmp/server
SKIP_VERIFY_CODE ?= $(if $(filter true,$(npm_config_skip_verify)),true,false)

install-air:
	go install github.com/air-verse/air@$(AIR_VERSION)

start:
	@test -x "$(AIR)" || (echo "air not found; run: pnpm install"; exit 1)
	@set -a \
		&& [ -f .env.development ] && . ./.env.development \
		&& [ -f .env ] && . ./.env \
		; set +a; \
		if [ "$(SKIP_VERIFY_CODE)" = "true" ]; then export SKIP_VERIFY_CODE=true; fi; \
		"$(AIR)"

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

test-integration:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL required; run: pnpm start:postgres"; exit 1)
	go test -tags="testhook,integration" -race -timeout 180s -p 1 ./tests/integration/...

test-fast:
	go test -tags=testhook ./tests/pkg/...

test-changed:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL required; run: pnpm start:postgres"; exit 1)
	@pkgs=$$(git diff --name-only HEAD 2>/dev/null | grep '^apps/backend/' | \
		sed 's|^apps/backend/||' | \
		sed -n 's|^internal/\(.*\)/[^/]*$$|\1|p' | sort -u | \
		while read -r dir; do \
			for prefix in tests/domain tests/handler tests/infra tests/store tests/http tests/identity tests/seed tests/app tests/config; do \
				match=$$(go list -tags=testhook "./$$prefix/$$dir/..." 2>/dev/null); \
				[ -n "$$match" ] && echo "$$match"; \
			done; \
			match=$$(go list -tags=testhook "./tests/$$dir/..." 2>/dev/null); \
			[ -n "$$match" ] && echo "$$match"; \
		done | sort -u); \
	if [ -z "$$pkgs" ]; then \
		echo "no changed packages detected; running test-fast"; \
		go test -tags=testhook ./tests/pkg/...; \
	else \
		echo "testing changed: $$pkgs"; \
		go test -tags=testhook -p $(TEST_PKG_PARALLEL) -parallel $(TEST_PARALLEL) $$pkgs ./tests/pkg/...; \
	fi

test-db-clean:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL required; run: pnpm start:postgres"; exit 1)
	@go run -tags=testhook ./cmd/testdbclean

clean:
	rm -rf tmp

lint: lint-clock lint-layer
	go vet -tags=testhook ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt required on:"; gofmt -l .; exit 1)

lint-layer:
	@chmod +x ./scripts/layer-guard.sh
	./scripts/layer-guard.sh

lint-clock:
	@chmod +x ./scripts/lint-clock.sh
	./scripts/lint-clock.sh

format:
	gofmt -w .

scaffold-domain:
	@test -n "$(DOMAIN)" || (echo "usage: make scaffold-domain DOMAIN=foo"; exit 1)
	@chmod +x ./scripts/scaffold-domain.sh
	./scripts/scaffold-domain.sh $(DOMAIN)
