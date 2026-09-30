<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Agent-access

Goal: Agents read and change the plan through ajiya status and an MCP server they register with in one step

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| AJ-0042 | ajiya | ajiya status | Milestones in order, tickets in progress, what can start, Needs a human, check counts and the last five activity entries; --brief and --json; golden output for a fixture plan | AJ-0034, AJ-0038 | 🟩 Done · cbdcde9 · 2026-09-29 · tests passed |
| AJ-0043 | ajiya | ajiya mcp server | stdio MCP server in the same binary calling the CLI code, tools as listed in spec 02 with a test of the tools not offered; protocol tests for initialize, tools/list and tools/call; tools/list golden; parity with --json | AJ-0042 | 🟩 Done · 0852975 · 2026-09-30 · tests passed |
| AJ-0044 | ajiya | ajiya mcp install, uninstall and status | Registers ajiya mcp with Claude Code, Codex and Cursor from their current documentation, idempotent, backups, other entries kept, malformed files refused; init offers it; tests with a temporary home | AJ-0043 | 🟩 Done · 592adc8 · 2026-09-30 · tests passed |
| AJ-0094 | ajiya | Terminal detection ignores /dev/null | Prompts appear only on a real terminal: stdin from /dev/null or a pipe counts as non-interactive on macOS, Linux and Windows; init and mcp install then print the command instead of asking; tests cover each | AJ-0044 | 🟥 Pending |
| AJ-0095 | ajiya | mcp install honours CLAUDE_CONFIG_DIR | When CLAUDE_CONFIG_DIR is set, mcp install, uninstall and status use Claude Code's config there (per its current docs) instead of ~/.claude.json; tests cover set and unset | AJ-0044 | 🟥 Pending |
| AJ-0096 | ajiya | Registering says how to undo it | Whenever init or mcp install registers Ajiya with an agent, the output ends with the command that undoes it (ajiya mcp uninstall --<agent>, with --scope project when used); tests check the line | AJ-0044 | 🟥 Pending |
