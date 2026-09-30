import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

import { PLATFORMS, archiveName } from '../build.mjs';

const here = path.dirname(fileURLToPath(import.meta.url));
const buildScript = path.join(here, '..', 'build.mjs');
const require = createRequire(import.meta.url);
const VERSION = '1.2.3';

function tmp(prefix) {
  return fs.mkdtempSync(path.join(os.tmpdir(), prefix));
}

function sh(cmd, args, cwd) {
  const r = spawnSync(cmd, args, { cwd, encoding: 'utf8' });
  assert.equal(r.status, 0, `${cmd} ${args.join(' ')}: ${r.stderr || r.error}`);
}

// makeZip uses zip when present, else bsdtar (Windows, macOS).
function makeZip(cwd, archive, file) {
  const r = spawnSync('zip', ['-q', archive, file], { cwd });
  if (r.status !== 0) sh('tar', ['-a', '-cf', archive, file], cwd);
}

// fakeDist writes a tiny GoReleaser dist/: one archive per platform holding a
// fake binary, and checksums.txt.
function fakeDist() {
  const dist = tmp('ajiya-dist-');
  const work = tmp('ajiya-work-');
  const sums = [];
  for (const p of PLATFORMS) {
    const name = archiveName(VERSION, p);
    fs.writeFileSync(path.join(work, p.exe), `fake ${p.dir}\n`);
    if (name.endsWith('.zip')) makeZip(work, path.join(dist, name), p.exe);
    else sh('tar', ['-czf', path.join(dist, name), p.exe], work);
    fs.rmSync(path.join(work, p.exe));
    const sum = createHash('sha256').update(fs.readFileSync(path.join(dist, name))).digest('hex');
    sums.push(`${sum}  ${name}`);
  }
  fs.writeFileSync(path.join(dist, 'checksums.txt'), sums.join('\n') + '\n');
  fs.rmSync(work, { recursive: true, force: true });
  return dist;
}

function runBuild(dist, out, version = `v${VERSION}`) {
  return spawnSync(process.execPath, [buildScript, dist, version, '--out', out], { encoding: 'utf8' });
}

const readJSON = (f) => JSON.parse(fs.readFileSync(f, 'utf8'));

test('build writes seven packages with os, cpu, version and a 755 binary', () => {
  const dist = fakeDist();
  const out = path.join(tmp('ajiya-out-'), 'npm');
  const r = runBuild(dist, out);
  assert.equal(r.status, 0, r.stderr);

  const expected = { 'darwin-arm64': ['darwin', 'arm64'], 'darwin-x64': ['darwin', 'x64'], 'linux-arm64': ['linux', 'arm64'], 'linux-x64': ['linux', 'x64'], 'win32-arm64': ['win32', 'arm64'], 'win32-x64': ['win32', 'x64'] };
  for (const [dir, [osName, cpu]] of Object.entries(expected)) {
    const pkg = readJSON(path.join(out, dir, 'package.json'));
    assert.equal(pkg.name, `@ajiya/${dir}`);
    assert.equal(pkg.version, VERSION);
    assert.deepEqual(pkg.os, [osName]);
    assert.deepEqual(pkg.cpu, [cpu]);
    assert.deepEqual(pkg.files, ['bin']);
    assert.equal(pkg.preferUnplugged, true);
    assert.equal(pkg.license, 'MIT');
    const bin = path.join(out, dir, 'bin', osName === 'win32' ? 'ajiya.exe' : 'ajiya');
    assert.equal(fs.readFileSync(bin, 'utf8'), `fake ${dir}\n`);
    if (process.platform !== 'win32') assert.equal(fs.statSync(bin).mode & 0o777, 0o755);
  }

  const cli = readJSON(path.join(out, 'cli', 'package.json'));
  assert.equal(cli.name, '@ajiya/cli');
  assert.equal(cli.version, VERSION);
  assert.deepEqual(cli.bin, { ajiya: 'bin/ajiya.js' });
  assert.deepEqual(cli.optionalDependencies, Object.fromEntries(Object.keys(expected).map((d) => [`@ajiya/${d}`, VERSION])));
  assert.ok(fs.existsSync(path.join(out, 'cli', 'README.md')));
  assert.ok(fs.existsSync(path.join(out, 'cli', 'LICENSE')));
});

test('build output is deterministic', () => {
  const dist = fakeDist();
  const a = path.join(tmp('ajiya-out-'), 'npm');
  const b = path.join(tmp('ajiya-out-'), 'npm');
  assert.equal(runBuild(dist, a).status, 0);
  assert.equal(runBuild(dist, b, VERSION).status, 0);
  for (const dir of fs.readdirSync(a)) {
    for (const f of ['package.json', 'README.md']) {
      assert.equal(fs.readFileSync(path.join(a, dir, f), 'utf8'), fs.readFileSync(path.join(b, dir, f), 'utf8'));
    }
  }
});

test('a corrupted archive fails with the mismatch message and writes nothing', () => {
  const dist = fakeDist();
  const victim = path.join(dist, archiveName(VERSION, PLATFORMS[3]));
  fs.appendFileSync(victim, 'corrupt');
  const out = path.join(tmp('ajiya-out-'), 'npm');
  const r = runBuild(dist, out);
  assert.notEqual(r.status, 0);
  assert.match(r.stderr, /checksum mismatch for ajiya_1\.2\.3_linux_amd64\.tar\.gz/);
  assert.equal(fs.existsSync(out), false);
});

test('a missing archive fails and writes nothing', () => {
  const dist = fakeDist();
  fs.rmSync(path.join(dist, archiveName(VERSION, PLATFORMS[5])));
  const out = path.join(tmp('ajiya-out-'), 'npm');
  const r = runBuild(dist, out);
  assert.notEqual(r.status, 0);
  assert.match(r.stderr, /missing archive ajiya_1\.2\.3_windows_amd64\.zip/);
  assert.equal(fs.existsSync(out), false);
});

test('the launcher table and the build table agree', () => {
  const { PACKAGES } = require('../cli/bin/ajiya.js');
  assert.deepEqual(
    Object.entries(PACKAGES).sort(),
    PLATFORMS.map((p) => [p.dir, p.package]).sort(),
  );
});
