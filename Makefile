.PHONY: build test lint install clean help

BIN_DIR := bin
CLX_BIN := $(BIN_DIR)/clx
CLXMAX_BIN := $(BIN_DIR)/clxmax

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build clx and clxmax binaries
	@mkdir -p $(BIN_DIR)
	@echo "build: stub — implement in Phase 1"
	# go build -o $(CLX_BIN) ./cmd/clx
	# go build -o $(CLXMAX_BIN) ./cmd/clxmax

test: ## Run unit and integration tests
	@echo "test: stub — implement in Phase 1"
	# go test ./...

lint: ## Run linters
	@echo "lint: stub — implement in Phase 1"
	# golangci-lint run ./...

install: build ## Install binaries to GOPATH/bin
	@echo "install: stub — implement in Phase 1"
	# go install ./cmd/clx
	# go install ./cmd/clxmax

clean: ## Remove build artifacts
	@rm -rf $(BIN_DIR)
	@echo "clean: done"
