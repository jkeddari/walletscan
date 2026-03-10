# 🪙 WalletScan

WalletScan is a lightweight micro-SaaS that lets you look up balances for popular cryptocurrency networks in seconds. It ships a minimal web UI backed by a Go service that talks directly to chain-specific RPCs and CoinGecko for price discovery.

- 🔍 **Instant lookups** — Paste any supported wallet address to fetch balances and USD value.
- 🌐 **Multi-chain aware** — Bitcoin, Ethereum, Solana, and Tron support are live; more chains are on the roadmap.
- ⚡ **Zero friction** — No authentication, cookies, or tracking pixels; just balances.
- 🧩 **Composable UI** — The front end is built with `templ`, HTMX, and Tailwind CSS for fast server-rendered pages.

---

## Prerequisites

| Tool | Why it is needed | Install hint |
| --- | --- | --- |
| Go 1.24+ | Core API/server code | [go.dev/doc/install](https://go.dev/doc/install) |
| `templ` | Renders `.templ` components to Go code | `go install github.com/a-h/templ/cmd/templ@latest` |
| Tailwind CSS CLI | Builds `assets/css/output.css` | `npm install -g tailwindcss` or use `npx tailwindcss` |
| `air` (optional) | Hot reload for the Go server | `go install github.com/cosmtrek/air@latest` |

> All Go tool dependencies listed in `go.mod` can be installed with `go mod download`. Node is only required if you do not already have a Tailwind binary available.

---

## Quick Start

```bash
git clone https://github.com/jkeddari/walletscan.git
cd walletscan
cp .env.example .env    # or create one manually – see below
go mod download
go generate ./...       # runs `templ generate`
```

Add your CoinGecko API key to `.env`:

```dotenv
COINGECKO_APIKEY=xxxxxxxxxxxxxxxxxxxx
PORT=8090                  # optional – defaults to 8090
```

Now launch the development stack:

```bash
task dev
```

`task dev` runs three watchers in parallel:

- `tailwind` – recompiles CSS on file change.
- `templ` – regenerates Go code for `.templ` files.
- `server` – rebuilds and restarts the Go API via `air`.

After everything boots, visit http://localhost:8090.

Want a lighter-weight run? You can skip the watchers and start the server directly:

```bash
COINGECKO_APIKEY=... go run ./cmd/api
```

Assets must exist before starting the server this way. If you need to (re)generate them manually, use:

```bash
task build-css
task build-templ
```

---

## Refreshing Coin Metadata

`coinlist.json` holds CoinGecko IDs and icons used to enrich balances. Regenerate it whenever you want to support new assets:

```bash
COINGECKO_APIKEY=... go run ./cmd/gecko
```

The command fetches up to 500 entries and overwrites `coinlist.json`.

---

## HTTP Interface

The server renders HTML and exposes a small set of routes:

| Route | Method | Description |
| --- | --- | --- |
| `/` | GET | Landing page with wallet search form. |
| `/scan-address` | POST | Validates an address and redirects (via HTMX) to its detail page. |
| `/scan/{address}` | GET | Fetches balances, prices, and renders the results table. |
| `/assets/*` | GET | Serves static assets such as CSS, JS, and icons. |

Use browser dev tools or CURL to inspect the rendered HTML if you plan to embed WalletScan elsewhere.

---

## Project Layout

```
cmd/           # Executables (API server and CoinGecko sync)
internal/      # Chain-specific scanners and utility packages
ui/            # Server-rendered components built with templ
assets/        # Tailwind CSS source and static assets
utils/         # Helper scripts (e.g., formatting)
```

The entry point for the web experience is `cmd/api/main.go`, which wires the HTMX flows and static file serving. Core balance logic sits in `walletscan.Scan` and dispatches to chain-specific scanners under `internal/`.

---

## Contributing & Next Steps

- File an issue or start a discussion if you plan to add new chains or data sources.
- Keep generated files (`*_templ.go`, `assets/css/output.css`) out of commits unless they are required for the change.
- Run `go fmt ./...` before opening a PR. Add tests (or fixtures) alongside new chain integrations to document expected responses.

Roadmap highlights (feel free to help out):

- ERC-20 token balances on EVM chains
- Additional chains (Polygon, Solana tokens, etc.)
- Public API endpoints for programmatic integrations
