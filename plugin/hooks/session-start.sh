#!/bin/sh
# Ajiya session-start hook. Claude Code adds this script's stdout to the
# session's context.
#   no ajiya.toml in the project    -> print nothing
#   ajiya.toml but no ajiya binary  -> print the install line, nothing else
#   otherwise                       -> print 'ajiya status --brief', then one line
#                                      saying to run 'ajiya mcp install' when Ajiya
#                                      is not registered with Claude Code (asked of
#                                      the binary: 'ajiya mcp status --claude --quiet'
#                                      exits 1 for "not registered"; any other
#                                      answer, including an older ajiya, adds nothing)
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
ajiya mcp status --claude --quiet >/dev/null 2>&1
if [ "$?" = 1 ]; then
  echo "Ajiya is not registered with Claude Code, so a session cannot use its tools. Register it with: ajiya mcp install --claude"
fi
exit 0
