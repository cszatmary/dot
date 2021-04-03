.DEFAULT_GOAL = build
COVERPKGS = ./client,./dotfile,./internal/log

# Get all dependencies
setup:
	@echo Installing dependencies
	@go mod tidy
	@echo Installing tool dependencies
	@shed install
	@shed run go-fish install
.PHONY: setup

build:
	@go build
.PHONY: build

build-snapshot:
	@shed run goreleaser build -- --snapshot --rm-dist
.PHONY: build-snapshot

release:
	@shed run goreleaser release -- --rm-dist
.PHONY: release

# Generate shell completions for distribution
completions:
	@mkdir -p completions
	@shed completions bash > completions/dot.bash
	@shed completions zsh > completions/_dot
.PHONY: completions

# Clean all build artifacts
clean:
	@rm -rf completions
	@rm -rf coverage
	@rm -rf dist
	@rm -f dot
.PHONY: clean

fmt:
	@shed run goimports -- -w .
.PHONY: fmt

check-fmt:
	@./scripts/check_fmt.sh
.PHONY: check-fmt

lint:
	@shed run golangci-lint run ./...
.PHONY: lint

# Remove version installed with go install
go-uninstall:
	@rm $(shell go env GOPATH)/bin/dot
.PHONY: go-uninstall

# Run tests and collect coverage data
test:
	@mkdir -p coverage
	@go test -coverpkg=$(COVERPKGS) -coverprofile=coverage/coverage.txt ./...
	@go tool cover -html=coverage/coverage.txt -o coverage/coverage.html
.PHONY: test

# Run tests and print coverage data to stdout
test-ci:
	@mkdir -p coverage
	@go test -coverpkg=$(COVERPKGS) -coverprofile=coverage/coverage.txt ./...
	@go tool cover -func=coverage/coverage.txt
.PHONY: test-ci
