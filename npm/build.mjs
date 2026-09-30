#!/usr/bin/env node
// Builds the npm packages from a GoReleaser dist/ folder.
//
//   node npm/build.mjs <dist> [<version>] [--out <dir>] [--check]
//
// <dist>     GoReleaser output folder (archives and checksums.txt).
// <version>  Package version; a leading v is stripped. Defaults to the tag in
//            $GITHUB_REF_NAME, if that starts with v.
// --out      Output folder (default npm/dist). It is emptied on success only.
// --check    After building, run `npm pack --dry-run` in each of the seven
//            packages; exit non-zero if any fails.
//
// Every archive is verified against checksums.txt before anything is written.
// Extraction uses the system `tar` for .tar.gz, and for .zip: `tar -xf` on
// Windows (bsdtar reads zip) and `unzip` elsewhere. No npm dependencies.

import { createHash } from 'node:crypto';
import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));

// npm os/cpu names, and the GoReleaser names they come from.
export const PLATFORMS = [
  { os: 'darwin', cpu: 'arm64', goos: 'darwin', goarch: 'arm64' },
  { os: 'darwin', cpu: 'x64', goos: 'darwin', goarch: 'amd64' },
  { os: 'linux', cpu: 'arm64', goos: 'linux', goarch: 'arm64' },
  { os: 'linux', cpu: 'x64', goos: 'linux', goarch: 'amd64' },
  { os: 'win32', cpu: 'arm64', goos: 'windows', goarch: 'arm64' },
  { os: 'win32', cpu: 'x64', goos: 'windows', goarch: 'amd64' },
].map((p) => ({
  ...p,
  package: `@ajiya/${p.os}-${p.cpu}`,
  dir: `${p.os}-${p.cpu}`,
  exe: p.os === 'win32' ? 'ajiya.exe' : 'ajiya',
}));

export class BuildError extends Error {}

export function normaliseVersion(v) {
  return String(v).replace(/^v/, '');
}

export function archiveName(version, p) {
  const ext = p.goos === 'windows' ? 'zip' : 'tar.gz';
  return `ajiya_${version}_${p.goos}_${p.goarch}.${ext}`;
}

function sha256(file) {
  return createHash('sha256').update(fs.readFileSync(file)).digest('hex');
}

function parseChecksums(file) {
  if (!fs.existsSync(file)) throw new BuildError(`missing ${file}`);
  const sums = new Map();
  for (const line of fs.readFileSync(file, 'utf8').split(/\r?\n/)) {
    const m = /^([0-9a-fA-F]{64})\s+\*?(.+?)\s*$/.exec(line);
    if (m) sums.set(m[2], m[1].toLowerCase());
  }
  return sums;
}

function run(cmd, args, opts = {}) {
  const r = spawnSync(cmd, args, { encoding: 'utf8', ...opts });
  if (r.error || r.status !== 0) {
    const why = r.error ? r.error.message : (r.stderr || '').trim();
    throw new BuildError(`${cmd} ${args.join(' ')} failed: ${why}`);
  }
  return r;
}

function extract(archive, dest) {
  fs.mkdirSync(dest, { recursive: true });
  if (archive.endsWith('.zip')) {
    if (process.platform === 'win32') run('tar', ['-xf', archive, '-C', dest]);
    else run('unzip', ['-q', '-o', archive, '-d', dest]);
  } else {
    run('tar', ['-xzf', archive, '-C', dest]);
  }
}

function writeJSON(file, data) {
  fs.writeFileSync(file, JSON.stringify(data, null, 2) + '\n');
}

function readJSON(file) {
  return JSON.parse(fs.readFileSync(file, 'utf8'));
}

function fillTemplate(text, p) {
  return text
    .replaceAll('@ajiya/PLATFORM', p.package)
    .replaceAll('PLATFORM', `${p.os} ${p.cpu}`);
}

// build verifies dist, then writes the seven packages to out. It throws a
// BuildError before writing anything if verification fails.
export function build({ dist, version, out, repoRoot }) {
  version = normaliseVersion(version);
  if (!/^\d+\.\d+\.\d+([-+][0-9A-Za-z.+-]+)?$/.test(version)) {
    throw new BuildError(`bad version "${version}"`);
  }
  const sums = parseChecksums(path.join(dist, 'checksums.txt'));

  // 1. Verify every archive first.
  for (const p of PLATFORMS) {
    const name = archiveName(version, p);
    const file = path.join(dist, name);
    if (!fs.existsSync(file)) throw new BuildError(`missing archive ${name} in ${dist}`);
    const want = sums.get(name);
    if (!want) throw new BuildError(`${name} is not listed in checksums.txt`);
    const got = sha256(file);
    if (got !== want) {
      throw new BuildError(`checksum mismatch for ${name}: checksums.txt says ${want}, the file is ${got}`);
    }
  }

  // 2. Extract every binary into a staging folder.
  const stage = fs.mkdtempSync(path.join(os.tmpdir(), 'ajiya-npm-'));
  try {
    for (const p of PLATFORMS) {
      const to = path.join(stage, p.dir);
      extract(path.join(dist, archiveName(version, p)), to);
      if (!fs.existsSync(path.join(to, p.exe))) {
        throw new BuildError(`${archiveName(version, p)} has no ${p.exe}`);
      }
    }

    // 3. Write the packages.
    fs.rmSync(out, { recursive: true, force: true });
    const platformTemplate = readJSON(path.join(here, 'platform', 'package.json'));
    const platformReadme = fs.readFileSync(path.join(here, 'platform', 'README.md'), 'utf8');
    const licence = path.join(repoRoot, 'LICENSE');

    for (const p of PLATFORMS) {
      const dir = path.join(out, p.dir);
      fs.mkdirSync(path.join(dir, 'bin'), { recursive: true });
      const bin = path.join(dir, 'bin', p.exe);
      fs.copyFileSync(path.join(stage, p.dir, p.exe), bin);
      fs.chmodSync(bin, 0o755);
      const pkg = JSON.parse(fillTemplate(JSON.stringify(platformTemplate), p));
      pkg.version = version;
      pkg.os = [p.os];
      pkg.cpu = [p.cpu];
      writeJSON(path.join(dir, 'package.json'), pkg);
      fs.writeFileSync(path.join(dir, 'README.md'), fillTemplate(platformReadme, p));
      fs.copyFileSync(licence, path.join(dir, 'LICENSE'));
    }

    const cli = path.join(out, 'cli');
    fs.cpSync(path.join(here, 'cli'), cli, { recursive: true });
    fs.chmodSync(path.join(cli, 'bin', 'ajiya.js'), 0o755);
    const wrapper = readJSON(path.join(cli, 'package.json'));
    wrapper.version = version;
    wrapper.optionalDependencies = Object.fromEntries(PLATFORMS.map((p) => [p.package, version]));
    writeJSON(path.join(cli, 'package.json'), wrapper);
    fs.copyFileSync(path.join(repoRoot, 'README.md'), path.join(cli, 'README.md'));
    fs.copyFileSync(licence, path.join(cli, 'LICENSE'));
  } finally {
    fs.rmSync(stage, { recursive: true, force: true });
  }
  return [...PLATFORMS.map((p) => ({ name: p.package, dir: path.join(out, p.dir) })), { name: '@ajiya/cli', dir: path.join(out, 'cli') }];
}

// check runs `npm pack --dry-run` in each package and returns the summary lines.
export function check(packages) {
  const lines = [];
  let failed = false;
  for (const pkg of packages) {
    const r = spawnSync('npm', ['pack', '--dry-run', '--json'], {
      cwd: pkg.dir,
      encoding: 'utf8',
      shell: process.platform === 'win32',
    });
    if (r.status !== 0) {
      failed = true;
      lines.push(`FAIL ${pkg.name}: ${(r.stderr || r.error?.message || '').trim()}`);
      continue;
    }
    try {
      const info = JSON.parse(r.stdout)[0];
      lines.push(`ok   ${pkg.name}@${info.version}  ${info.files.length} files, ${info.size} bytes packed, ${info.unpackedSize} unpacked`);
    } catch {
      failed = true;
      lines.push(`FAIL ${pkg.name}: unreadable npm pack output`);
    }
  }
  return { lines, failed };
}

function main(argv) {
  const args = [];
  let out = path.join(here, 'dist');
  let doCheck = false;
  for (let i = 0; i < argv.length; i++) {
    if (argv[i] === '--check') doCheck = true;
    else if (argv[i] === '--out') out = path.resolve(argv[++i] ?? '');
    else args.push(argv[i]);
  }
  const dist = args[0];
  let version = args[1];
  if (!version && /^v\d/.test(process.env.GITHUB_REF_NAME ?? '')) version = process.env.GITHUB_REF_NAME;
  if (!dist || !version) {
    console.error('usage: node npm/build.mjs <dist> [<version>] [--out <dir>] [--check]');
    return 2;
  }
  try {
    const packages = build({ dist: path.resolve(dist), version, out, repoRoot: path.resolve(here, '..') });
    console.log(`wrote ${packages.length} packages to ${out}`);
    if (doCheck) {
      const { lines, failed } = check(packages);
      console.log(lines.join('\n'));
      return failed ? 1 : 0;
    }
    return 0;
  } catch (err) {
    if (err instanceof BuildError) {
      console.error(`error: ${err.message}`);
      return 1;
    }
    throw err;
  }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  process.exitCode = main(process.argv.slice(2));
}
