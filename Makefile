GOLANGCI_LINT_VERSION    ?= v2.1.5

.PHONY: lint
lint: ## Run lint against code.
	docker run --rm -w /workdir -v $(PWD):/workdir golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run -c .golangci.yml --fix

.PHONY: build
build: #lint
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/capi-mcp main.go

