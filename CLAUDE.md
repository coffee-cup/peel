# peel

Container image inspector — Go backend + React frontend, single binary for prod.

## Dev workflow

```
Terminal 1: mise run dev-backend     # Go on :8080
Terminal 2: mise run dev-frontend    # Vite on :5173 w/ HMR, proxies /api → :8080
Browser:    http://localhost:5173
```

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
- `internal/embed/dist/index.html` is a stub checked into git so bare `go build` works. `mise run build` replaces it with real assets.
- `go-containerregistry` for all OCI/Docker image operations
- API is JSON over REST. Endpoints defined in `docs/ARCHITECTURE.md`.
