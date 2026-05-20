# Swipe Merchant CLI Makefile
# Run `make help` for a list of targets.

# === Configuration ===

# Update this once and you're done. Used in ldflags for version embedding.
MODULE                := github.com/BML-Digital/swipe-merchants-dev/cli
BIN_NAME              := swipe
BIN_DIR               := ./bin
BIN                   := $(BIN_DIR)/$(BIN_NAME)
PKG                   := ./...

# Pinned versions (bumping these is a deliberate change reviewed in PR)
GOLANGCI_LINT_VERSION := v2.1.6
GORELEASER_VERSION    := v2.5.0

# Build metadata. Stays cross-platform: `git` is the only external
# process we shell out to (it's on PATH on every supported dev host),
# and the spec-version file is read through the local syncspec tool so
# we don't depend on `cat`. macOS ships GNU make 3.81 (no $(file ...)
# builtin) and Windows make may run without an `sh` to interpret
# `cat ... 2>/dev/null || echo`, so the Go helper is the lowest common
# denominator.
VERSION               := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT                := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
# Commit ISO 8601 date instead of `date -u` so Windows builds don't need
# coreutils. Drops back to "unknown" outside a git tree.
BUILD_DATE            := $(shell git log -1 --format=%cI 2>/dev/null || echo "unknown")
SPEC_VERSION          := $(shell go run ./tools/syncspec -cat ../spec/.spec-version 2>/dev/null)
ifeq ($(SPEC_VERSION),)
SPEC_VERSION          := 0.0.0
endif

GOPATH_BIN            := $(shell go env GOPATH)/bin

# Static, CGO-free build is non-negotiable (D-029). Exporting here
# instead of prefixing each recipe with `CGO_ENABLED=0 …` makes the
# build work under make on Windows where recipes are passed directly to
# CreateProcess (no shell to interpret `VAR=value cmd` syntax).
export CGO_ENABLED := 0

LDFLAGS := -s -w \
  -X '$(MODULE)/internal/version.Version=$(VERSION)' \
  -X '$(MODULE)/internal/version.Commit=$(COMMIT)' \
  -X '$(MODULE)/internal/version.BuildDate=$(BUILD_DATE)' \
  -X '$(MODULE)/internal/version.SpecVersion=$(SPEC_VERSION)'

# === Targets ===

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@echo "Swipe Merchant CLI — make targets"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

.PHONY: tidy
tidy: ## go mod tidy
	go mod tidy

.PHONY: sync-spec
sync-spec: ## Copy ../spec/app.yaml into internal/spec/app.yaml for go:embed
	@go run ./tools/syncspec ../spec/app.yaml internal/spec/app.yaml
	@go run ./tools/syncspec ../spec/.spec-version internal/spec/.spec-version

.PHONY: check-spec-sync
check-spec-sync: ## Fail if internal/spec/app.yaml drifts from ../spec/app.yaml
	@go run ./tools/syncspec -check ../spec/app.yaml internal/spec/app.yaml
	@go run ./tools/syncspec -check ../spec/.spec-version internal/spec/.spec-version

.PHONY: build
build: sync-spec ## Build the binary (CGO disabled, static)
	@go run ./tools/syncspec -mkdir $(BIN_DIR)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/swipe
	@echo "built: $(BIN)"
	@echo "  version:      $(VERSION)"
	@echo "  spec version: $(SPEC_VERSION)"
	@echo "  commit:       $(COMMIT)"

.PHONY: run
run: build ## Build and start the mock server
	$(BIN) mock start

.PHONY: install
install: ## Install binary to GOPATH/bin
	go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/swipe
	@echo "installed: $(GOPATH_BIN)/$(BIN_NAME)"

.PHONY: test
test: sync-spec ## Run all tests with race detector
	go test -race -count=1 -timeout 60s $(PKG)

.PHONY: test-cover
test-cover: sync-spec ## Run tests and print coverage summary
	go test -race -count=1 -coverprofile=coverage.out $(PKG)
	@go tool cover -func=coverage.out | tail -20
	@echo ""
	@go tool cover -func=coverage.out | grep total | awk '{print "TOTAL coverage: " $$3}'

.PHONY: cover-html
cover-html: test-cover ## Open coverage report in browser
	go tool cover -html=coverage.out

.PHONY: lint
lint: install-tools sync-spec ## Run golangci-lint
	$(GOPATH_BIN)/golangci-lint run --timeout 5m $(PKG)

.PHONY: lint-fix
lint-fix: install-tools ## Run golangci-lint with --fix
	$(GOPATH_BIN)/golangci-lint run --fix --timeout 5m $(PKG)

.PHONY: install-tools
install-tools: ## Install pinned tool versions (built with the local Go toolchain)
	@if ! [ -x "$(GOPATH_BIN)/golangci-lint" ] || ! $(GOPATH_BIN)/golangci-lint version 2>/dev/null | grep -q "$(GOLANGCI_LINT_VERSION:v%=%)"; then \
	  echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
	  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi

.PHONY: verify
verify: lint test ## Lint + test (run before committing)

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) coverage.out

.PHONY: spec-validate
spec-validate: build ## Validate ../spec/app.yaml is parseable by the embedded loader
	$(BIN) spec validate-self || (echo "spec self-validation failed"; exit 1)

.PHONY: release-snapshot
release-snapshot: ## Goreleaser snapshot (no publish)
	goreleaser release --snapshot --clean

.PHONY: release-check
release-check: ## Validate goreleaser config without building
	goreleaser check

# === CI convenience ===

.PHONY: ci
ci: tidy verify build ## Run the full CI pipeline locally
	@echo ""
	@echo "✓ ci passed"