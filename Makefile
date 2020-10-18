.PHONY: test lint build acceptance

check: test lint build acceptance

test:
	@go test -v -race ./...

lint: 
	@golangci-lint run

build:
	@docker build -t docker.local/userservice:latest .

acceptance:
	@(cd acceptance && go test -v -count=1 -race ./...)
