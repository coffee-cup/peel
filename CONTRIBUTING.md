# Contributing

## Prerequisites

Install [Mise](https://mise.jdx.dev/) — it manages Go, Node, and Bun.

## Setup

```sh
git clone https://github.com/coffee-cup/peel.git
cd peel
mise install
cd web && bun install && cd ..
```

## Development

Run in separate terminals:

```sh
mise run dev-backend    # Go server on :9870
mise run dev-frontend   # Vite dev server with HMR
```

The frontend proxies `/api` to the Go backend via Vite config.

## Validation

```sh
mise run check   # typecheck Go + TypeScript
mise run test    # Go + Vitest tests
mise run build   # production binary with embedded frontend
```

## Project Structure

```
cmd/peel/        CLI entrypoint, flag parsing
internal/server/ HTTP server + API handlers
internal/image/  image loading, layer extraction, filesystem tree
internal/embed/  go:embed frontend assets (generated, not committed)
web/             React + Vite + Tailwind frontend
```

## Pull Requests

Branch off `main`. CI runs check, test, and build automatically on PRs.
