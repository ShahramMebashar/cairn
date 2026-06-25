#!/usr/bin/env node
// Stamps a release version (from a git tag like v1.2.3) into the desktop manifests
// so installers carry the right version. Used by the release workflow.
import { readFileSync, writeFileSync } from "node:fs";

const rawVersion = process.argv[2];

if (!rawVersion) {
  throw new Error("usage: node scripts/set-desktop-version.mjs <version-or-tag>");
}

const version = rawVersion.replace(/^refs\/tags\//, "").replace(/^v/, "");

const configPath = "desktop/src-tauri/tauri.conf.json";
const config = JSON.parse(readFileSync(configPath, "utf8"));
config.version = version;
writeFileSync(configPath, `${JSON.stringify(config, null, 2)}\n`);

const cargoPath = "desktop/src-tauri/Cargo.toml";
const cargo = readFileSync(cargoPath, "utf8");
writeFileSync(cargoPath, cargo.replace(/^version = "[^"]+"/m, `version = "${version}"`));

console.log(`desktop version set to ${version}`);
