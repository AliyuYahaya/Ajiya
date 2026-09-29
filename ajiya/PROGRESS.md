<!-- Written by ajiya build. Do not edit: run 'ajiya build'. -->
# Ajiya: progress

Generated 2026-09-29.

## Milestones

| Milestone | Closed | Required | % | Can start now |
|---|---|---|---|---|
| v0-1 | 37 | 37 | 100% | 0 |
| v0-2 | 49 | 68 | 72% | 4 |
| v0-3 | 49 | 78 | 62% | 4 |
| after-v0-3 | 49 | 77 | 63% | 4 |

## Phases

| Phase | Done | In progress | Pending | Total | % | Milestone |
|---|---|---|---|---|---|---|
| Commit-rule | 8 | 0 | 0 | 8 | 100% | v0-1 |
| Apps-launch | 14 | 0 | 0 | 14 | 100% | v0-1 |
| Outputs | 12 | 0 | 0 | 12 | 100% | v0-1 |
| Onboarding | 16 | 0 | 1 | 17 | 94% | v0-1 |
| Release | 5 | 0 | 0 | 5 | 100% | v0-1 |
| Agent-access | 0 | 0 | 3 | 3 | 0% | v0-2 |
| Install | 6 | 0 | 15 | 21 | 28% | v0-2 |
| Decisions-and-changes | 0 | 0 | 11 | 11 | 0% | v0-3 |

## Can start now

- **AJ-0042** ajiya status (ajiya) · v0-2
- **AJ-0075** npm packages: @ajiya/cli and one package per platform (ajiya) · v0-2
- **AJ-0083** Apple signing credentials as release secrets (ajiya) · v0-2
- **AJ-0091** Guide fixes from the walkthrough test (ajiya) · v0-2

## Waiting

- **AJ-0043** ajiya mcp server (ajiya) waits on AJ-0042
- **AJ-0044** ajiya mcp install, uninstall and status (ajiya) waits on AJ-0043
- **AJ-0045** install.sh for macOS and Linux (ajiya) waits on AJ-0044
- **AJ-0046** Claude Code plugin (ajiya) waits on AJ-0042, AJ-0043
- **AJ-0047** v0.2 definition of done (ajiya) waits on AJ-0042, AJ-0043, AJ-0044, AJ-0045, AJ-0046, AJ-0082, AJ-0091
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
- **AJ-0073** install.ps1 for Windows (ajiya) waits on AJ-0044
- **AJ-0074** Installer CI (ajiya) waits on AJ-0045, AJ-0073
- **AJ-0076** npm launcher tests (ajiya) waits on AJ-0075
- **AJ-0077** Homebrew tap wiring (ajiya) waits on AJ-0044
- **AJ-0078** Install scripts attached to every release (ajiya) waits on AJ-0045, AJ-0073
- **AJ-0079** Gated publish jobs (ajiya) waits on AJ-0075, AJ-0077
- **AJ-0080** npm trusted publishing (ajiya) waits on AJ-0075
- **AJ-0081** Release rehearsal on a release candidate (ajiya) waits on AJ-0074, AJ-0076, AJ-0078, AJ-0079, AJ-0080, AJ-0084
- **AJ-0082** Install docs for every channel (ajiya) waits on AJ-0081
- **AJ-0084** Sign and notarise the macOS binaries (ajiya) waits on AJ-0077, AJ-0083

## Needs a human

- **AJ-0047** v0.2 definition of done (ajiya): final sign-off before publishing v0.2
- **AJ-0058** v0.3 definition of done (ajiya): final sign-off before publishing v0.3
- **AJ-0080** npm trusted publishing (ajiya): needs the maintainer's npm account
- **AJ-0081** Release rehearsal on a release candidate (ajiya): publishing anything needs the maintainer's approval
- **AJ-0083** Apple signing credentials as release secrets (ajiya): needs the maintainer's Apple Developer account

## Recent activity

- 2026-09-29 `4a25d7c` Merge branch 'worktree-agent-afca5fbf28b0fec73' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-29 `f38b713` AJ-0072 done (AJ-0072)
  - AJ-0072: pending → done
- 2026-09-29 `d1df3f0` Release workflow: tag builds a draft release, pull requests a snapshot (AJ-0072)
- 2026-09-29 `7363ad4` Install scripts are served as GitHub release assets (AJ-0070, AJ-0078)
  - AJ-0070: pending → done
- 2026-09-29 `ab57528` npm account, org and @ajiya/cli placeholder in place (AJ-0068)
  - AJ-0068: pending → done
- 2026-09-29 `2680bc1` phase remove updates [phases] order before deleting the file (AJ-0086)
- 2026-09-29 `184720e` Merge branch 'worktree-agent-ac75a90c83d346edb' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-29 `607c7b6` Merge branch 'worktree-agent-af1c1cc18e0e82a16' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-29 `8a67799` Record AJ-0085, AJ-0086 and AJ-0090 as done (AJ-0085, AJ-0086, AJ-0090)
  - AJ-0085: in_progress → done
  - AJ-0086: in_progress → done
  - AJ-0090: in_progress → done
- 2026-09-29 `23f1de8` ajiya init suggests the test command (AJ-0090)
  - AJ-0090: pending → in_progress
- 2026-09-29 `47eac13` npm wrapper is @ajiya/cli (AJ-0068, AJ-0075)
- 2026-09-29 `4e23e52` Add ajiya phase remove for an empty phase (AJ-0086)
  - AJ-0086: pending → in_progress
- 2026-09-29 `45d9f00` Mark AJ-0087, AJ-0088 and AJ-0089 done (AJ-0087, AJ-0088, AJ-0089)
  - AJ-0087: pending → done
  - AJ-0088: pending → done
  - AJ-0089: pending → done
- 2026-09-29 `ac6842b` Add ticket list, group help and required-flag usage lines (AJ-0087, AJ-0088, AJ-0089)
- 2026-09-29 `7f1ede7` Homebrew tap and its token in place (AJ-0069)
  - AJ-0069: in_progress → done
- and 142 more in data.js

## Checks

No problems found.
