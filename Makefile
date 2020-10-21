.PHONY: help test lint build acceptance

help: ## Show this help
	@echo "Help"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[36m%-20s\033[93m %s\n", $$1, $$2}'

check: ## Run all checks
check: test lint build acceptance

test: ## Run unit tests
	@go test -v -race ./...

lint: ## Lint the code
	@golangci-lint run

build: ## Compile the service and build the docker image
	@docker build -t docker.local/userservice:latest .

acceptance: ## Run the acceptance tests for the compiled image
	@(cd acceptance && go test -v -count=1 -race ./...)
