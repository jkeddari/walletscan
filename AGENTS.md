# Repository Guidelines

## Project Structure & Module Organization
`cmd/api` contains the web server entry point and route wiring. `cmd/gecko` refreshes `coinlist.json` from CoinGecko. Core scanning logic lives in `internal/bitcoin`, `internal/evmscan`, `internal/solana`, `internal/tronscan`, and shared types in `internal/types`. Server-rendered UI is under `ui/`: `pages/` for full screens, `modules/` for reusable sections, `components/` for smaller building blocks, and `layouts/` for shared shells. Static files live in `assets/`, with Tailwind input and output in `assets/css/`.

## Build, Test, and Development Commands
Use `go mod download` to install Go dependencies. Run `go generate ./...` after editing `.templ` files or `internal/types/types.go`; this regenerates `*_templ.go` files and the `stringer` output. `task dev` starts the full local loop: Tailwind watch, `templ` watch, and `air` for hot reload. For a lighter run, use `COINGECKO_APIKEY=... go run ./cmd/api` or `task api`. Refresh token metadata with `COINGECKO_APIKEY=... go run ./cmd/gecko` or `task gecko`. Run `go test ./...` or `task test` before opening a PR, even though the suite is currently minimal.

## Coding Style & Naming Conventions
Format Go code with `go fmt ./...`; keep standard Go tabs and import ordering. Package names stay lowercase (`evmscan`, `tronscan`), exported identifiers use `CamelCase`, and private helpers use `camelCase`. Keep HTTP handlers thin in `cmd/api`; put chain-specific logic in `internal/...`. Update `.templ` source files, not generated `*_templ.go` files, and keep UI component names aligned with their directory purpose, for example `ui/modules/scanform.templ`.

## Testing Guidelines
There are no committed `_test.go` files yet, so new features should add tests beside the package they cover. Prefer table-driven Go tests and deterministic fixtures over live RPC/API calls. For scanner changes, validate parsing, address detection, and aggregation behavior locally with `go test ./...`.

## Commit & Pull Request Guidelines
Recent history uses short imperative subjects such as `ui: add token icons`, `support 500 tokens`, and `fix issue`. Follow that style: optional scope, present tense, under 72 characters. PRs should summarize the user-visible change, list verification commands, link the relevant issue, and include screenshots when `ui/` or rendered output changes.

## Security & Configuration Tips
Keep real secrets out of version control; `.env` is loaded locally and must provide `COINGECKO_APIKEY`. Treat `coinlist.json` and generated CSS as intentional artifacts: regenerate them when needed, but avoid unrelated churn in commits.
