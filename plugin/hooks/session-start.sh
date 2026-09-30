#!/bin/sh
# Ajiya session-start hook. Claude Code adds this script's stdout to the
# session's context.
#   no ajiya.toml in the project    -> print nothing
#   ajiya.toml but no ajiya binary  -> print the install line, nothing else
#   otherwise                       -> print 'ajiya status --brief'
# Every other failure is silent: a broken hook must never disturb a session.

# On Windows (Git Bash) session-start.ps1 does this job; exit so it does not run twice.
case "$(uname -s 2>/dev/null)" in MINGW* | MSYS* | CYGWIN*) exit 0 ;; esac

cd "${CLAUDE_PROJECT_DIR:-.}" 2>/dev/null || exit 0
[ -f ajiya.toml ] || exit 0

if ! command -v ajiya >/dev/null 2>&1; then
  echo "Ajiya is not installed. Install it with: go install github.com/AliyuYahaya/Ajiya/cmd/ajiya@latest (npm and Homebrew installs arrive with v0.2)"
  exit 0
fi

ajiya status --brief 2>/dev/null
exit 0
