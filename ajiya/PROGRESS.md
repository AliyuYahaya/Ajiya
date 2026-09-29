<!-- Written by ajiya build. Do not edit: run 'ajiya build'. -->
# Ajiya: progress

Generated 2026-09-29.

## Milestones

| Milestone | Closed | Required | % | Can start now |
|---|---|---|---|---|
| v0-1 | 36 | 37 | 97% | 1 |
| v0-2 | 37 | 61 | 60% | 1 |
| v0-3 | 37 | 71 | 52% | 1 |
| after-v0-3 | 37 | 70 | 52% | 1 |

## Phases

| Phase | Done | In progress | Pending | Total | % | Milestone |
|---|---|---|---|---|---|---|
| Commit-rule | 8 | 0 | 0 | 8 | 100% | v0-1 |
| Apps-launch | 14 | 0 | 0 | 14 | 100% | v0-1 |
| Outputs | 12 | 0 | 0 | 12 | 100% | v0-1 |
| Onboarding | 10 | 0 | 0 | 10 | 100% | v0-1 |
| Release | 4 | 0 | 1 | 5 | 80% | v0-1 |
| Agent-access | 0 | 0 | 3 | 3 | 0% | v0-2 |
| Install | 1 | 0 | 20 | 21 | 4% | v0-2 |
| Decisions-and-changes | 0 | 0 | 11 | 11 | 0% | v0-3 |

## Can start now

- **AJ-0034** v0.1 definition of done (ajiya) · v0-1

## Waiting

- **AJ-0042** ajiya status (ajiya) waits on AJ-0034
- **AJ-0043** ajiya mcp server (ajiya) waits on AJ-0042
- **AJ-0044** ajiya mcp install, uninstall and status (ajiya) waits on AJ-0043
- **AJ-0045** install.sh for macOS and Linux (ajiya) waits on AJ-0044
- **AJ-0046** Claude Code plugin (ajiya) waits on AJ-0042, AJ-0043
- **AJ-0047** v0.2 definition of done (ajiya) waits on AJ-0042, AJ-0043, AJ-0044, AJ-0045, AJ-0046, AJ-0082
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
- **AJ-0068** npm account and package names (ajiya) waits on AJ-0034
- **AJ-0069** Homebrew tap repository and token (ajiya) waits on AJ-0034
- **AJ-0070** Choose the install script URL (ajiya) waits on AJ-0034
- **AJ-0071** Release environments with required reviewers (ajiya) waits on AJ-0034
- **AJ-0072** Release workflow: build to a draft release (ajiya) waits on AJ-0034
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

## Needs a human

- **AJ-0034** v0.1 definition of done (ajiya): final sign-off before publishing
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

- 2026-09-29 `7748eae` Merge branch 'worktree-agent-a658cbd8dac8f5d6c' (-)
- 2026-09-29 `d0c5d61` AJ-0066 done (AJ-0066)
  - AJ-0066: in_progress → done
- 2026-09-29 `5a75b38` Guides: how to set the test command (AJ-0066)
  - AJ-0066: pending → in_progress
- 2026-09-29 `60ba1a3` Plan: guides explain the test command (chore)
  - AJ-0066: new → pending
- 2026-09-29 `c6f6b5a` AJ-0033 done (AJ-0033)
  - AJ-0033: in_progress → done
- 2026-09-29 `6fb1bcc` README: how it works, install, and the two walkthroughs (AJ-0033)
  - AJ-0033: pending → in_progress
- 2026-09-29 `edf9e22` AJ-0031 done (AJ-0031)
  - AJ-0031: in_progress → done
- 2026-09-29 `bb6e0eb` Homebrew cask config (AJ-0031)
  - AJ-0031: pending → in_progress
- 2026-09-29 `03b432d` AJ-0030 done (AJ-0030)
  - AJ-0030: in_progress → done
- 2026-09-29 `9a2c69c` GoReleaser config (AJ-0030)
  - AJ-0030: pending → in_progress
- 2026-09-29 `d18ad9e` AJ-0029 done (AJ-0029)
  - AJ-0029: in_progress → done
- 2026-09-29 `c93b4be` This repository runs the kit it ships (AJ-0029)
  - AJ-0029: pending → in_progress
- 2026-09-29 `298f56f` AJ-0025 done (AJ-0025)
  - AJ-0025: in_progress → done
- 2026-09-29 `15cad8a` init writes and updates the agent kit (AJ-0025)
  - AJ-0025: pending → in_progress
- 2026-09-29 `4f9b1a6` Plan: ticket done evidence must come after the ticket started (chore)
  - AJ-0065: new → pending
- and 108 more in data.js

## Checks

No problems found.
