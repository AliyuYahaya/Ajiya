<!-- Written by ajiya build. Do not edit: run 'ajiya build'. -->
# Ajiya: progress

Generated 2026-09-30.

## Milestones

| Milestone | Closed | Required | % | Can start now |
|---|---|---|---|---|
| v0-1 | 37 | 37 | 100% | 0 |
| v0-2 | 66 | 74 | 89% | 3 |
| v0-3 | 66 | 84 | 78% | 3 |
| after-v0-3 | 66 | 83 | 79% | 3 |

## Phases

| Phase | Done | In progress | Pending | Total | % | Milestone |
|---|---|---|---|---|---|---|
| Commit-rule | 8 | 0 | 0 | 8 | 100% | v0-1 |
| Apps-launch | 14 | 0 | 0 | 14 | 100% | v0-1 |
| Outputs | 12 | 0 | 0 | 12 | 100% | v0-1 |
| Onboarding | 17 | 0 | 0 | 17 | 100% | v0-1 |
| Release | 5 | 0 | 0 | 5 | 100% | v0-1 |
| Agent-access | 6 | 0 | 0 | 6 | 100% | v0-2 |
| Install | 16 | 4 | 4 | 24 | 66% | v0-2 |
| Decisions-and-changes | 0 | 0 | 11 | 11 | 0% | v0-3 |

## Can start now

- **AJ-0073** install.ps1 for Windows (ajiya) · v0-2 · in progress
- **AJ-0079** Gated publish jobs (ajiya) · v0-2 · in progress
- **AJ-0084** Sign and notarise the macOS binaries (ajiya) · v0-2 · in progress

## Waiting

- **AJ-0047** v0.2 definition of done (ajiya) waits on AJ-0082
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

- 2026-09-30 `d83da99` Merge branch 'worktree-agent-a920ee3d0cb59f8ae' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-30 `5f3e837` Mark AJ-0096 and AJ-0097 done (AJ-0096, AJ-0097)
  - AJ-0096: in_progress → done
  - AJ-0097: in_progress → done
- 2026-09-30 `c5c8987` Plugin leaves MCP registration to ajiya mcp install (AJ-0097)
  - AJ-0097: pending → in_progress
- 2026-09-30 `9e15dbf` Merge branch 'worktree-agent-ac065ef1f3558d188' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-30 `e6cbb0d` Sign and notarise the darwin binaries; drop the cask quarantine hook (AJ-0084)
  - AJ-0084: pending → in_progress
- 2026-09-30 `f582d50` Gated npm and tap publish jobs in the release workflow (AJ-0079)
  - AJ-0079: pending → in_progress
- 2026-09-30 `7ab4260` Installer CI: setup-python v7, the current major (AJ-0074)
- 2026-09-30 `74a4e46` Merge branch 'worktree-agent-a8cf3a95133ce824d' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-30 `926cdc8` Registering with an agent ends with the command that undoes it (AJ-0096)
  - AJ-0096: pending → in_progress
- 2026-09-30 `72ed8c7` Start AJ-0074 and rebuild plan outputs (AJ-0074)
  - AJ-0074: pending → in_progress
- 2026-09-30 `8bbf335` install.ps1: restore the caller's preferences and use an approved verb (AJ-0073)
- 2026-09-30 `fb07b1e` Add the installer CI job and harden test-install.ps1 (AJ-0074)
- 2026-09-30 `634ca27` Resolve the agent-access plan conflict left in the last merge (AJ-0094, AJ-0095, AJ-0096)
  - AJ-0042: new → done
  - AJ-0043: new → done
  - AJ-0044: new → done
  - AJ-0094: new → done
  - AJ-0095: new → done
  - AJ-0096: new → pending
- 2026-09-30 `4b5a995` Merge branch 'worktree-agent-a4b3db9090249aeb4' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-30 `1894907` Installers offer ajiya mcp install for the detected agents (AJ-0045, AJ-0073)
- and 204 more in data.js

## Checks

No problems found.
