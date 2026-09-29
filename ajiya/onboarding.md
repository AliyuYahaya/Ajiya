<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Onboarding

Goal: A new or existing project can be set up by an agent using only the installed kit and importers

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| AJ-0022 | ajiya | Setup guide | .ajiya/guide/setup.md covers planning from a goal and organising imported tickets, per spec section 6 | AJ-0021 | 🟩 Done · 8d5cafe · 2026-09-29 · tests passed |
| AJ-0023 | ajiya | Daily guide | .ajiya/guide/daily.md covers the loop from ajiya next to ajiya check, per spec section 6 | AJ-0021 | 🟩 Done · 8d5cafe · 2026-09-29 · tests passed |
| AJ-0024 | ajiya | Skill and agent instruction blocks | SKILL.md and the marked AGENTS.md and CLAUDE.md blocks point to the guides and list the rules | AJ-0022, AJ-0023 | 🟩 Done · 39cb092 · 2026-09-28 · tests passed |
| AJ-0025 | ajiya | init writes and updates the agent kit | Re-running init updates the kit and marked blocks without touching the user's text | AJ-0008, AJ-0024 | 🟩 Done · 15cad8a · 2026-09-29 · tests passed |
| AJ-0026 | ajiya | ajiya import todo | Checkbox and bullet items land in the inbox phase unchanged; ticked items are done before Ajiya | - | 🟩 Done · 5cf99f1 · 2026-09-29 · tests passed |
| AJ-0027 | ajiya | ajiya import github | Uses gh to import open issues as pending and closed issues as done, each keeping its URL | - | 🟩 Done · da34caf · 2026-09-29 · tests passed |
| AJ-0028 | ajiya | ajiya import legacy | Reads the collate.py format with new IDs, old IDs kept as aliases and a duplicate report; tested on reference/private if present | - | 🟩 Done · 3b606af · 2026-09-29 · tests passed |
| AJ-0029 | ajiya | Re-run init on this repo | This repository's kit is the one users get | AJ-0025 | 🟩 Done · c93b4be · 2026-09-29 · tests passed |
| AJ-0041 | ajiya | import legacy maps each RO package to a phase | Per spec 01 Part B: RO package to phase, area app per source, packages without scope, root-to-sink links, first package wins, cross-references dropped, inbox and archive, headings to milestones, area mode kept without a rollout file; fixture goldens for phases, ajiya.toml and report; check passes; package comment and the importing guide updated | AJ-0022, AJ-0038, AJ-0040 | 🟥 Pending |
