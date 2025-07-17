GOLANGCI_LINT_VERSION    ?= v2.1.5

.PHONY: lint
lint: ## Run golangci-lint against code, installing if necessary.
	@if ! command -v golangci-lint >/dev/null 2>&1 && [ ! -f "./bin/golangci-lint" ]; then \
	  echo "golangci-lint not found, installing..."; \
	  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- $(GOLANGCI_LINT_VERSION); \
	fi; \
	if [ -f "./bin/golangci-lint" ]; then \
	  ./bin/golangci-lint run -c .golangci.yml; \
	else \
	  golangci-lint run -c .golangci.yml; \
	fi

.PHONY: lint-fix
lint-fix: ## Run golangci-lint against code with auto-fix enabled, installing if necessary.
	@if ! command -v golangci-lint >/dev/null 2>&1 && [ ! -f "./bin/golangci-lint" ]; then \
	  echo "golangci-lint not found, installing..."; \
	  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- $(GOLANGCI_LINT_VERSION); \
	fi; \
	if [ -f "./bin/golangci-lint" ]; then \
	  ./bin/golangci-lint run -c .golangci.yml --fix; \
	else \
	  golangci-lint run -c .golangci.yml --fix; \
	fi

.PHONY: build
build: #lint
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/capi-mcp main.go

