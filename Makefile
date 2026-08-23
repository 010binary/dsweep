BINARY_NAME := dsweep
MAIN_PACKAGE := ./cmd/dsweep
BUILDINFO := github.com/010binary/dsweep/internal/buildinfo

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
  -X $(BUILDINFO).Version=$(VERSION) \
  -X $(BUILDINFO).Commit=$(COMMIT) \
  -X $(BUILDINFO).Date=$(DATE)

PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

.DEFAULT_GOAL := help
.PHONY: help run build build-all install test cover fmt vet tidy lint vuln snapshot check ci clean

help: ## Show this help
	@awk 'BEGIN{FS=":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Build and run from source
	go run $(MAIN_PACKAGE)

build: ## Build the binary into bin/
	go build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(BINARY_NAME) $(MAIN_PACKAGE)

build-all: ## Cross-compile for every supported platform into dist/
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		ext=''; if [ "$$os" = windows ]; then ext='.exe'; fi; \
		echo "  building $$os/$$arch"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
			go build -trimpath -ldflags '$(LDFLAGS)' \
			-o dist/$(BINARY_NAME)_$${os}_$${arch}$$ext $(MAIN_PACKAGE) || exit 1; \
	done

install: ## Install the binary into GOBIN
	go install -trimpath -ldflags '$(LDFLAGS)' $(MAIN_PACKAGE)

test: ## Run tests with the race detector
	go test -race ./...

cover: ## Run tests and open an HTML coverage report
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "wrote coverage.html"

fmt: ## Format all Go source
	go fmt ./...

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy and verify module dependencies
	go mod tidy
	go mod verify

lint: ## Run golangci-lint (requires golangci-lint on PATH)
	golangci-lint run

vuln: ## Scan dependencies for known vulnerabilities
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

snapshot: ## Build an unpublished local release into dist/ (requires goreleaser)
	goreleaser release --snapshot --clean

check: fmt vet test ## Format, vet, and test — run this before pushing

ci: ## Run everything CI runs, locally
	go mod tidy -diff
	go mod verify
	golangci-lint run
	go test -race -shuffle=on ./...
	$(MAKE) build-all
	$(MAKE) vuln

clean: ## Remove build and coverage artifacts
	rm -rf bin/ dist/ coverage.out coverage.html
