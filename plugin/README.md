# Ajiya plugin for Claude Code

Bundles two things for Claude Code:

- the **Ajiya skill** (`skills/ajiya/SKILL.md`, a copy of the kit's skill);
- a **session-start hook** (`hooks/session-start.sh` on macOS and Linux, `hooks/session-start.ps1` on
  Windows) that runs
  `ajiya status --brief` in projects that have `ajiya.toml`, so each session
  starts with the current state. In other projects it prints nothing.

The plugin does **not** declare the MCP server. `ajiya mcp install` registers it
with Claude Code (and Codex or Cursor), so a plugin entry as well would list
Ajiya's tools twice. When the project has `ajiya.toml` but Ajiya is not registered
with Claude Code, the hook adds one line after the status saying to run
`ajiya mcp install --claude`. It asks the binary, `ajiya mcp status --claude
--quiet`, which prints nothing and exits 0 when registered (user or project
scope), 1 when not; any other answer (an older `ajiya`, an unreadable config)
adds no line.

The plugin does not ship the `ajiya` binary. Install it first:

    npm install -g @ajiya/cli

(or the install script or Homebrew; see the main README.) If `ajiya` is missing in a project
that has `ajiya.toml`, the hook prints this install line and nothing else.

## Install

The repository root holds a marketplace file (`.claude-plugin/marketplace.json`).
Inside Claude Code:

    /plugin marketplace add AliyuYahaya/Ajiya
    /plugin install ajiya@ajiya

Or from a shell:

    claude plugin marketplace add AliyuYahaya/Ajiya
    claude plugin install ajiya@ajiya

To try the plugin from a checkout without installing it:

    claude --plugin-dir ./plugin

## Notes

- On Windows the hook runs through `powershell.exe` and does not need Git Bash.
  `hooks/hooks.json` declares two shell-form hooks, one for `sh` and one for
  PowerShell. Each ends in `; exit 0`, and each script exits silently on the
  wrong platform, so only one of them prints. Claude Code has no OS-specific hook
  field; see https://code.claude.com/docs/en/hooks (command hook fields) and
  https://code.claude.com/docs/en/plugins-reference (hooks).
- After the kit's `SKILL.md` changes, refresh the copy with
  `cp internal/kit/files/SKILL.md plugin/skills/ajiya/SKILL.md`; a test fails
  until you do.
- Check the plugin with `claude plugin validate ./plugin` and the marketplace
  with `claude plugin validate .`.
