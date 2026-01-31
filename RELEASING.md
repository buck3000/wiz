# Releasing Wiz

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [GoReleaser](https://goreleaser.com/install/) (`brew install goreleaser`)
- [GitHub CLI](https://cli.github.com/) (`brew install gh`) — authenticated with `gh auth login`
- A production signing key (see below)

## Signing Key Setup

Wiz uses HMAC-SHA256 to sign license keys. The signing key is injected at build time via `-ldflags`.

### Generate a production signing key (once)

```bash
openssl rand -hex 32
```

Save this somewhere secure (1Password, environment secret, etc.). This is the key that signs all customer license JWTs. If you lose it, all issued licenses become invalid.

### Local development

No setup needed. Local `go build` and `go run` use a built-in dev key. License keys generated locally with `cmd/genkey` will only work with dev builds.

### CI / GoReleaser

Set `WIZ_SIGNING_KEY_HEX` as a repository secret in GitHub Actions, or export it before running goreleaser locally.

## Generating License Keys

The `cmd/genkey/` tool (gitignored, local only) generates license keys signed with the dev key:

```bash
go run ./cmd/genkey/
```

To generate keys signed with the production key, build genkey with the production key:

```bash
go build -ldflags "-X github.com/buck3000/wiz/internal/license.SigningKeyHex=YOUR_PROD_KEY_HEX" -o genkey-prod ./cmd/genkey/
./genkey-prod
```

Edit `cmd/genkey/main.go` to change the email, tier, or expiry.

## License Tiers

| Tier | Max Contexts | Orchestra Deps | Cost Tracking | AI Review | Team Features |
|------|-------------|----------------|---------------|-----------|---------------|
| Free | 10 | No | No | No | No |
| Pro | Unlimited | Yes | Yes | No | No |
| Team | Unlimited | Yes | Yes | Yes | Yes |
| Enterprise | Unlimited | Yes | Yes | Yes | Yes |

## User License Installation

Users activate a license in one of two ways:

**Environment variable** (recommended for CI):
```bash
export WIZ_LICENSE_KEY="ey..."
```

**License file** (recommended for personal use):
```bash
mkdir -p ~/.config/wiz
echo '{"key":"ey..."}' > ~/.config/wiz/license.json
```

## Cutting a Release

### 1. Ensure everything passes

```bash
go build ./...
go test ./...
```

### 2. Tag the release

```bash
git tag v0.1.0
git push origin v0.1.0
```

### 3a. Release via GitHub Actions (recommended)

Add this workflow at `.github/workflows/release.yaml`:

```yaml
name: Release
on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: "1.25"
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          WIZ_SIGNING_KEY_HEX: ${{ secrets.WIZ_SIGNING_KEY_HEX }}
```

Then add `WIZ_SIGNING_KEY_HEX` as a repository secret in GitHub Settings > Secrets.

### 3b. Release locally

```bash
export WIZ_SIGNING_KEY_HEX="your-production-hex-key"
goreleaser release --clean
```

### 4. Verify

The release will appear at `https://github.com/buck3000/wiz/releases` with:
- `wiz_VERSION_darwin_amd64.tar.gz`
- `wiz_VERSION_darwin_arm64.tar.gz`
- `wiz_VERSION_linux_amd64.tar.gz`
- `wiz_VERSION_linux_arm64.tar.gz`
- `wiz_VERSION_windows_amd64.zip`
- `checksums.txt`

## User Installation

### From source (requires Go)

```bash
go install github.com/buck3000/wiz@latest
```

Note: `go install` does not inject the production signing key. Users installing from source get the dev key and cannot validate production license keys. This is fine — source users can modify the code anyway.

### From release binaries

Download from the [releases page](https://github.com/buck3000/wiz/releases) and add to PATH:

```bash
tar xzf wiz_*_darwin_arm64.tar.gz
sudo mv wiz /usr/local/bin/
```

### Homebrew (future)

Once you create a `firewood-buck-3000/homebrew-tap` repo, uncomment the `brews` section in `.goreleaser.yaml`. Then users can:

```bash
brew install firewood-buck-3000/tap/wiz
```

## Dry Run

Test the release process without publishing:

```bash
export WIZ_SIGNING_KEY_HEX="your-production-hex-key"
goreleaser release --snapshot --clean
```

Binaries will be in `dist/`.
