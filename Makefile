GOLANGCI_LINT_VERSION    ?= v2.1.5

.PHONY: lint
lint: ## Run golangci-lint against code, installing if necessary.
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
	  echo "golangci-lint not found, installing..."; \
	  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- $(GOLANGCI_LINT_VERSION); \
	  export PATH="$(PWD)/bin:$${PATH}"; \
	fi; \
	./bin/golangci-lint run -c .golangci.yml --fix || golangci-lint run -c .golangci.yml --fix

.PHONY: build
build: #lint
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/capi-mcp main.go

