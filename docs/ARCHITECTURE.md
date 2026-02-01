# peel

A CLI tool that inspects OCI/Docker container images and serves a web UI for exploring layers and filesystem contents. Inspired by [dive](https://github.com/wagoodman/dive).

## Overview

`peel` takes an image reference, extracts layer information, and opens a browser-based UI for inspecting the image's contents. The goal is a better visual experience for understanding what's in your container images—useful for debugging bloated images, testing builds, and understanding what buildkit LLB produces.

## CLI Behavior

```
peel <image-reference> [flags]
```

**Arguments:**

- `<image-reference>` — Local image name/ID or remote registry reference (e.g., `myapp:latest`, `ghcr.io/org/repo:tag`)

**Flags:**

- `--platform <os/arch>` — Target platform for multi-arch images (default: host architecture)
- `--no-open` — Don't auto-open browser
- `--port <port>` — Override random port selection (optional)

**Behavior:**

1. Resolve and pull/load image
2. Extract layer metadata and filesystem trees
3. Start ephemeral HTTP server on random available port
4. Auto-open browser to UI
5. Server terminates when process is killed

**Auth:** Defers to Docker credential helpers. MVP assumes local images or public registries.

## Features

### Layer List

- Display all layers in order (base → top)
- Show size per layer
- Selecting a layer updates the filesystem view

### Layer Diffing

- Show changes between layers: added, modified, deleted files
- Visual indicators for change type
- Whiteout files (`.wh.*`) exposed explicitly as deletions

### Filesystem Tree

- Cumulative filesystem state at selected layer
- Expandable directory tree
- Symlinks displayed as symlinks (not followed)
- File sizes shown

### File Content Viewer

- Text files: syntax highlighting based on extension/content
- Binary files: hex view with offset and ASCII columns
- Large files: truncate with option to load more

### Image Metadata

- Config display: ENV, ENTRYPOINT, CMD, WORKDIR, USER, LABELS
- Layer history/commands that created each layer
- Image digest, architecture, OS

### Keyboard Navigation

- Full keyboard accessibility
- Arrow keys for tree navigation
- Tab between panels
- Enter to expand/select
- `/` or similar for search (nice-to-have)

## Technical Architecture

### Backend (Go)

```
cmd/
  peel/
    main.go           # CLI entrypoint, flag parsing
internal/
  image/
    loader.go         # Image loading (local + remote)
    layer.go          # Layer extraction and diffing
    filesystem.go     # Filesystem tree construction
  server/
    server.go         # HTTP server setup
    handlers.go       # API route handlers
  embed/
    embed.go          # Embedded frontend assets
```

**Dependencies:**

- `google/go-containerregistry` — Image pulling, layer extraction, registry protocol
- Standard library for HTTP server

**API Endpoints:**

```
GET  /api/image          — Image metadata
GET  /api/layers         — Layer list with sizes
GET  /api/layers/:id/tree    — Filesystem tree for layer (cumulative)
GET  /api/layers/:id/diff    — Diff from previous layer
GET  /api/files/:layer/*path — File content (text or hex-encoded binary)
GET  /                   — Serve frontend (index.html)
GET  /*                  — Static assets
```

### Frontend (React + TypeScript)

```
web/
  src/
    components/
      LayerList.tsx
      FileTree.tsx
      FileViewer.tsx
      MetadataPanel.tsx
    hooks/
      useImage.ts
      useKeyboardNav.ts
    App.tsx
    main.tsx
  index.html
  vite.config.ts
  tailwind.config.js
```

**Stack:**

- Vite + React + TypeScript
- TailwindCSS for styling
- Syntax highlighting library (e.g., Shiki, Prism, or highlight.js)

**Layout:**

```
┌─────────────────────────────────────────────────────┐
│  peel — image:tag                          metadata │
├──────────────┬──────────────────────────────────────┤
│              │                                      │
│   Layers     │         File Tree                    │
│              │                                      │
│   [layer 1]  │    ├── bin/                         │
│   [layer 2]  │    ├── etc/                         │
│   [layer 3]◄─│    │   └── nginx.conf [modified]    │
│              │    └── var/                         │
│              │                                      │
├──────────────┴──────────────────────────────────────┤
│                                                     │
│   File Content Viewer                               │
│   (syntax highlighted or hex)                       │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Build & Development

**Development:**

```bash
# Terminal 1: Frontend with hot reload
cd web && npm run dev

# Terminal 2: Backend (watches for changes, proxies to Vite)
go run ./cmd/peel --dev <image>
```

Backend in dev mode proxies frontend requests to Vite dev server.

**Release Build:**

```bash
# Build frontend
cd web && npm run build

# Embed and build Go binary
go build -o peel ./cmd/peel
```

Frontend assets embedded via `//go:embed` directive.

**Distribution:** Single static binary, no runtime dependencies.

## Edge Cases

| Case                          | Behavior                                                   |
| ----------------------------- | ---------------------------------------------------------- |
| Multi-platform image          | Default to host arch, override with `--platform`           |
| Squashed image (single layer) | Show layer and tree, no diff (nothing to diff against)     |
| Empty layer                   | Show in list with 0 bytes, empty tree                      |
| Very large files              | Truncate content view, show file size, option to load more |
| Symlinks                      | Display as symlinks, show target path, don't follow        |
| Whiteout files                | Show in diff as explicit deletions                         |

## Non-Goals (MVP)

- Private registry authentication UI
- Multi-platform selection in UI (use CLI flag)
- Wasted space / efficiency analysis
- Image modification or export
- Comparing two different images
- Persistent server mode

## Future Considerations

- Private registry auth configuration
- Search within file tree and content
- Export layer as tarball
- Side-by-side image comparison
- Efficiency scoring (à la dive)
- Dockerfile/buildkit command correlation with layers
