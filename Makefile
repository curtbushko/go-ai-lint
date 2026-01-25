OS ?= $(shell uname | tr '[:upper:]' '[:lower:]')
ARCH ?= $(shell uname -m | tr '[:upper:]' '[:lower:]')
DATELOG := "[$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')]"
BINARY := go-ai-lint

ifeq ($(ARCH),x86_64)
	ARCH=amd64
endif

.PHONY: default
default: help

.PHONY: all
all: lint arch-check test build ## Run all quality checks and build

.PHONY: build
build: ## Build the CLI binary
	@mkdir -p $(CURDIR)/bin/$(OS)-$(ARCH)
	@echo "$(DATELOG) Building binary"
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(CURDIR)/bin/$(OS)-$(ARCH)/$(BINARY) ./cmd/go-ai-lint
	@chmod +x $(CURDIR)/bin/$(OS)-$(ARCH)/$(BINARY)

.PHONY: run
run: build ## Run the linter on ./...
	$(CURDIR)/bin/$(OS)-$(ARCH)/$(BINARY) ./...

.PHONY: clean
clean: ## Clean /bin directory and coverage files
	@rm -rf $(CURDIR)/bin
	@rm -f coverage.out coverage.html

.PHONY: install
install: ## Install the CLI using go install
	@echo "$(DATELOG) Installing $(BINARY)"
	GOOS=$(OS) GOARCH=$(ARCH) go install ./cmd/go-ai-lint

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(DATELOG) Linting plugin"
	golangci-lint run -v -c $(CURDIR)/.golangci.yml

.PHONY: test
test: ## Run go tests with race detector
	@echo "$(DATELOG) Running tests"
	go test -race ./...

.PHONY: arch-check
arch-check: ## Run architecture checks
	@echo "$(DATELOG) Running architecture checks"
	go-arch-lint check

.PHONY: coverage
coverage: ## Run tests with coverage report
	@echo "$(DATELOG) Running tests with coverage"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: tidy
tidy: ## Run go mod tidy
	@echo "$(DATELOG) Running go mod tidy"
	go mod tidy

.PHONY: vet
vet: ## Run go vet
	@echo "$(DATELOG) Running go vet"
	go vet ./...

.PHONY: check
check: lint test ## Quick check (lint + test)

.PHONY: help
help: ## Show this help
	@echo "Specify a command. The choices are:"
	@grep -hE '^[0-9a-zA-Z_-]+:.*?## .*$$' ${MAKEFILE_LIST} | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[0;36m%-20s\033[m %s\n", $$1, $$2}'
	@echo ""
