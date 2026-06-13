#!/usr/bin/env node

const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const targets = {
  "win32-x64": "queryli-windows-amd64.exe",
  "linux-x64": "queryli-linux-amd64",
  "linux-arm64": "queryli-linux-arm64",
  "darwin-x64": "queryli-darwin-amd64",
  "darwin-arm64": "queryli-darwin-arm64"
};

const binaryName = targets[`${process.platform}-${process.arch}`];

if (!binaryName) {
  console.error(`queryli does not ship a binary for ${process.platform}-${process.arch}.`);
  process.exit(1);
}

const binaryPath = path.join(__dirname, "..", "dist", binaryName);

if (!fs.existsSync(binaryPath)) {
  console.error(`queryli binary is missing: ${binaryPath}`);
  console.error("Reinstall the package or run `npm run build:npm` from the source repository.");
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
  windowsHide: false
});

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

process.exit(result.status ?? 1);
