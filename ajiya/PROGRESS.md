<!-- Written by ajiya build. Do not edit: run 'ajiya build'. -->
# Ajiya: progress

Generated 2026-09-29.

## Milestones

| Milestone | Closed | Required | % | Can start now |
|---|---|---|---|---|
| v0-1 | 37 | 37 | 100% | 0 |
| v0-2 | 38 | 68 | 55% | 12 |
| v0-3 | 38 | 78 | 48% | 12 |
| after-v0-3 | 38 | 77 | 49% | 12 |

## Phases

| Phase | Done | In progress | Pending | Total | % | Milestone |
|---|---|---|---|---|---|---|
| Commit-rule | 8 | 0 | 0 | 8 | 100% | v0-1 |
| Apps-launch | 14 | 0 | 0 | 14 | 100% | v0-1 |
| Outputs | 12 | 0 | 0 | 12 | 100% | v0-1 |
| Onboarding | 10 | 0 | 7 | 17 | 58% | v0-1 |
| Release | 5 | 0 | 0 | 5 | 100% | v0-1 |
| Agent-access | 0 | 0 | 3 | 3 | 0% | v0-2 |
| Install | 1 | 0 | 20 | 21 | 4% | v0-2 |
| Decisions-and-changes | 0 | 0 | 11 | 11 | 0% | v0-3 |

## Can start now

- **AJ-0042** ajiya status (ajiya) · v0-2
- **AJ-0068** npm account and package names (ajiya) · v0-2
- **AJ-0069** Homebrew tap repository and token (ajiya) · v0-2
- **AJ-0071** Release environments with required reviewers (ajiya) · v0-2
- **AJ-0072** Release workflow: build to a draft release (ajiya) · v0-2
- **AJ-0070** Choose the install script URL (ajiya) · v0-2
- **AJ-0085** Imports use the project's app (ajiya) · v0-2
- **AJ-0086** ajiya phase remove for an empty phase (ajiya) · v0-2
- **AJ-0088** ajiya ticket list (ajiya) · v0-2
- **AJ-0087** Group help: ajiya ticket --help and ajiya phase --help (ajiya) · v0-2
- **AJ-0089** ticket edit always prints the ticket (ajiya) · v0-2
- **AJ-0090** init suggests the test command (ajiya) · v0-2

## Waiting

- **AJ-0043** ajiya mcp server (ajiya) waits on AJ-0042
- **AJ-0044** ajiya mcp install, uninstall and status (ajiya) waits on AJ-0043
- **AJ-0045** install.sh for macOS and Linux (ajiya) waits on AJ-0044
- **AJ-0046** Claude Code plugin (ajiya) waits on AJ-0042, AJ-0043
- **AJ-0047** v0.2 definition of done (ajiya) waits on AJ-0042, AJ-0043, AJ-0044, AJ-0045, AJ-0046, AJ-0082, AJ-0085, AJ-0086, AJ-0087, AJ-0088, AJ-0089, AJ-0090, AJ-0091
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
- **AJ-0075** npm packages: a wrapper and one package per platform (ajiya) waits on AJ-0068
- **AJ-0076** npm launcher tests (ajiya) waits on AJ-0075
- **AJ-0077** Homebrew tap wiring (ajiya) waits on AJ-0044, AJ-0069
- **AJ-0078** Install scripts served at a stable URL (ajiya) waits on AJ-0045, AJ-0070, AJ-0073
- **AJ-0079** Gated publish jobs (ajiya) waits on AJ-0071, AJ-0072, AJ-0075, AJ-0077
- **AJ-0080** npm trusted publishing (ajiya) waits on AJ-0075
- **AJ-0081** Release rehearsal on a release candidate (ajiya) waits on AJ-0074, AJ-0076, AJ-0078, AJ-0079, AJ-0080, AJ-0084
- **AJ-0082** Install docs for every channel (ajiya) waits on AJ-0081
- **AJ-0083** Apple signing credentials as release secrets (ajiya) waits on AJ-0071
- **AJ-0084** Sign and notarise the macOS binaries (ajiya) waits on AJ-0072, AJ-0077, AJ-0083
- **AJ-0091** Guide fixes from the walkthrough test (ajiya) waits on AJ-0085, AJ-0086, AJ-0088

## Needs a human

- **AJ-0047** v0.2 definition of done (ajiya): final sign-off before publishing v0.2
- **AJ-0058** v0.3 definition of done (ajiya): final sign-off before publishing v0.3
- **AJ-0068** npm account and package names (ajiya): needs the maintainer's npm account
- **AJ-0069** Homebrew tap repository and token (ajiya): needs the maintainer's GitHub account
- **AJ-0070** Choose the install script URL (ajiya): a domain costs money and is the maintainer's choice
- **AJ-0071** Release environments with required reviewers (ajiya): repository settings need the maintainer
- **AJ-0080** npm trusted publishing (ajiya): needs the maintainer's npm account
- **AJ-0081** Release rehearsal on a release candidate (ajiya): publishing anything needs the maintainer's approval
- **AJ-0083** Apple signing credentials as release secrets (ajiya): needs the maintainer's Apple Developer account

## Recent activity

- 2026-09-29 `cc18043` npm names: ajiya plus @ajiya/<os>-<arch> under an ajiya org (AJ-0068, AJ-0075)
- 2026-09-29 `6a75e60` v0.1 signed off (AJ-0034)
  - AJ-0034: pending → done
- 2026-09-29 `e803fc9` Merge branch 'worktree-agent-a658cbd8dac8f5d6c' (-)
- 2026-09-29 `e0873df` Licence held by Yavid PTY Ltd; the name and logo are not licensed (chore)
- 2026-09-29 `2603a19` MIT licence (chore)
- 2026-09-29 `110dcb2` Rebuild plan outputs after the merge (chore)
- 2026-09-29 `cc33bdb` Merge branch 'worktree-agent-a5a988be55680e087' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-29 `2d65774` Plan: AJ-0041 done (AJ-0041)
  - AJ-0041: pending → done
- 2026-09-29 `7d547a5` import legacy: map each RO package to a phase (AJ-0041)
- 2026-09-29 `ff9cdca` Plan: sign and notarise the macOS binaries (chore)
  - AJ-0083: new → pending
  - AJ-0084: new → pending
- 2026-09-29 `810c3c0` npm packaging decided: per-platform packages behind a wrapper (AJ-0067, AJ-0075)
  - AJ-0067: pending → done
- 2026-09-29 `2ba31d5` Reflow the daily guide's done paragraph (AJ-0065)
- 2026-09-29 `622e678` Merge branch 'worktree-agent-af0c9a6c1037d2858' into worktree-agent-a658cbd8dac8f5d6c (-)
- 2026-09-29 `140a432` AJ-0065 done (AJ-0065)
  - AJ-0065: pending → done
- 2026-09-29 `5c1af99` Ticket done counts only commits made after the ticket started (AJ-0065)
- and 124 more in data.js

## Checks

No problems found.
