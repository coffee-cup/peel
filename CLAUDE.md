# peel

Container image inspector — Go backend + React frontend, single binary for prod.

## Important

Never run dev servers (`mise run dev-backend`, `mise run dev-frontend`). The human manages those.

## Validation

Run `mise run check` after making changes to typecheck Go and TypeScript.

ONLY RUN mise commands. Do not run `go vet` or `tsc --noEmit` directly. Use `mise run check` instead.

## Build

`mise run build` — builds frontend, embeds into Go binary, outputs `./peel`

## Project structure

- `cmd/peel/` — CLI entrypoint, flag parsing
- `internal/server/` — HTTP server + API handlers
- `internal/embed/` — `//go:embed` frontend assets
- `internal/image/` — image loading, layer extraction, filesystem tree
- `web/` — React + Vite + Tailwind frontend

## Key conventions

- Go stdlib HTTP server (no framework). `http.ServeMux` with method-pattern routing (`GET /api/...`)
- Frontend proxies `/api` to Go in dev via Vite config. No dev proxy on Go side.
- `internal/embed/dist/` is fully generated, never committed. `generate:embed-stub` mise task creates a placeholder so `go build` works on fresh clones; `dev-backend` and `build` depend on it automatically.
- `go-containerregistry` for all OCI/Docker image operations
- API is JSON over REST. Endpoints defined in `docs/ARCHITECTURE.md`.
