.PHONY: install-deps lint openapi

install-deps:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/v2/cmd/swag@latest

lint:
	@golangci-lint run ./...

openapi:
	@swag init -g main.go --dir cmd/gateway,internal/analytics/domain,internal/analytics/infra/http,internal/core/http,internal/module/infra/http,internal/module/infra/http/handlers,internal/response/infra/grpc,internal/user/infra/http,internal/webhook/infra/http,internal/webhook/infra/http/handlers --parseInternal --parseDependency --output cmd/gateway/api --outputTypes yaml,json --v3.1
	@powershell -NoProfile -Command "Move-Item -Force cmd/gateway/api/swagger.yaml cmd/gateway/api/openapi.yml; Move-Item -Force cmd/gateway/api/swagger.json cmd/gateway/api/openapi.json"
