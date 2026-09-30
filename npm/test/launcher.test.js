'use strict';
const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { test } = require('node:test');

const launcherSource = path.join(__dirname, '..', 'cli', 'bin', 'ajiya.js');
const { platformPackage, missingMessage } = require(launcherSource);

test('platformPackage maps the six supported platforms', () => {
  assert.deepEqual(platformPackage('darwin', 'arm64'), { name: '@ajiya/darwin-arm64', binary: '@ajiya/darwin-arm64/bin/ajiya' });
  assert.deepEqual(platformPackage('darwin', 'x64'), { name: '@ajiya/darwin-x64', binary: '@ajiya/darwin-x64/bin/ajiya' });
  assert.deepEqual(platformPackage('linux', 'arm64'), { name: '@ajiya/linux-arm64', binary: '@ajiya/linux-arm64/bin/ajiya' });
  assert.deepEqual(platformPackage('linux', 'x64'), { name: '@ajiya/linux-x64', binary: '@ajiya/linux-x64/bin/ajiya' });
  assert.deepEqual(platformPackage('win32', 'arm64'), { name: '@ajiya/win32-arm64', binary: '@ajiya/win32-arm64/bin/ajiya.exe' });
  assert.deepEqual(platformPackage('win32', 'x64'), { name: '@ajiya/win32-x64', binary: '@ajiya/win32-x64/bin/ajiya.exe' });
});

test('platformPackage returns null for anything else', () => {
  assert.equal(platformPackage('freebsd', 'x64'), null);
  assert.equal(platformPackage('linux', 'ia32'), null);
});

test('missingMessage names the package, the install script and Homebrew', () => {
  const m = missingMessage('linux', 'x64');
  assert.match(m, /@ajiya\/linux-x64/);
  assert.match(m, /https:\/\/github\.com\/AliyuYahaya\/Ajiya\/releases\/latest\/download\/install\.sh/);
  assert.match(m, /brew install --cask AliyuYahaya\/tap\/ajiya/);
  assert.match(missingMessage('freebsd', 'x64'), /no npm package for freebsd x64/);
});

// installLauncher copies the launcher into a fresh folder shaped like an npm
// install: node_modules/@ajiya/cli/bin/ajiya.js.
function installLauncher() {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'ajiya-launcher-'));
  const bin = path.join(root, 'node_modules', '@ajiya', 'cli', 'bin');
  fs.mkdirSync(bin, { recursive: true });
  fs.copyFileSync(launcherSource, path.join(bin, 'ajiya.js'));
  return { root, script: path.join(bin, 'ajiya.js') };
}

test('with no platform package installed the launcher prints one error and exits 1', () => {
  const { script } = installLauncher();
  const r = spawnSync(process.execPath, [script, 'version'], { encoding: 'utf8' });
  assert.equal(r.status, 1);
  assert.match(r.stderr, /is not installed|no npm package/);
  assert.match(r.stderr, /install\.sh/);
  assert.match(r.stderr, /brew install --cask AliyuYahaya\/tap\/ajiya/);
  assert.equal(r.stdout, '');
});

// buildFakeBinary compiles a tiny real program: it prints its arguments after
// the first as JSON and exits with the code in the first argument. A real
// executable works the same on every operating system, unlike a shell script.
function buildFakeBinary(out) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'ajiya-fake-'));
  const src = path.join(dir, 'main.go');
  fs.writeFileSync(src, [
    'package main',
    'import ("encoding/json"; "os"; "strconv")',
    'func main() {',
    '  code, _ := strconv.Atoi(os.Args[1])',
    '  b, _ := json.Marshal(os.Args[2:])',
    '  os.Stdout.Write(append(b, 10))',
    '  os.Exit(code)',
    '}',
    '',
  ].join('\n'));
  return spawnSync('go', ['build', '-o', out, src], { cwd: dir, encoding: 'utf8' });
}

const goCheck = spawnSync('go', ['version'], { encoding: 'utf8' });
const hasGo = !goCheck.error && goCheck.status === 0;

test('the launcher runs the platform binary with its arguments and exit status', { skip: !hasGo && 'go is not installed, so the stand-in binary cannot be built' }, () => {
  const { root, script } = installLauncher();
  const pkg = platformPackage(process.platform, process.arch);
  assert.ok(pkg, 'this test needs a supported platform');
  const binDir = path.join(root, 'node_modules', pkg.name, 'bin');
  fs.mkdirSync(binDir, { recursive: true });
  fs.writeFileSync(path.join(root, 'node_modules', pkg.name, 'package.json'), JSON.stringify({ name: pkg.name, version: '0.0.0' }));
  const built = buildFakeBinary(path.join(binDir, path.basename(pkg.binary)));
  assert.equal(built.status, 0, built.stderr);

  const args = ['ticket', 'show', 'two words', '--flag', '-x', 'a"b'];
  for (const code of [0, 3]) {
    const r = spawnSync(process.execPath, [script, String(code), ...args], { encoding: 'utf8' });
    assert.equal(r.status, code);
    assert.deepEqual(JSON.parse(r.stdout), args);
  }
});
