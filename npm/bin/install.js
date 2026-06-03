#!/usr/bin/env node
"use strict";

// Downloads the native `adl` binary that matches the host platform from the
// GitHub release matching this package's version, and places it next to this
// script so `bin/run.js` can exec it. Runs as an npm `postinstall` hook.

const fs = require("fs");
const os = require("os");
const path = require("path");
const https = require("https");
const crypto = require("crypto");
const { spawnSync } = require("child_process");

const REPO = "inference-gateway/adl-cli";
const { version } = require("../package.json");
const BASE_URL =
  process.env.ADL_CLI_BASE_URL ||
  `https://github.com/${REPO}/releases/download/v${version}`;

// Maps Node's process.platform / process.arch onto the goreleaser asset names.
const PLATFORMS = { linux: "linux", darwin: "darwin" };
const ARCHES = { x64: "amd64", arm64: "arm64" };

const binName = process.platform === "win32" ? "adl.exe" : "adl";
const binPath = path.join(__dirname, binName);

function fail(message) {
  console.error(`\nadl-cli: ${message}\n`);
  process.exit(1);
}

function resolveAsset() {
  const goos = PLATFORMS[process.platform];
  const goarch = ARCHES[process.arch];
  if (!goos || !goarch) {
    fail(
      `unsupported platform ${process.platform}/${process.arch}. ` +
        `Prebuilt binaries are published for linux and darwin on amd64/arm64 only. ` +
        `Install from source instead: https://github.com/${REPO}#installation`
    );
  }
  return `adl_${goos}_${goarch}.tar.gz`;
}

function download(url) {
  return new Promise((resolve, reject) => {
    https
      .get(url, { headers: { "User-Agent": "adl-cli-npm" } }, (res) => {
        const { statusCode, headers } = res;
        if (statusCode >= 300 && statusCode < 400 && headers.location) {
          res.resume();
          resolve(download(headers.location));
          return;
        }
        if (statusCode !== 200) {
          res.resume();
          reject(
            new Error(`request to ${url} failed with status ${statusCode}`)
          );
          return;
        }
        const chunks = [];
        res.on("data", (chunk) => chunks.push(chunk));
        res.on("end", () => resolve(Buffer.concat(chunks)));
      })
      .on("error", reject);
  });
}

async function verifyChecksum(archive, assetName) {
  let checksums;
  try {
    checksums = (await download(`${BASE_URL}/checksums.txt`)).toString("utf8");
  } catch (err) {
    console.warn(
      `adl-cli: could not fetch checksums.txt (${err.message}); skipping integrity check`
    );
    return;
  }
  const line = checksums
    .split("\n")
    .find((l) => l.trim().endsWith(assetName));
  if (!line) {
    console.warn(
      `adl-cli: no checksum entry for ${assetName}; skipping integrity check`
    );
    return;
  }
  const expected = line.trim().split(/\s+/)[0];
  const actual = crypto.createHash("sha256").update(archive).digest("hex");
  if (expected !== actual) {
    fail(
      `checksum mismatch for ${assetName} (expected ${expected}, got ${actual})`
    );
  }
}

function extract(archive, assetName) {
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "adl-cli-"));
  const archivePath = path.join(tmpDir, assetName);
  try {
    fs.writeFileSync(archivePath, archive);
    const result = spawnSync("tar", ["-xzf", archivePath, "-C", tmpDir], {
      stdio: "inherit",
    });
    if (result.error || result.status !== 0) {
      fail(
        `failed to extract ${assetName}: ${
          result.error ? result.error.message : `tar exited with ${result.status}`
        }`
      );
    }
    const extracted = path.join(tmpDir, binName);
    if (!fs.existsSync(extracted)) {
      fail(`binary "${binName}" not found inside ${assetName}`);
    }
    fs.mkdirSync(path.dirname(binPath), { recursive: true });
    fs.copyFileSync(extracted, binPath);
    fs.chmodSync(binPath, 0o755);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

async function main() {
  if (process.env.ADL_CLI_SKIP_DOWNLOAD) {
    console.log("adl-cli: ADL_CLI_SKIP_DOWNLOAD set, skipping binary download");
    return;
  }
  if (fs.existsSync(binPath)) {
    return;
  }

  const assetName = resolveAsset();
  const url = `${BASE_URL}/${assetName}`;
  console.log(`adl-cli: downloading ${assetName} (v${version})`);

  let archive;
  try {
    archive = await download(url);
  } catch (err) {
    fail(
      `failed to download ${url}: ${err.message}. ` +
        `If you are offline, install from source: https://github.com/${REPO}#installation`
    );
  }

  await verifyChecksum(archive, assetName);
  extract(archive, assetName);
  console.log(`adl-cli: installed adl ${version} to ${binPath}`);
}

main().catch((err) => fail(err.message));
