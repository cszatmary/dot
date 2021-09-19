.DEFAULT_GOAL = build
COVERPKGS = ./client,./dotfile,./internal/log

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

setup: ## Install all dependencies
	@echo Installing dependencies
	@go mod tidy
	@echo Installing tool dependencies
	@shed get
	@shed run go-fish install
.PHONY: setup

build: ## Build shed
	@go build
.PHONY: build

build-snapshot: ## Create a snapshot release build
	@shed run goreleaser build --snapshot --rm-dist
.PHONY: build-snapshot

release: ## Create a new release of dot
	$(if $(version),,$(error version variable is not set))
	git tag -a v$(version) -m "v$(version)"
	git push origin v$(version)
	shed run goreleaser release --rm-dist
.PHONY: release

completions: ## Generate shell completions for distribution
	@mkdir -p completions
	@go run main.go completions bash > completions/dot.bash
	@go run main.go completions zsh > completions/_dot
.PHONY: completions

clean: ## Clean all build artifacts
	@rm -rf completions
	@rm -rf coverage
	@rm -rf dist
	@rm -f dot
.PHONY: clean

fmt: ## Format all go files
	@shed run goimports -w .
.PHONY: fmt

check-fmt: ## Check if any go files need to be formatted
	@./scripts/check_fmt.sh
.PHONY: check-fmt

lint: ## Lint go files
	@shed run golangci-lint run ./...
.PHONY: lint

# Run tests and collect coverage data
test: ## Run all tests
	@mkdir -p coverage
	@go test -coverpkg=$(COVERPKGS) -coverprofile=coverage/coverage.txt ./...
.PHONY: test

cover: test ## Run all tests and generate coverage data
	@go tool cover -html=coverage/coverage.txt -o coverage/coverage.html
.PHONY: cover
