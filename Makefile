SHELL := /bin/sh

SERVICES ?= analytics fetcher modules scheduler users
BIN_DIR ?= bin
REPORTS_DIR ?= reports
COVERAGE_DIR ?= coverage
COVERAGE_PROFILE ?= $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML ?= $(COVERAGE_DIR)/index.html
COVERAGE_TEXT ?= $(COVERAGE_DIR)/coverage.txt
COVERAGE_XML ?= $(COVERAGE_DIR)/coverage.xml
COVERAGE_THRESHOLD ?= 30
GO_PACKAGES ?= ./...
GO_BUILD_FLAGS ?= -trimpath
GO_LDFLAGS ?= -s -w

.PHONY: install-deps install-lint-deps install-test-deps lint openapi build test coverage coverage-check service-coverage-check clean
.PHONY: $(SERVICES:%=build-%)

install-deps: install-lint-deps install-test-deps

install-lint-deps:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/v2/cmd/swag@latest

install-test-deps:
	@go install gotest.tools/gotestsum@latest
	@go install github.com/boumenot/gocover-cobertura@latest

lint:
	@golangci-lint run ./...

openapi:
	@swag init -g main.go --dir cmd/gateway,internal/analytics/domain,internal/analytics/infra/http,internal/core/http,internal/module/infra/http,internal/module/infra/http/handlers,internal/response/infra/grpc,internal/user/infra/http,internal/webhook/infra/http,internal/webhook/infra/http/handlers --parseInternal --parseDependency --output api --outputTypes yaml,json --v3.1
	@mv -f api/swagger.yaml api/openapi.yml
	@mv -f api/swagger.json api/openapi.json

build: $(SERVICES:%=build-%)

$(SERVICES:%=build-%): build-%:
	@mkdir -p $(BIN_DIR)
	@go build $(GO_BUILD_FLAGS) -ldflags="$(GO_LDFLAGS)" -o $(BIN_DIR)/$* ./cmd/$*

test:
	@mkdir -p $(REPORTS_DIR) $(COVERAGE_DIR)
	@gotestsum --junitfile $(REPORTS_DIR)/junit.xml --format standard-verbose -- -covermode=count "-coverprofile=$(COVERAGE_PROFILE)" $(GO_PACKAGES)
	@$(MAKE) coverage

coverage:
	@go tool cover "-func=$(COVERAGE_PROFILE)" > $(COVERAGE_TEXT)
	@grep '^total:' $(COVERAGE_TEXT)
	@go tool cover "-html=$(COVERAGE_PROFILE)" -o $(COVERAGE_HTML)
	@gocover-cobertura < $(COVERAGE_PROFILE) > $(COVERAGE_XML)

coverage-check:
	@coverage=$$(go tool cover "-func=$(COVERAGE_PROFILE)" | awk '/^total:/ { sub(/%/,"",$$3); print $$3 }'); \
	echo "Total coverage: $$coverage%"; \
	awk -v coverage="$$coverage" -v threshold="$(COVERAGE_THRESHOLD)" 'BEGIN { if (coverage + 0 < threshold + 0) { printf "Coverage %.1f%% is below required %.1f%%\n", coverage, threshold; exit 1 } }'

service-coverage-check:
	@COVERAGE_DIR="$(COVERAGE_DIR)" COVERAGE_THRESHOLD="$(COVERAGE_THRESHOLD)" SERVICES="$(SERVICES)" sh ./scripts/check-service-coverage.sh

clean:
	@rm -rf $(BIN_DIR) $(REPORTS_DIR) $(COVERAGE_DIR)
