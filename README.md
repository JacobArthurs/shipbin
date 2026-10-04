# shipbin

[![GitHub Release](https://img.shields.io/github/v/release/JacobArthurs/shipbin)](https://github.com/JacobArthurs/shipbin/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/jacobarthurs/shipbin.svg)](https://pkg.go.dev/github.com/jacobarthurs/shipbin)
[![ci](https://img.shields.io/github/actions/workflow/status/JacobArthurs/shipbin/ci.yml?branch=main)](https://github.com/JacobArthurs/shipbin/actions/workflows/ci.yml)
[![go version](https://img.shields.io/github/go-mod/go-version/JacobArthurs/shipbin)](./go.mod)
[![License](https://img.shields.io/github/license/JacobArthurs/shipbin)](LICENSE)

Publish native binaries to npm and PyPI from any language. Built for GitHub Actions.

`shipbin` takes pre-built binaries for multiple platforms and publishes them as installable packages. Users run `npm install -g mytool` or `pip install mytool` and get the right binary for their platform automatically.

## GitHub Actions

The examples below use OIDC, no tokens or secrets required. Token-based auth is also supported; see [Authentication](#authentication).

### npm

Requires an npm org with [provenance](https://docs.npmjs.com/generating-provenance-statements) enabled. Because shipbin publishes a root package plus one package per platform, all package names must be accessible with your publish credentials.

```yaml
permissions:
  id-token: write

steps:
  - uses: jacobarthurs/shipbin@v1
    with:
      registry: npm
      name: mytool
      org: myorg
      artifacts: |
        linux/amd64:./dist/mytool-linux-amd64
        linux/arm64:./dist/mytool-linux-arm64
        darwin/amd64:./dist/mytool-darwin-amd64
        darwin/arm64:./dist/mytool-darwin-arm64
        windows/amd64:./dist/mytool-windows-amd64.exe
        windows/arm64:./dist/mytool-windows-arm64.exe
```

### PyPI

Requires a [trusted publisher](https://pypi.org/manage/account/publishing/) registered on PyPI for your repository.

```yaml
permissions:
  id-token: write

steps:
  - uses: jacobarthurs/shipbin@v1
    with:
      registry: pypi
      name: mytool
      artifacts: |
        linux/amd64:./dist/mytool-linux-amd64
        linux/arm64:./dist/mytool-linux-arm64
        darwin/amd64:./dist/mytool-darwin-amd64
        darwin/arm64:./dist/mytool-darwin-arm64
        windows/amd64:./dist/mytool-windows-amd64.exe
        windows/arm64:./dist/mytool-windows-arm64.exe
```

### Action inputs

| Input             | Required | Default  | Description                                               |
|-------------------|----------|----------|-----------------------------------------------------------|
| `registry`        | Yes      |          | `npm` or `pypi`                                           |
| `name`            | Yes      |          | Binary name                                               |
| `artifacts`       | Yes      |          | Newline-separated `os/arch:path` mappings                 |
| `version`         | No       |          | Release version. Defaults to the current git tag          |
| `summary`         | No       |          | Short package description                                 |
| `license`         | No       |          | License identifier (e.g. `MIT`, `Apache-2.0`)             |
| `readme`          | No       |          | Path to a README to include in the published package      |
| `no-readme`       | No       | `false`  | Disable README auto-detection and omit readme from package |
| `dry-run`         | No       | `false`  | Print what would be published without publishing          |
| `org`             | npm only |          | npm org scope. `myorg` produces `@myorg/mytool-linux-x64` |
| `tag`             | No       | `latest` | npm dist-tag (e.g. `latest`, `next`, `beta`)              |
| `provenance`      | No       | `true`   | Publish with npm provenance attestation                   |
| `shipbin-version` | No       | `latest` | Pin a specific shipbin version                            |

## How it works

### npm

For each artifact, shipbin builds a platform-specific package (e.g. `@myorg/mytool-linux-x64`) containing the binary. It then publishes a root package (`mytool`) that declares all platform packages as optional dependencies and includes a Node.js wrapper script. When users install the root package, npm resolves and installs only the package matching their platform.

### PyPI

For each artifact, shipbin builds a platform-specific wheel containing the binary and a Python shim (`__init__.py`). The shim locates and `exec`s the bundled binary at runtime. Each wheel targets a specific platform tag (e.g. `manylinux_2_17_x86_64`), so pip resolves and installs only the correct wheel for the user's platform.

## Installation

```sh
go install github.com/jacobarthurs/shipbin@latest
```

Or download a pre-built binary from the [releases page](https://github.com/jacobarthurs/shipbin/releases).

## Prerequisites

### npm

- **[npm](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)** must be installed and on your `PATH`. shipbin calls `npm publish` directly.
- The **npm org** (`--org`) must already exist on [npmjs.com](https://www.npmjs.com/org/create) before publishing. shipbin does not create it.

### Both

- **[git](https://git-scm.com/)** must be installed if you omit `--version`. shipbin uses `git describe --tags --exact-match` to resolve the version from the current tag.

## Supported platforms

| Go target         | npm suffix      | PyPI wheel tag                                           |
|-------------------|-----------------|----------------------------------------------------------|
| `linux/amd64`     | `linux-x64`     | `manylinux_2_17_x86_64.manylinux2014_x86_64`            |
| `linux/arm64`     | `linux-arm64`   | `manylinux_2_17_aarch64.manylinux2014_aarch64`           |
| `darwin/amd64`    | `darwin-x64`    | `macosx_10_12_x86_64`                                    |
| `darwin/arm64`    | `darwin-arm64`  | `macosx_11_0_arm64`                                      |
| `windows/amd64`   | `win32-x64`     | `win_amd64`                                              |
| `windows/arm64`   | `win32-arm64`   | `win_arm64`                                              |

You don't need to publish for all platforms, pass only the artifacts you have.

## Usage

### npm

```sh
shipbin npm \
  --name mytool \
  --org myorg \
  --artifact linux/amd64:./dist/mytool-linux-amd64 \
  --artifact linux/arm64:./dist/mytool-linux-arm64 \
  --artifact darwin/amd64:./dist/mytool-darwin-amd64 \
  --artifact darwin/arm64:./dist/mytool-darwin-arm64 \
  --artifact windows/amd64:./dist/mytool-windows-amd64.exe \
  --artifact windows/arm64:./dist/mytool-windows-arm64.exe \
  --summary "My new tool" \
  --license MIT \
  --readme ./README.md
```

This publishes:

- `@myorg/mytool-linux-x64@<version>`
- `@myorg/mytool-linux-arm64@<version>`
- `@myorg/mytool-darwin-x64@<version>`
- `@myorg/mytool-darwin-arm64@<version>`
- `@myorg/mytool-win32-x64@<version>`
- `@myorg/mytool-win32-arm64@<version>`
- `mytool@<version>` (root package with optional dependencies + wrapper script)

### PyPI

```sh
shipbin pypi \
  --name mytool \
  --artifact linux/amd64:./dist/mytool-linux-amd64 \
  --artifact linux/arm64:./dist/mytool-linux-arm64 \
  --artifact darwin/amd64:./dist/mytool-darwin-amd64 \
  --artifact darwin/arm64:./dist/mytool-darwin-arm64 \
  --artifact windows/amd64:./dist/mytool-windows-amd64.exe \
  --artifact windows/arm64:./dist/mytool-windows-arm64.exe \
  --summary "My new tool" \
  --license MIT \
  --readme ./README.md
```

This publishes six platform-specific wheels to the `mytool` package on PyPI. pip automatically selects the correct wheel for the user's platform.

## Flags

### Common (all subcommands)

| Flag         | Required | Description |
|--------------|----------|-------------|
| `--name`     | Yes      | Binary name |
| `--artifact` | Yes      | `os/arch:path` mapping, repeatable |
| `--version`  | No       | Release version. Defaults to the current exact git tag |
| `--summary`  | No       | Short description included in package metadata |
| `--license`  | No       | License identifier (e.g. `MIT`, `Apache-2.0`) |
| `--readme`   | No       | Path to a README file to include in the published package |
| `--dry-run`  | No       | Print what would be published without publishing |

### npm

| Flag           | Default    | Description |
|----------------|------------|-------------|
| `--org`        | (required) | npm org scope. `myorg` produces `@myorg/mytool-linux-x64` |
| `--tag`        | `latest`   | dist-tag to publish under (e.g. `latest`, `next`, `beta`) |
| `--provenance` | `true`     | Publish with npm provenance attestation (requires CI) |
| `--repository` |            | Repository URL for `package.json`. Auto-detected in GitHub Actions. |

### Version resolution

If `--version` is not provided, shipbin runs `git describe --tags --exact-match` to read the version from the current git tag. The version must be valid semver (e.g. `1.2.3`, `1.2.3-beta.1`). A leading `v` prefix is stripped automatically. For PyPI, the version must also be valid PEP 440 (e.g. `1.0.0`, `1.0.0a1`, `1.0.0rc1`).

## Authentication

### npm

shipbin calls `npm publish` using whatever credentials are configured in your environment. Use [npm's built-in `NODE_AUTH_TOKEN`](https://docs.github.com/en/actions/publishing-packages/publishing-nodejs-packages)

Provenance attestation (`--provenance`) requires CI environment variables set by GitHub Actions. Use `--provenance=false` when publishing locally.

For local publishing, set `NODE_AUTH_TOKEN` to a token with 2FA disabled, otherwise the publish will be rejected.

### PyPI

shipbin supports two authentication methods:

1. **`PYPI_TOKEN` environment variable** — set this for local publishing or when using a classic API token in CI.
2. **GitHub OIDC trusted publisher** — if the GitHub Actions OIDC environment variables are present, shipbin mints a short-lived upload token automatically. This requires:
   - A [trusted publisher](https://pypi.org/manage/account/publishing/) registered on PyPI for your repository.
   - `id-token: write` permission in your workflow.

## Contributing

Contributions are welcome! To get started:

1. Fork the repository
2. Create a feature branch (`git checkout -b my-new-feature`)
3. Open a pull request

CI will automatically run tests and linting on your PR.

## License

This project is licensed under the [MIT License](LICENSE).
