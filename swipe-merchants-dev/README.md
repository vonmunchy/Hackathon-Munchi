# swipe

`swipe` is a single-binary CLI for merchants integrating with the Swipe
payment platform. It ships with an embedded mock of the Swipe Merchants
API so you can build and test your integration end-to-end on your
laptop — no shared sandbox required.

The OpenAPI spec under [`spec/app.yaml`](spec/app.yaml) is the source
of truth; the same spec drives the mock's routing/validation and the
binary's `swipe spec show` / `swipe spec validate` commands.

## Installation

Three options, all producing the same `swipe` binary. All three need
[Go 1.26 or later](https://go.dev/doc/install) — Homebrew installs
Go for you as a build dependency; the other two expect it on your
`PATH`.

### Homebrew (macOS, Linux, WSL)

```sh
brew tap BML-Digital/tap
brew install swipe
```

To upgrade later:

```sh
brew update
brew upgrade swipe
```

### `go install`

After install, ensure `$(go env GOBIN)` — or `$(go env GOPATH)/bin`
when `GOBIN` is unset — is on your `PATH`.

```sh
go install github.com/BML-Digital/swipe-merchants-dev/cli/cmd/swipe@latest
```

### Manual build

Also requires `make`.

```sh
git clone https://github.com/BML-Digital/swipe-merchants-dev.git
cd swipe-merchants-dev
make build
# binary lands at cli/bin/swipe — move or symlink it onto your PATH
```

Verify with `swipe version`.

## Quickstart

Assumes `swipe` is on your `PATH` (see [Installation](#installation)).
For a manual build, run from the repo root with `./cli/bin/swipe`
instead.

```sh
swipe mock start &                  # boots the embedded mock at :8080
swipe keys create --name "my-test" --scopes wallet:balance,payments:qr
# copy the printed client_id + client_secret into the next command
swipe auth login --client-id cli_... --client-secret sec_...
swipe auth whoami
swipe wallet balance
swipe payments create --amount 25.00 --currency MVR --type QR
swipe mock stop
```

A more polished run: `make sample` (Phase 1-4 demo) or
`make sample-webhooks` / `make sample-scenarios` for the later phases.

## Command index

```
swipe
├── auth                # OAuth2 client credentials
│   ├── login           # POST /oauth2/token, cache at ~/.swipe/token.json
│   ├── token           # print cached access token (masked by default)
│   ├── whoami          # GET /api/v1/whoami
│   └── logout          # remove cached token
├── keys                # OAuth client management (mock state)
│   ├── list / show <id>
│   ├── create --name --scopes [--merchant]
│   ├── update <id> [--name --scopes --enabled]
│   ├── rotate <id> | revoke <id> | delete <id>
│   └── test <id> --secret      # exchange + whoami in one call
├── wallet
│   ├── balance         # GET /api/v1/balance
│   └── accounts        # GET /api/v1/bank-accounts
├── payments
│   ├── create [--amount --currency --type --description --recipient-vpa] | -f FILE
│   ├── get <id>        # GET /api/v1/payments/{id}
│   └── watch <id>      # SSE /api/v1/payments/{id}/stream
├── payouts
│   └── create [--amount --bank-account-id] | -f FILE
├── transactions
│   └── history [--limit --offset]
├── webhooks
│   ├── listen [--listen --forward-to --secret --skip-verify --print-body]
│   ├── verify -f FILE --id --timestamp --signature --secret
│   └── secret          # print mock secret via /_admin/status
├── mock
│   ├── start [--port --webhook-url --payment-ttl]
│   ├── stop | status
│   └── scenarios
│       ├── list | show <name>
│       └── enable <name> [--args k=v,...] | disable <name>
├── logs
│   ├── tail [--n]
│   └── show <id>
├── spec
│   ├── show [--format yaml|json]
│   ├── version
│   └── validate --operation OP_ID -f FILE
├── health
│   ├── alive
│   └── ready
├── config path
├── completion <bash|zsh|fish|powershell>
└── version
```

Global flags: `--output json|table|yaml` (default `table`), `--quiet`,
`--verbose` (shows the full error chain on failure), `--no-color`,
`--port <n>` (override mock port for this invocation), `--config <path>`.

## Errors

Errors render with the API's RFC 9457 `type` + `detail`, plus a short
hint when one applies. Example:

```
$ swipe wallet balance
Error: FORBIDDEN
  Token missing required scope: wallet:balance
  hint: update the client's scopes via `swipe keys update --scopes ...`
```

Use `--verbose` for the full unwrap chain.

## Scenarios

Ten built-in scenarios cover the unhappy paths a merchant integration
should handle: `latency_injection`, `random_5xx`, `rate_limit_burst`,
`token_short_ttl`, `client_revoked_after`, `insufficient_funds`,
`payment_stuck_pending`, `payment_transitions`, `scope_downgrade`,
`webhook_delivery_fail`.

```sh
swipe mock scenarios enable random_5xx --args rate=0.5
swipe wallet balance     # may now fail with a 5xx
swipe logs tail --n 5    # the failed entries carry "scenarios":["random_5xx:503"]
swipe mock scenarios disable random_5xx
```

## Webhooks

```sh
# terminal 1
swipe mock start --webhook-url http://127.0.0.1:9000 --payment-ttl 2s

# terminal 2
swipe webhooks listen --listen :9000

# terminal 3
swipe keys create --name "wh" --scopes payments:qr
swipe auth login --client-id ... --client-secret ...
swipe payments create --amount 5 --currency MVR --type QR
# the listener prints "signature=verified" for the PENDING + COMPLETED webhook
```

Offline signature verification:

```sh
swipe webhooks verify -f payload.json \
  --id msg_... --timestamp 1700... --signature "v1,..." --secret whsec_...
```

## Files

`swipe` keeps every file under `~/.swipe/`:

```
~/.swipe/
├── config.yaml            # CLI config (D-025)
├── token.json             # cached access token
└── mock/
    ├── state.db           # BoltDB state
    ├── keys/              # RSA signing keypair
    └── mock.pid / mock.addr
```

`swipe config path` prints the exact resolved paths.

## Development

This repo is a Go workspace.

- `cli/` — the swipe binary (one module).
- `example/` — a runnable merchant-style integration demonstrating each phase.
- `spec/app.yaml` — canonical OpenAPI document, embedded via `//go:embed`.
- `docs/.scratched/` — gitignored working docs (CLAUDE.md, scaffold, decisions, TODO tracking). Read these before contributing.

Top-level make targets:

| Target               | Description                                                          |
|----------------------|----------------------------------------------------------------------|
| `make build`         | Build `cli/bin/swipe` (CGO disabled, static).                        |
| `make test`          | Run tests on both modules with the race detector.                    |
| `make lint`          | golangci-lint on `cli/`.                                             |
| `make verify`        | `lint` + `test`.                                                     |
| `make sample`        | Phase 1-4 end-to-end demo against the embedded mock.                 |
| `make sample-webhooks` | Phase 5 demo with `swipe webhooks listen`.                         |
| `make sample-scenarios` | Phase 6 demo: enable `random_5xx` + observe annotations.          |
| `make clean`         | Remove build artifacts and stop any running mock.                    |
| `make tidy`          | `go work sync` + `go mod tidy` in each module.                       |

Releases use [goreleaser](.goreleaser.yml). Snapshot:

```sh
make -C cli release-snapshot
```

## Layout

```
swipe-merchants-dev/
├── go.work                 # use ./cli, ./example
├── Makefile                # top-level delegator
├── README.md               # this file
├── .goreleaser.yml
├── spec/
│   ├── app.yaml            # OpenAPI 3.1 source of truth
│   └── .spec-version
├── cli/
│   ├── go.mod
│   ├── Makefile
│   ├── cmd/swipe/main.go
│   └── internal/...
└── example/
    ├── go.mod
    └── main.go
```
