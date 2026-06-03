# @inference-gateway/adl-cli

npm distribution of [`adl`](https://github.com/inference-gateway/adl-cli) — the
CLI that generates complete A2A (Agent-to-Agent) project scaffolding from YAML
Agent Definition Language (ADL) manifests.

This package is a thin wrapper: on install it downloads the native `adl` binary
that matches your platform from the matching GitHub release, then proxies every
command to it. The CLI itself is unchanged, so `npx @inference-gateway/adl-cli --help`
produces the same output as a natively installed `adl --help`.

## Usage

Run without installing:

```bash
npx @inference-gateway/adl-cli init my-agent
npx @inference-gateway/adl-cli generate --file agent.yaml --output ./agent
npx @inference-gateway/adl-cli validate agent.yaml
```

Or install it (globally or as a dev dependency):

```bash
npm install -g @inference-gateway/adl-cli
adl --help
```

## Supported platforms

Prebuilt binaries are published for **Linux** and **macOS** on **x64** and
**arm64**. On other platforms, install from source — see the
[main README](https://github.com/inference-gateway/adl-cli#installation).

## How it works

- `bin/install.js` runs as a `postinstall` hook. It maps `process.platform` /
  `process.arch` to a release asset (`adl_linux_amd64.tar.gz`,
  `adl_darwin_arm64.tar.gz`, …), downloads it, verifies its SHA-256 against the
  release `checksums.txt`, extracts the binary, and stores it inside the package.
- `bin/run.js` is the `adl` entrypoint. It locates the downloaded binary and
  forwards all arguments and the exit code.

### Environment variables

- `ADL_CLI_SKIP_DOWNLOAD=1` — skip the binary download during install (useful in
  CI when the binary is provided another way).
- `ADL_CLI_BASE_URL` — override the release download base URL (advanced/testing).

## License

Apache-2.0. See [LICENSE](https://github.com/inference-gateway/adl-cli/blob/main/LICENSE).
