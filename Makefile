# QueryX Makefile
#
# Local targets use your host Go toolchain.
# `docker-*` targets run the same workflows inside containers via docker compose,
# so contributors don't need Go installed locally.

BINARY      := queryx
CMD         := ./cmd/queryx
COMPOSE     := docker compose

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ---------------------------------------------------------------------------
# Local (host Go toolchain)
# ---------------------------------------------------------------------------

.PHONY: build
build: ## Build the CLI binary
	go build -trimpath -o $(BINARY) $(CMD)

.PHONY: run
run: ## Run the CLI (pass args via ARGS="-type fivem -host ...")
	go run $(CMD) $(ARGS)

.PHONY: test
test: ## Run all tests (unit + integration)
	go test ./... -v

.PHONY: test-short
test-short: ## Run unit tests only (fast)
	go test ./... -short

.PHONY: test-integration
test-integration: ## Run integration tests only
	go test -v -run TestIntegration

.PHONY: coverage
coverage: ## Generate coverage profile and HTML report
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint installed)
	golangci-lint run ./...

.PHONY: tidy
tidy: ## Tidy go.mod / go.sum
	go mod tidy

.PHONY: clean
clean: ## Remove build and coverage artifacts
	rm -f $(BINARY) $(BINARY).exe coverage.out coverage.html

# ---------------------------------------------------------------------------
# Docker (no local Go toolchain required)
# ---------------------------------------------------------------------------

.PHONY: docker-build
docker-build: ## Build the runtime CLI Docker image (queryx:local)
	$(COMPOSE) build queryx

.PHONY: docker-test
docker-test: ## Run the full test suite in Docker
	$(COMPOSE) run --rm test

.PHONY: docker-test-short
docker-test-short: ## Run unit tests in Docker
	$(COMPOSE) run --rm test-short

.PHONY: docker-lint
docker-lint: ## Run golangci-lint in Docker
	$(COMPOSE) run --rm lint

.PHONY: docker-dev
docker-dev: ## Open an interactive dev shell in Docker
	$(COMPOSE) run --rm dev

.PHONY: docker-run
docker-run: ## Run the CLI in Docker (pass args via ARGS="-type rust -host ...")
	$(COMPOSE) run --rm queryx $(ARGS)

.PHONY: docker-clean
docker-clean: ## Remove Docker containers, images and cache volumes
	$(COMPOSE) down --rmi local --volumes
