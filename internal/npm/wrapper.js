#!/usr/bin/env node
"use strict";

const { execFileSync } = require("child_process");

const BIN_NAME = "__BIN_NAME__";
const ORG_NAME = "__ORG_NAME__";

const platforms = {
  "linux-x64":    `@${ORG_NAME}/${BIN_NAME}-linux-x64`,
  "linux-arm64":  `@${ORG_NAME}/${BIN_NAME}-linux-arm64`,
  "darwin-x64":   `@${ORG_NAME}/${BIN_NAME}-darwin-x64`,
  "darwin-arm64": `@${ORG_NAME}/${BIN_NAME}-darwin-arm64`,
  "win32-x64":    `@${ORG_NAME}/${BIN_NAME}-win32-x64`,
  "win32-arm64":  `@${ORG_NAME}/${BIN_NAME}-win32-arm64`,
};

const key = `${process.platform}-${process.arch === "x64" ? "x64" : process.arch}`;
const pkg = platforms[key];

if (!pkg) {
  console.error(
    `${BIN_NAME}: unsupported platform ${process.platform}/${process.arch}\n` +
    `supported platforms: ${Object.keys(platforms).join(", ")}`
  );
  process.exit(1);
}

let binPath;
try {
  binPath = require.resolve(`${pkg}/bin/${BIN_NAME}`);
} catch (e) {
  console.error(
    `${BIN_NAME}: could not find platform package ${pkg}\n` +
    `try reinstalling: npm install -g ${BIN_NAME}`
  );
  process.exit(1);
}

try {
  execFileSync(binPath, process.argv.slice(2), { stdio: "inherit" });
} catch (e) {
  process.exit(e.status ?? 1);
}
