<!-- Written by ajiya build. Do not edit: run 'ajiya build'. -->
# Ajiya: progress

Generated 2026-09-30.

## Milestones

| Milestone | Closed | Required | % | Can start now |
|---|---|---|---|---|
| v0-1 | 37 | 37 | 100% | 0 |
| v0-2 | 64 | 74 | 86% | 5 |
| v0-3 | 64 | 84 | 76% | 5 |
| after-v0-3 | 64 | 83 | 77% | 5 |

## Phases

| Phase | Done | In progress | Pending | Total | % | Milestone |
|---|---|---|---|---|---|---|
| Commit-rule | 8 | 0 | 0 | 8 | 100% | v0-1 |
| Apps-launch | 14 | 0 | 0 | 14 | 100% | v0-1 |
| Outputs | 12 | 0 | 0 | 12 | 100% | v0-1 |
| Onboarding | 17 | 0 | 0 | 17 | 100% | v0-1 |
| Release | 5 | 0 | 0 | 5 | 100% | v0-1 |
| Agent-access | 5 | 0 | 1 | 6 | 83% | v0-2 |
| Install | 15 | 1 | 8 | 24 | 62% | v0-2 |
| Decisions-and-changes | 0 | 0 | 11 | 11 | 0% | v0-3 |

## Can start now

- **AJ-0073** install.ps1 for Windows (ajiya) · v0-2 · in progress
- **AJ-0079** Gated publish jobs (ajiya) · v0-2
- **AJ-0084** Sign and notarise the macOS binaries (ajiya) · v0-2
- **AJ-0096** Registering says how to undo it (ajiya) · v0-2
- **AJ-0097** Plugin leaves MCP registration to ajiya mcp install (ajiya) · v0-2

## Waiting

- **AJ-0047** v0.2 definition of done (ajiya) waits on AJ-0082, AJ-0096, AJ-0097
- **AJ-0048** Tags column (ajiya) waits on AJ-0047
- **AJ-0049** ajiya ticket files (ajiya) waits on AJ-0048
- **AJ-0050** Decision records and commands (ajiya) waits on AJ-0048
- **AJ-0051** Superseded status (ajiya) waits on AJ-0048
- **AJ-0052** Snapshots as git tags (ajiya) waits on AJ-0051
- **AJ-0053** change open and change impact (ajiya) waits on AJ-0049, AJ-0050, AJ-0051, AJ-0052
- **AJ-0054** change assess (ajiya) waits on AJ-0053
- **AJ-0055** change approve and change reject (ajiya) waits on AJ-0054
- **AJ-0056** Guides and skill for decisions and changes (ajiya) waits on AJ-0055
- **AJ-0057** Dashboard: open changes and decisions (ajiya) waits on AJ-0055
- **AJ-0058** v0.3 definition of done (ajiya) waits on AJ-0048, AJ-0049, AJ-0050, AJ-0051, AJ-0052, AJ-0053, AJ-0054, AJ-0055, AJ-0056
- **AJ-0074** Installer CI (ajiya) waits on AJ-0073
- **AJ-0078** Install scripts attached to every release (ajiya) waits on AJ-0073
- **AJ-0081** Release rehearsal on a release candidate (ajiya) waits on AJ-0074, AJ-0078, AJ-0079, AJ-0084
- **AJ-0082** Install docs for every channel (ajiya) waits on AJ-0081

## Needs a human

- **AJ-0047** v0.2 definition of done (ajiya): final sign-off before publishing v0.2
- **AJ-0058** v0.3 definition of done (ajiya): final sign-off before publishing v0.3
- **AJ-0081** Release rehearsal on a release candidate (ajiya): publishing anything needs the maintainer's approval

## Recent activity

- 2026-09-30 `634ca27` Resolve the agent-access plan conflict left in the last merge (AJ-0094, AJ-0095, AJ-0096)
  - AJ-0042: new → done
  - AJ-0043: new → done
  - AJ-0044: new → done
  - AJ-0094: new → done
  - AJ-0095: new → done
  - AJ-0096: new → pending
- 2026-09-30 `4b5a995` Merge branch 'worktree-agent-a4b3db9090249aeb4' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-30 `1894907` Installers offer ajiya mcp install for the detected agents (AJ-0045, AJ-0073)
- 2026-09-30 `e7aa392` Merge branch 'worktree-agent-a808e23b5b5a187e3' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-30 `d72350f` Mark AJ-0094 and AJ-0095 done (AJ-0094, AJ-0095)
  - AJ-0094: pending → done
  - AJ-0095: pending → done
- 2026-09-30 `c765d5a` Mark AJ-0045 done, AJ-0073 in progress (AJ-0045)
  - AJ-0045: pending → done
  - AJ-0073: pending → in_progress
- 2026-09-30 `a961c6f` Plan: the plugin leaves MCP registration to ajiya mcp install (chore)
  - AJ-0097: new → pending
- 2026-09-30 `7f92963` Add install.ps1 for Windows (AJ-0073)
- 2026-09-30 `6e41253` mcp install honours CLAUDE_CONFIG_DIR (AJ-0095)
- 2026-09-30 `3c274b9` Merge branch 'worktree-agent-af876fcf91ec1c89b' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-30 `262b8ce` Mark AJ-0077 done (AJ-0077)
  - AJ-0077: pending → done
- 2026-09-30 `0ae147d` Add install.sh for macOS and Linux (AJ-0045)
- 2026-09-30 `7e61f39` Merge branch 'worktree-agent-a3a05b52f4045e08c' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-30 `f805edf` Wire the Homebrew cask to the tap without publishing (AJ-0077)
- 2026-09-30 `cf289d1` Mark AJ-0092 and AJ-0093 done (AJ-0092, AJ-0093)
  - AJ-0092: pending → done
  - AJ-0093: pending → done
- and 192 more in data.js

## Checks

No problems found.
