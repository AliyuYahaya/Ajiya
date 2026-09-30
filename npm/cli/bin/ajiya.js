#!/usr/bin/env node
'use strict';
// Launcher for the Ajiya binary. It finds the platform package that npm
// installed next to this one (an optional dependency), runs its binary with
// the same arguments and exits with the same status.

const { spawnSync } = require('node:child_process');

const PACKAGES = {
  'darwin-arm64': '@ajiya/darwin-arm64',
  'darwin-x64': '@ajiya/darwin-x64',
  'linux-arm64': '@ajiya/linux-arm64',
  'linux-x64': '@ajiya/linux-x64',
  'win32-arm64': '@ajiya/win32-arm64',
  'win32-x64': '@ajiya/win32-x64',
};

// platformPackage maps process.platform and process.arch to the package that
// holds the binary, or returns null when the platform is not supported.
function platformPackage(platform, arch) {
  const name = PACKAGES[`${platform}-${arch}`];
  if (!name) return null;
  const file = platform === 'win32' ? 'bin/ajiya.exe' : 'bin/ajiya';
  return { name, binary: `${name}/${file}` };
}

// missingMessage is the one error shown when no binary can be found.
function missingMessage(platform, arch) {
  const pkg = platformPackage(platform, arch);
  const why = pkg
    ? `The package ${pkg.name} is not installed (was it installed with --no-optional or --omit=optional?).`
    : `Ajiya has no npm package for ${platform} ${arch}.`;
  return [
    `ajiya: ${why}`,
    'Install Ajiya another way:',
    '  curl -fsSL https://github.com/AliyuYahaya/Ajiya/releases/latest/download/install.sh | sh',
    '  brew install --cask AliyuYahaya/tap/ajiya',
  ].join('\n');
}

function main() {
  const pkg = platformPackage(process.platform, process.arch);
  let binary = null;
  if (pkg) {
    try {
      binary = require.resolve(pkg.binary);
    } catch (err) {
      binary = null;
    }
  }
  if (!binary) {
    process.stderr.write(missingMessage(process.platform, process.arch) + '\n');
    process.exit(1);
  }
  const result = spawnSync(binary, process.argv.slice(2), { stdio: 'inherit' });
  if (result.error) {
    process.stderr.write(`ajiya: cannot run ${binary}: ${result.error.message}\n`);
    process.exit(1);
  }
  if (result.signal) {
    process.kill(process.pid, result.signal);
    return;
  }
  process.exit(result.status === null ? 1 : result.status);
}

module.exports = { PACKAGES, platformPackage, missingMessage };

if (require.main === module) main();
