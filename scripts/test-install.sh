#!/bin/sh
# Tests scripts/install.sh against a local fixture release.
#
# Usage: scripts/test-install.sh
#
# Environment:
#   AJIYA_DIST     folder holding the archives and checksums.txt of a GoReleaser
#                  snapshot (default: dist/, built here with goreleaser if missing)
#   INSTALL_SH     script under test (default: scripts/install.sh next to this file)
#   TEST_SHELLS    shells to run it with (default: "sh bash")
#   GORELEASER     goreleaser binary (default: goreleaser on PATH)
#
# Nothing is installed outside a temporary folder, and the real PATH and shell
# files are never touched. Needs python3 to serve the fixture over HTTP.
# Exit status: 0 when every check passes.

set -eu

here=$(cd "$(dirname "$0")" && pwd)
root=$(cd "$here/.." && pwd)
INSTALL_SH=${INSTALL_SH:-$here/install.sh}
TEST_SHELLS=${TEST_SHELLS:-sh bash}
dist=${AJIYA_DIST:-$root/dist}

fails=0
pass() { printf 'ok   %s\n' "$*"; }
fail() { printf 'FAIL %s\n' "$*"; fails=$((fails + 1)); }
check() { # check NAME COMMAND...: pass when the command succeeds
  name=$1
  shift
  if "$@" >/dev/null 2>&1; then pass "$name"; else fail "$name"; fi
}
contains() { grep -q -- "$1" "$2"; }

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  *) arch=arm64 ;;
esac
# Under Rosetta uname says x86_64 but the installer picks arm64 on purpose.
if [ "$os" = darwin ] && [ "$(sysctl -n sysctl.proc_translated 2>/dev/null || echo 0)" = 1 ]; then
  arch=arm64
fi

work=$(mktemp -d "${TMPDIR:-/tmp}/ajiya-test-install.XXXXXX")
server_pid=""
cleanup() {
  [ -z "$server_pid" ] || kill "$server_pid" 2>/dev/null || true
  rm -rf "$work"
}
trap cleanup EXIT INT TERM HUP

# 1. The snapshot to serve.
if ! ls "$dist"/ajiya_*_"${os}"_"${arch}".tar.gz >/dev/null 2>&1; then
  gr=${GORELEASER:-goreleaser}
  command -v "$gr" >/dev/null 2>&1 || { echo "no snapshot in $dist and no goreleaser; set AJIYA_DIST or GORELEASER" >&2; exit 2; }
  echo "building a snapshot with goreleaser"
  (cd "$root" && HOMEBREW_TAP_TOKEN=unused "$gr" release --snapshot --clean --skip=publish >"$work/goreleaser.log" 2>&1) ||
    { cat "$work/goreleaser.log" >&2; exit 2; }
fi
archive=$(cd "$dist" && ls ajiya_*_"${os}"_"${arch}".tar.gz | head -n 1)
version=${archive#ajiya_}
version=${version%_"${os}"_"${arch}".tar.gz}
tag="v$version"
echo "fixture: $archive (tag $tag)"

# 2. Lay it out like GitHub releases and serve it. /releases/latest redirects
#    to /releases/tag/<tag>, like GitHub does.
site=$work/site
mkdir -p "$site/releases/download/$tag" "$site/releases/tag"
cp "$dist"/ajiya_*.tar.gz "$dist"/ajiya_*.zip "$dist/checksums.txt" "$site/releases/download/$tag/"
cat >"$work/serve.py" <<PY
import http.server, socketserver, sys
tag = "$tag"
class H(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/releases/latest":
            self.send_response(302)
            self.send_header("Location", "/releases/tag/" + tag)
            self.end_headers()
        elif self.path.startswith("/releases/tag/"):
            self.send_response(200)
            self.send_header("Content-Length", "0")
            self.end_headers()
        else:
            super().do_GET()
    def log_message(self, *a): pass
socketserver.TCPServer.allow_reuse_address = True
with socketserver.TCPServer(("127.0.0.1", 0), lambda *a, **k: H(*a, directory="$site", **k)) as s:
    open("$work/port", "w").write(str(s.server_address[1]))
    s.serve_forever()
PY
python3 "$work/serve.py" &
server_pid=$!
i=0
while [ ! -s "$work/port" ]; do
  i=$((i + 1))
  [ "$i" -lt 100 ] || { echo "fixture server did not start" >&2; exit 2; }
  sleep 0.1
done
url="http://127.0.0.1:$(cat "$work/port")"

# 3. Run the script with each shell.
n=0
for sh_bin in $TEST_SHELLS; do
  command -v "$sh_bin" >/dev/null 2>&1 || { fail "$sh_bin is not installed"; continue; }
  n=$((n + 1))
  d=$work/run$n
  mkdir -p "$d/home"
  run() { # run INSTALLDIR ARGS...: install.sh with the fixture, output in $d/out
    idir=$1
    shift
    env -u AJIYA_VERSION HOME="$d/home" PATH="$PATH" AJIYA_NO_PROMPT=1 \
      AJIYA_BASE_URL="$url/releases/download" AJIYA_LATEST_URL="$url/releases/latest" \
      AJIYA_INSTALL_DIR="$idir" "$@" >"$d/out" 2>&1
  }

  echo "-- $sh_bin"
  check "$sh_bin: --help exits 0" "$sh_bin" "$INSTALL_SH" --help
  if "$sh_bin" "$INSTALL_SH" --bogus >/dev/null 2>&1; then fail "$sh_bin: unknown option is refused"; else pass "$sh_bin: unknown option is refused"; fi

  # Install a pinned version.
  if run "$d/bin" env AJIYA_VERSION="$tag" "$sh_bin" "$INSTALL_SH"; then pass "$sh_bin: install exits 0"; else fail "$sh_bin: install exits 0"; cat "$d/out"; fi
  check "$sh_bin: binary is executable" test -x "$d/bin/ajiya"
  if "$d/bin/ajiya" version 2>&1 | grep -q -F "$version"; then pass "$sh_bin: ajiya version reports $version"; else fail "$sh_bin: ajiya version reports $version"; fi
  contains "Checksum verified" "$d/out" && pass "$sh_bin: checksum was verified" || fail "$sh_bin: checksum was verified"
  contains "export PATH=\"$d/bin:" "$d/out" && pass "$sh_bin: PATH advice printed" || fail "$sh_bin: PATH advice printed"
  contains "ajiya mcp install" "$d/out" && pass "$sh_bin: mcp hint printed without a prompt" || fail "$sh_bin: mcp hint printed without a prompt"

  # No PATH advice when the folder is already on PATH; version without the v; no home files touched.
  if run "$d/bin" env PATH="$d/bin:$PATH" AJIYA_VERSION="$version" "$sh_bin" "$INSTALL_SH"; then pass "$sh_bin: reinstall with a bare version"; else fail "$sh_bin: reinstall with a bare version"; fi
  contains "export PATH" "$d/out" && fail "$sh_bin: no PATH advice when on PATH" || pass "$sh_bin: no PATH advice when on PATH"
  [ -z "$(ls -A "$d/home")" ] && pass "$sh_bin: home folder untouched" || fail "$sh_bin: home folder untouched"

  # Latest release, resolved through the redirect, into a new folder.
  if run "$d/latest" "$sh_bin" "$INSTALL_SH"; then pass "$sh_bin: latest release resolved"; else fail "$sh_bin: latest release resolved"; cat "$d/out"; fi
  check "$sh_bin: latest binary runs" "$d/latest/ajiya" version

  # Default folder is ~/.local/bin.
  if env -u AJIYA_INSTALL_DIR -u AJIYA_VERSION HOME="$d/home" AJIYA_NO_PROMPT=1 AJIYA_BASE_URL="$url/releases/download" \
    AJIYA_LATEST_URL="$url/releases/latest" "$sh_bin" "$INSTALL_SH" >"$d/out" 2>&1 && [ -x "$d/home/.local/bin/ajiya" ]; then
    pass "$sh_bin: default folder is ~/.local/bin"
  else
    fail "$sh_bin: default folder is ~/.local/bin"
  fi

  # A tampered archive is refused and nothing is installed.
  mkdir -p "$site/tampered-run$n/$tag"
  cp "$site/releases/download/$tag/"* "$site/tampered-run$n/$tag/"
  printf tampered >>"$site/tampered-run$n/$tag/$archive"
  if env HOME="$d/home" AJIYA_NO_PROMPT=1 AJIYA_VERSION="$tag" AJIYA_INSTALL_DIR="$d/tbin" \
    AJIYA_BASE_URL="$url/tampered-run$n" "$sh_bin" "$INSTALL_SH" >"$d/out" 2>&1; then
    fail "$sh_bin: tampered archive is refused"
  else
    pass "$sh_bin: tampered archive is refused"
  fi
  contains "checksum mismatch" "$d/out" && pass "$sh_bin: mismatch is reported" || fail "$sh_bin: mismatch is reported"
  [ ! -e "$d/tbin/ajiya" ] && pass "$sh_bin: nothing installed after a mismatch" || fail "$sh_bin: nothing installed after a mismatch"

  # A missing release fails cleanly.
  if run "$d/nbin" env AJIYA_VERSION=v9.9.9 "$sh_bin" "$INSTALL_SH"; then fail "$sh_bin: unknown version fails"; else pass "$sh_bin: unknown version fails"; fi
  [ ! -e "$d/nbin/ajiya" ] && pass "$sh_bin: nothing installed for an unknown version" || fail "$sh_bin: nothing installed for an unknown version"

  # Uninstall.
  if run "$d/bin" "$sh_bin" "$INSTALL_SH" --uninstall; then pass "$sh_bin: --uninstall exits 0"; else fail "$sh_bin: --uninstall exits 0"; fi
  [ ! -e "$d/bin/ajiya" ] && pass "$sh_bin: binary removed" || fail "$sh_bin: binary removed"
  check "$sh_bin: --uninstall twice is fine" run "$d/bin" "$sh_bin" "$INSTALL_SH" --uninstall
done

echo
if [ "$fails" -eq 0 ]; then echo "all install.sh checks passed"; else echo "$fails check(s) failed"; exit 1; fi
