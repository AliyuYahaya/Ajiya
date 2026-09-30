#!/bin/sh
# Ajiya installer for macOS and Linux.
#
#   curl -fsSL https://github.com/AliyuYahaya/Ajiya/releases/latest/download/install.sh | sh
#
# Downloads the release archive for this machine, checks it against the
# release's checksums.txt, and installs the binary into a folder you own.
# It never uses sudo and never edits your shell files.
#
# Environment:
#   AJIYA_VERSION      version to install, for example v0.2.0 (default: latest release)
#   AJIYA_INSTALL_DIR  where to put the binary (default: $HOME/.local/bin)
#   AJIYA_NO_PROMPT=1  never ask questions
#   AJIYA_BASE_URL     release download base (default: the GitHub releases URL);
#                      archives are read from $AJIYA_BASE_URL/<tag>/
#   AJIYA_LATEST_URL   page that redirects to the latest tag
#                      (default: https://github.com/AliyuYahaya/Ajiya/releases/latest)

set -eu

REPO_URL="https://github.com/AliyuYahaya/Ajiya"

say() { printf '%s\n' "$*"; }
err() { printf 'ajiya install: %s\n' "$*" >&2; }
die() { err "$*"; exit 1; }

usage() {
  cat <<'USAGE'
Install Ajiya.

Usage: install.sh [--uninstall] [--help]

  --uninstall  remove the ajiya binary from the install folder
  --help       show this help

Environment:
  AJIYA_VERSION      version to install, for example v0.2.0 (default: latest)
  AJIYA_INSTALL_DIR  install folder (default: $HOME/.local/bin)
  AJIYA_NO_PROMPT=1  never ask questions
  AJIYA_BASE_URL     release download base URL (for mirrors and tests)
USAGE
}

have() { command -v "$1" >/dev/null 2>&1; }

# fetch URL DEST: download to a file, failing on HTTP errors.
fetch() {
  if have curl; then
    curl -fsSL -o "$2" "$1"
  elif have wget; then
    wget -q -O "$2" "$1"
  else
    die "need curl or wget to download files"
  fi
}

detect_os() {
  case "$(uname -s)" in
    Darwin) echo darwin ;;
    Linux) echo linux ;;
    *) die "unsupported OS $(uname -s): this script covers macOS and Linux (Windows: install.ps1)" ;;
  esac
}

detect_arch() {
  os=$1
  case "$(uname -m)" in
    x86_64 | amd64) arch=amd64 ;;
    arm64 | aarch64) arch=arm64 ;;
    *) die "unsupported CPU architecture $(uname -m)" ;;
  esac
  # A shell running under Rosetta reports x86_64 on an Apple silicon Mac.
  if [ "$os" = darwin ] && [ "$arch" = amd64 ] &&
    [ "$(sysctl -n sysctl.proc_translated 2>/dev/null || echo 0)" = 1 ]; then
    arch=arm64
  fi
  echo "$arch"
}

# latest_tag: follow the /releases/latest redirect and read the tag from it.
latest_tag() {
  latest=${AJIYA_LATEST_URL:-$REPO_URL/releases/latest}
  if have curl; then
    final=$(curl -fsSL -o /dev/null -w '%{url_effective}' "$latest") || return 1
  elif have wget; then
    final=$(wget -S --spider --max-redirect=0 "$latest" 2>&1 |
      sed -n 's/^ *[Ll]ocation: *//p' | tr -d '\r' | tail -n 1)
  else
    die "need curl or wget to download files"
  fi
  case "$final" in
    */tag/*) echo "${final##*/tag/}" ;;
    *) return 1 ;;
  esac
}

sha256_of() {
  if have sha256sum; then
    sha256sum "$1" | awk '{print $1}'
  elif have shasum; then
    shasum -a 256 "$1" | awk '{print $1}'
  else
    die "need sha256sum or shasum to verify the download"
  fi
}

on_path() {
  case ":$PATH:" in
    *":$1:"*) return 0 ;;
    *) return 1 ;;
  esac
}

install_dir() {
  if [ -n "${AJIYA_INSTALL_DIR:-}" ]; then
    echo "$AJIYA_INSTALL_DIR"
  else
    [ -n "${HOME:-}" ] || die "HOME is not set; set AJIYA_INSTALL_DIR"
    echo "$HOME/.local/bin"
  fi
}

# have_terminal: true when a person can answer a question.
have_terminal() {
  [ "${AJIYA_NO_PROMPT:-}" = 1 ] && return 1
  # Opening /dev/tty fails when there is no controlling terminal.
  (: </dev/tty) 2>/dev/null
}

uninstall() {
  dir=$(install_dir)
  if [ -e "$dir/ajiya" ]; then
    rm -f "$dir/ajiya"
    say "Removed $dir/ajiya"
  else
    say "Nothing to remove: $dir/ajiya does not exist"
  fi
  say "Your projects and agent settings were not touched. To unregister from agents, run 'ajiya mcp uninstall' before removing the binary."
}

do_install() {
  os=$(detect_os)
  arch=$(detect_arch "$os")
  dir=$(install_dir)

  tag=${AJIYA_VERSION:-}
  if [ -z "$tag" ]; then
    say "Looking up the latest release"
    tag=$(latest_tag) || die "could not find the latest release; set AJIYA_VERSION, for example AJIYA_VERSION=v0.2.0"
  fi
  case "$tag" in v*) ;; *) tag="v$tag" ;; esac
  version=${tag#v}

  base=${AJIYA_BASE_URL:-$REPO_URL/releases/download}
  archive="ajiya_${version}_${os}_${arch}.tar.gz"

  tmp=$(mktemp -d "${TMPDIR:-/tmp}/ajiya-install.XXXXXX") || die "could not create a temporary folder"
  trap 'rm -rf "$tmp"' EXIT INT TERM HUP

  say "Downloading Ajiya $tag for $os/$arch"
  fetch "$base/$tag/$archive" "$tmp/$archive" || die "could not download $base/$tag/$archive"
  fetch "$base/$tag/checksums.txt" "$tmp/checksums.txt" || die "could not download $base/$tag/checksums.txt"

  want=$(awk -v f="$archive" '$2 == f || $2 == "*" f { print $1; exit }' "$tmp/checksums.txt")
  [ -n "$want" ] || die "checksums.txt has no entry for $archive"
  got=$(sha256_of "$tmp/$archive")
  if [ "$want" != "$got" ]; then
    err "checksum mismatch for $archive"
    err "  expected $want"
    err "  got      $got"
    die "refusing to install; nothing was changed"
  fi
  say "Checksum verified"

  mkdir "$tmp/x"
  tar -xzf "$tmp/$archive" -C "$tmp/x" ajiya || die "could not extract the archive"

  mkdir -p "$dir" || die "could not create $dir"
  # Copy beside the target then rename, so a running ajiya is replaced safely.
  cp "$tmp/x/ajiya" "$dir/.ajiya.new.$$" || die "could not write to $dir"
  chmod 755 "$dir/.ajiya.new.$$"
  mv -f "$dir/.ajiya.new.$$" "$dir/ajiya"

  say "Installed Ajiya $tag to $dir/ajiya"

  if ! on_path "$dir"; then
    say ""
    say "$dir is not on your PATH. Add it by putting this line in your shell profile"
    say "(for example ~/.zshrc or ~/.bashrc), then open a new terminal:"
    say ""
    say "  export PATH=\"$dir:\$PATH\""
  fi

  if have_terminal; then
    say ""
    printf 'Register Ajiya with the coding agents found on this machine (ajiya mcp install --all)? [y/N] '
    ans=""
    read -r ans </dev/tty || ans=""
    case "$ans" in
      y | Y | yes | YES | Yes) "$dir/ajiya" mcp install --all || err "ajiya mcp install failed; run it yourself later" ;;
      *) say "Skipped. Run 'ajiya mcp install --all' whenever you want to register." ;;
    esac
  else
    say ""
    say "To register Ajiya with your coding agents, run: ajiya mcp install --all"
  fi
}

main() {
  action=do_install
  for arg in "$@"; do
    case "$arg" in
      --uninstall) action=uninstall ;;
      -h | --help) usage; exit 0 ;;
      *) err "unknown option $arg"; usage >&2; exit 2 ;;
    esac
  done
  "$action"
}

main "$@"
