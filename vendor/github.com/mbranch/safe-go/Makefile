# Ensure that we use vendored binaries before consulting the system.
GOBIN=$(shell pwd)/bin
export PATH := $(GOBIN):$(PATH)

# Use Go modules.
export GO111MODULE := on

golangci-lint=$(GOBIN)/golangci-lint
$(golangci-lint):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.51.2

all: install lint test

.PHONY: install
install: ## Install the library.
	@go install ./...

.PHONY: lint
lint: $(golangci-lint) ## Lint the project with golangci-lint.
	$(golangci-lint) run --config ./.golangci.yml ./...

.PHONY: setup
setup:  ## Download dependencies.
	@GOBIN=$(GOBIN) go mod download

.PHONY: test
test:  ## Run tests.
	@go test ./...

.PHONY: help
help:
	@grep -E '^[/a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
