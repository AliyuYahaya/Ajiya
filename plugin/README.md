# Ajiya plugin for Claude Code

Bundles three things for Claude Code:

- the **Ajiya skill** (`skills/ajiya/SKILL.md`, a copy of the kit's skill);
- the **MCP server** (`ajiya mcp`, declared in `.mcp.json`);
- a **session-start hook** (`hooks/session-start.sh` on macOS and Linux, `hooks/session-start.ps1` on
  Windows) that runs
  `ajiya status --brief` in projects that have `ajiya.toml`, so each session
  starts with the current state. In other projects it prints nothing.

The plugin does not ship the `ajiya` binary. Install it first:

    go install github.com/AliyuYahaya/Ajiya/cmd/ajiya@latest

(npm and Homebrew installs come with v0.2.) If `ajiya` is missing in a project
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
