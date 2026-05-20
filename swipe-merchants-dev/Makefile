# Swipe merchants dev — workspace-level Makefile.
# Delegates to per-module Makefiles in cli/ and example/.
# Run `make help` for a list of targets.

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@echo "Swipe merchants dev — workspace targets"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Per-module help: 'make -C cli help'"

.PHONY: build
build: ## Build the cli binary (cli/bin/swipe)
	$(MAKE) -C cli build

.PHONY: test
test: ## Run tests across all modules
	$(MAKE) -C cli test
	cd example && go test -race -count=1 -timeout 60s ./...

.PHONY: lint
lint: ## Lint all modules
	$(MAKE) -C cli lint

.PHONY: verify
verify: lint test ## Lint + test across all modules

.PHONY: tidy
tidy: ## go work sync + tidy each module
	go work sync
	$(MAKE) -C cli tidy
	cd example && go mod tidy

.PHONY: clean
clean: ## Remove build artifacts and stop any running mock
	$(MAKE) -C cli clean
	@pkill -f "swipe mock start" 2>/dev/null || true

.PHONY: sample
sample: build ## Run the example integration end-to-end against the mock
	@command -v jq >/dev/null || { echo "jq is required for the sample target"; exit 1; }
	@./cli/bin/swipe mock stop >/dev/null 2>&1 || true
	@echo "→ starting mock in background (payment-ttl=2s for the Phase 4 demo)"
	@./cli/bin/swipe mock start --payment-ttl 2s >/tmp/swipe-sample-mock.log 2>&1 &
	@sleep 1
	@echo "→ creating sample OAuth client (wallet:balance, transactions:history, transactions:status, payments:qr)"
	@./cli/bin/swipe keys create --name "swipe-sample" --scopes wallet:balance,transactions:history,transactions:status,payments:qr --output json >/tmp/swipe-sample-client.json
	@id=$$(jq -r .id /tmp/swipe-sample-client.json); \
	  secret=$$(jq -r .client_secret /tmp/swipe-sample-client.json); \
	  echo "→ running example with client $$id"; \
	  cd example && SWIPE_CLIENT_ID=$$id SWIPE_CLIENT_SECRET=$$secret go run . ; \
	  status=$$?; \
	  echo "→ cleaning up sample client"; \
	  ../cli/bin/swipe keys delete $$id >/dev/null 2>&1 || true; \
	  echo "→ stopping mock"; \
	  ../cli/bin/swipe mock stop >/dev/null 2>&1 || true; \
	  exit $$status

.PHONY: sample-scenarios
sample-scenarios: build ## Phase 6 demo: enable random_5xx, observe failures + log annotations
	@command -v jq >/dev/null || { echo "jq is required for sample-scenarios"; exit 1; }
	@./cli/bin/swipe mock stop >/dev/null 2>&1 || true
	@./cli/bin/swipe mock start >/tmp/swipe-sample-mock.log 2>&1 &
	@sleep 1
	@echo "→ enabling random_5xx --rate=0.7"
	@./cli/bin/swipe mock scenarios enable random_5xx --args rate=0.7 --output json >/dev/null
	@./cli/bin/swipe keys create --name "scenarios-sample" --scopes wallet:balance --output json >/tmp/swipe-sample-client.json
	@id=$$(jq -r .id /tmp/swipe-sample-client.json); \
	  secret=$$(jq -r .client_secret /tmp/swipe-sample-client.json); \
	  ./cli/bin/swipe auth login --client-id $$id --client-secret $$secret --output json >/dev/null; \
	  echo "→ running 10 wallet balance calls (expect ~7 failures)"; \
	  ok=0; fail=0; \
	  for i in $$(seq 1 10); do \
	    if ./cli/bin/swipe wallet balance --output json >/dev/null 2>&1; then ok=$$((ok+1)); else fail=$$((fail+1)); fi; \
	  done; \
	  echo "→ results: $$ok ok, $$fail failed"; \
	  echo ""; \
	  echo "=== last 5 request log entries (scenarios field) ==="; \
	  ./cli/bin/swipe logs tail --n 5 --output json | jq -c '.[] | {path, status, scenarios}'; \
	  ./cli/bin/swipe mock scenarios disable random_5xx >/dev/null 2>&1 || true; \
	  ./cli/bin/swipe keys delete $$id >/dev/null 2>&1 || true; \
	  ./cli/bin/swipe mock stop >/dev/null 2>&1 || true

.PHONY: sample-webhooks
sample-webhooks: build ## Phase 5 demo: mock + `webhooks listen` + create payment
	@command -v jq >/dev/null || { echo "jq is required for sample-webhooks"; exit 1; }
	@./cli/bin/swipe mock stop >/dev/null 2>&1 || true
	@pkill -f "swipe webhooks listen" >/dev/null 2>&1 || true
	@echo "→ starting mock with --webhook-url http://127.0.0.1:9001 (payment-ttl=2s)"
	@./cli/bin/swipe mock start --webhook-url http://127.0.0.1:9001 --payment-ttl 2s >/tmp/swipe-sample-mock.log 2>&1 &
	@sleep 1
	@echo "→ starting webhooks listener on :9001 (reads secret from mock)"
	@./cli/bin/swipe webhooks listen --listen :9001 >/tmp/swipe-sample-listen.log 2>&1 &
	@sleep 1
	@echo "→ creating sample client + QR payment"
	@./cli/bin/swipe keys create --name "wh-sample" --scopes payments:qr,transactions:status --output json >/tmp/swipe-sample-client.json
	@id=$$(jq -r .id /tmp/swipe-sample-client.json); \
	  secret=$$(jq -r .client_secret /tmp/swipe-sample-client.json); \
	  ./cli/bin/swipe auth login --client-id $$id --client-secret $$secret --output json >/dev/null; \
	  ./cli/bin/swipe payments create --amount 7.50 --currency MVR --type QR --output json >/dev/null; \
	  echo "→ waiting ~5s for TTL transition + webhook delivery"; \
	  sleep 5; \
	  echo ""; \
	  echo "=== webhooks listen output ==="; \
	  cat /tmp/swipe-sample-listen.log; \
	  ./cli/bin/swipe keys delete $$id >/dev/null 2>&1 || true; \
	  pkill -f "swipe webhooks listen" >/dev/null 2>&1 || true; \
	  ./cli/bin/swipe mock stop >/dev/null 2>&1 || true
