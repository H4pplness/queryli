const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const root = path.resolve(__dirname, "..");
const dist = path.join(root, "dist");

const targets = [
  { goos: "windows", goarch: "amd64", output: "queryli-windows-amd64.exe" },
  { goos: "linux", goarch: "amd64", output: "queryli-linux-amd64" },
  { goos: "linux", goarch: "arm64", output: "queryli-linux-arm64" },
  { goos: "darwin", goarch: "amd64", output: "queryli-darwin-amd64" },
  { goos: "darwin", goarch: "arm64", output: "queryli-darwin-arm64" }
];

fs.rmSync(dist, { recursive: true, force: true });
fs.mkdirSync(dist, { recursive: true });

for (const target of targets) {
  const outputPath = path.join(dist, target.output);
  const env = {
    ...process.env,
    CGO_ENABLED: "0",
    GOOS: target.goos,
    GOARCH: target.goarch
  };

  console.log(`Building ${target.goos}/${target.goarch} -> dist/${target.output}`);

  const result = spawnSync(
    "go",
    ["build", "-trimpath", "-ldflags=-s -w", "-o", outputPath, "."],
    {
      cwd: root,
      env,
      stdio: "inherit"
    }
  );

  if (result.error) {
    console.error(result.error.message);
    process.exit(1);
  }

  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }

  if (target.goos !== "windows") {
    fs.chmodSync(outputPath, 0o755);
  }
}
