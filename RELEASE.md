# Releasing

Releases are triggered by pushing a `v*` tag. GitHub Actions handles the rest.

## How It Works

1. Push a `v*` tag triggers the [release workflow](.github/workflows/release.yml)
2. CI builds the frontend (`bun run build`)
3. Frontend assets are copied to `internal/embed/dist/`
4. GoReleaser builds multi-platform binaries and creates a GitHub Release

### Platforms

| OS      | amd64 | arm64 |
|---------|-------|-------|
| macOS   | yes   | yes   |
| Linux   | yes   | yes   |
| Windows | yes   | —     |

## Steps

```sh
git tag v0.X.0
git push origin v0.X.0
```

GitHub Actions builds binaries and publishes the release.

## Installing from Releases

```sh
curl -fsSL https://raw.githubusercontent.com/coffee-cup/peel/main/install.sh | bash
```

Or with options:

```sh
curl -fsSL .../install.sh | bash -s -- -b ~/.local/bin -f
```

The install script auto-detects OS/arch and downloads the correct binary.
