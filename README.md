<p align="center">
  <img src="web/public/favicon.svg" width="120" alt="peel logo" />
</p>

<h1 align="center">peel</h1>

<p align="center">
  A browser-based container image inspector. Like <a href="https://github.com/wagoodman/dive">dive</a>, but visual.
</p>

<p align="center">
  <a href="https://github.com/coffee-cup/peel/releases/latest"><img src="https://img.shields.io/github/release/coffee-cup/peel.svg" alt="GitHub release" /></a>
  <a href="https://github.com/coffee-cup/peel/actions/workflows/ci.yml"><img src="https://github.com/coffee-cup/peel/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <a href="https://github.com/coffee-cup/peel/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT" /></a>
</p>

---

<p align="center">
  <img src="docs/screenshot.png" alt="peel inspecting nginx image" width="800" />
</p>

## Features

- Layer-by-layer filesystem explorer with cumulative and diff views
- Syntax-highlighted file viewer for text, hex view for binaries
- Full keyboard navigation (arrow keys, vim bindings, tab between panels)
- Image metadata panel (ENV, ENTRYPOINT, CMD, labels, layer history)
- Whiteout/deletion tracking across layers
- Single static binary, no runtime dependencies

## Install

**Go install:**

```
go install github.com/coffee-cup/peel/cmd/peel@latest
```

**GitHub Releases:**

Pre-built binaries for macOS, Linux, and Windows are available on the [Releases](https://github.com/coffee-cup/peel/releases) page.

## Usage

```
peel <image-reference> [flags]
```

`<image-reference>` is a local image name/ID or remote registry reference (e.g. `myapp:latest`, `ghcr.io/org/repo:tag`).

**Flags:**

| Flag | Description |
|------|-------------|
| `--platform <os/arch>` | Target platform for multi-arch images (default: host) |
| `--no-open` | Don't auto-open browser |
| `--port <port>` | Port to listen on (default: 8080) |

## Build from source

```
mise run build
```

Builds the frontend, embeds it into the Go binary, and outputs `./peel`.
