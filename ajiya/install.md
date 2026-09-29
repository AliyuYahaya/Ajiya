<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Install

Goal: Ajiya installs in one step, with a Claude Code plugin

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| AJ-0045 | ajiya | Install script and npm package | Script verifies checksums and offers mcp install --all; npm package fetches the verified binary; Homebrew prints the register command; README install lines; nothing published without approval | AJ-0030, AJ-0031, AJ-0044 | 🟥 Pending |
| AJ-0046 | ajiya | Claude Code plugin | plugin/ with the skill, the MCP entry and a session-start hook running ajiya status --brief; validates against the documented schema; hook tested with and without ajiya on PATH | AJ-0042, AJ-0043 | 🟥 Pending |
| AJ-0047 | ajiya | v0.2 definition of done | On a clean account in under ten minutes from the README: install, register with Claude Code and Codex, plan a project and finish one ticket through MCP tools | AJ-0042, AJ-0043, AJ-0044, AJ-0045, AJ-0046 | 🟥 Pending · Needs a human: final sign-off before publishing v0.2 |
