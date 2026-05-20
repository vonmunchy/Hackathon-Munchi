# Swipe merchant sample integration

A small, runnable end-to-end example that exercises the embedded Swipe mock
server the way a real merchant integration would: HTTP only, no `internal/`
imports.

## Phase coverage

This binary grows with the project. Today it covers:

- **Phase 1** done — `/health/alive`, `/health/ready` probes.
- **Phase 2** done — OAuth `client_credentials` -> `POST /oauth2/token` -> `GET /api/v1/whoami`. Activated when `SWIPE_CLIENT_ID` and `SWIPE_CLIENT_SECRET` are set.
- **Phase 3** done — `GET /api/v1/balance`, `GET /api/v1/bank-accounts`, `GET /api/v1/history` (called after whoami, when Phase 2 envs are set).
- **Phase 4** done — `POST /api/v1/payments` + SSE `GET /api/v1/payments/{id}/stream` (creates a QR payment, watches the transition).
- **Phase 5** done — customer-facing `/pay/{shortCode}` pay page (mock-only). The example creates a fresh QR payment, surfaces the `payment_url`, drives the transition via the public `POST /pay/{code}/complete` route (no auth — a real customer wouldn't have credentials), then reads the resulting transaction back through the merchant API.
- **Webhook delivery** — covered separately by `make sample-webhooks` (this target runs `swipe webhooks listen` in the background, points the mock at it via `--webhook-url`, and creates a payment so the listener prints the verified delivery). See `swipe webhooks --help`.
- **Scenario toggles** — covered separately by `make sample-scenarios` (arms `random_5xx` and observes the failure cluster in the request log).

## Running it

The simplest path is from the repo root:

```
make sample
```

That target builds the CLI, starts the mock, creates a sample OAuth client
via `swipe keys create`, runs this binary with the resulting credentials in
`SWIPE_CLIENT_ID` / `SWIPE_CLIENT_SECRET`, then cleans up. (`jq` is
required.)

## Running it by hand

```
# terminal 1
cli/bin/swipe mock start

# terminal 2 -- Phase 1 only
go run ./example

# terminal 2 -- Phase 2 with auth
cli/bin/swipe keys create --name sample --scopes wallet:balance --output json
# copy the id + client_secret from the output, then:
SWIPE_CLIENT_ID=cli_... SWIPE_CLIENT_SECRET=sec_... go run ./example
```

The mock binds to `127.0.0.1:8080` by default. Set `SWIPE_MOCK_URL` if you
started the mock on a different address.
