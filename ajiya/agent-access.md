<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Agent-access

Goal: Agents read and change the plan through ajiya status and an MCP server they register with in one step

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| AJ-0042 | ajiya | ajiya status | Milestones in order, tickets in progress, what can start, Needs a human, check counts and the last five activity entries; --brief and --json; golden output for a fixture plan | AJ-0034, AJ-0038 | 🟩 Done · cbdcde9 · 2026-09-29 · tests passed |
| AJ-0043 | ajiya | ajiya mcp server | stdio MCP server in the same binary calling the CLI code, tools as listed in spec 02 with a test of the tools not offered; protocol tests for initialize, tools/list and tools/call; tools/list golden; parity with --json | AJ-0042 | 🟥 Pending |
| AJ-0044 | ajiya | ajiya mcp install, uninstall and status | Registers ajiya mcp with Claude Code, Codex and Cursor from their current documentation, idempotent, backups, other entries kept, malformed files refused; init offers it; tests with a temporary home | AJ-0043 | 🟥 Pending |
