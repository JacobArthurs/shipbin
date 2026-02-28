# shipbin

[![GitHub Release](https://img.shields.io/github/v/release/JacobArthurs/shipbin)](https://github.com/JacobArthurs/shipbin/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/jacobarthurs/shipbin.svg)](https://pkg.go.dev/github.com/jacobarthurs/shipbin)
[![Go Report Card](https://goreportcard.com/badge/github.com/jacobarthurs/shipbin)](https://goreportcard.com/report/github.com/jacobarthurs/shipbin)
[![ci](https://img.shields.io/github/actions/workflow/status/JacobArthurs/shipbin/ci.yml?branch=main)](https://github.com/JacobArthurs/shipbin/actions/workflows/ci.yml)
[![go version](https://img.shields.io/github/go-mod/go-version/JacobArthurs/shipbin)](./go.mod)
[![License](https://img.shields.io/github/license/JacobArthurs/shipbin)](LICENSE)

Publish native binaries to npm and PyPI from any language. Built for GitHub Actions.

`shipbin` takes pre-built binaries for multiple platforms and publishes them as installable packages. Users run `npm install -g mytool` or `pip install mytool` and get the right binary for their platform automatically.

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
  --summary "My awesome tool" \
  --license MIT \
  --readme ./README.md
```

This publishes:

- `@myorg/mytool-linux-x64@<version>`
- `@myorg/mytool-linux-arm64@<version>`
- `@myorg/mytool-darwin-x64@<version>`
- `@myorg/mytool-darwin-arm64@<version>`
- `@myorg/mytool-win32-x64@<version>`
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
  --summary "My awesome tool" \
  --license MIT \
  --readme ./README.md
```

This publishes five platform-specific wheels to the `mytool` package on PyPI. pip automatically selects the correct wheel for the user's platform.

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

### Version resolution

If `--version` is not provided, shipbin runs `git describe --tags --exact-match` to read the version from the current git tag. The version must be valid semver (e.g. `1.2.3`, `1.2.3-beta.1`). A leading `v` prefix is stripped automatically. For PyPI, the version must also be valid PEP 440 (e.g. `1.0.0`, `1.0.0a1`, `1.0.0rc1`).

## Authentication

### npm

shipbin calls `npm publish` using whatever credentials are configured in your environment. In GitHub Actions, use [npm's built-in `NODE_AUTH_TOKEN`](https://docs.github.com/en/actions/publishing-packages/publishing-nodejs-packages):

```yaml
- uses: actions/setup-node@v4
  with:
    node-version: '20'
    registry-url: 'https://registry.npmjs.org'
- run: shipbin npm ...
  env:
    NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

Provenance attestation (`--provenance`) requires CI environment variables set by GitHub Actions. Use `--provenance=false` when publishing locally.

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
