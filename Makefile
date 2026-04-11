.PHONY: install-deps lint

install-deps:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

lint:
	@golangci-lint run ./...
