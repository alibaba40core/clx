.PHONY: build test lint install clean help bootstrap-local budgets

BIN_DIR := bin
CLX_BIN := $(BIN_DIR)/clx
CLX_AI_BIN := $(BIN_DIR)/clx-ai
CLXMAX_BIN := $(BIN_DIR)/clxmax

# v1.0.0 = Phases 1–4 + 3.5 (clx). v2.0.0-dev = Phase 5 clxmax (in development).
VERSION ?= 1.0.1
CLXMAX_VERSION ?= 2.0.0-dev
COMMIT ?= unknown
LDFLAGS := -s -w -X github.com/alibaba40core/clx/internal/cliversion.Version=$(VERSION) -X github.com/alibaba40core/clx/internal/cliversion.Commit=$(COMMIT)
CLXMAX_LDFLAGS := -s -w -X github.com/alibaba40core/clx/internal/cliversion.Version=$(CLXMAX_VERSION) -X github.com/alibaba40core/clx/internal/cliversion.Commit=$(COMMIT)

GOFLAGS := -trimpath

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

build: ## Build clx (lite), clx-ai (worker), and clxmax binaries
ifeq ($(OS),Windows_NT)
	@if not exist "$(BIN_DIR)" mkdir "$(BIN_DIR)"
else
	@mkdir -p $(BIN_DIR)
endif
	go run ./cmd/genrules
	go build $(GOFLAGS) -tags=lite -ldflags="$(LDFLAGS)" -o $(CLX_BIN) ./cmd/clx
	go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(CLX_AI_BIN) ./cmd/clx-ai
	go build $(GOFLAGS) -ldflags="$(CLXMAX_LDFLAGS)" -o $(CLXMAX_BIN) ./cmd/clxmax

test: ## Run unit and integration tests
	go test -race ./...
	go test -tags=lite -race ./internal/pipeline/ ./internal/clxsidecar/ ./cmd/clx/

lint: ## Run go vet
	go vet ./...

install: build ## Install via dev scripts (OS-specific)
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File scripts/install.ps1
else
	bash scripts/install.sh
endif

clean: ## Remove build artifacts
ifeq ($(OS),Windows_NT)
	@if exist "$(BIN_DIR)" rmdir /s /q "$(BIN_DIR)" 2>nul
	@if exist clx.exe del /f /q clx.exe 2>nul
	@if exist clx-ai.exe del /f /q clx-ai.exe 2>nul
	@if exist clxmax.exe del /f /q clxmax.exe 2>nul
else
	@rm -rf $(BIN_DIR)
	@rm -f clx clx-ai clxmax clx.exe clx-ai.exe clxmax.exe
endif
	@echo "clean: done"

bootstrap-local: build ## Bootstrap ~/.clx using CLX_HOME=./.local-clx
	CLX_HOME=./.local-clx $(CLX_BIN) --version

budgets: build ## Enforce runtime footprint budgets (size, cold start)
	bash scripts/check-budgets.sh
