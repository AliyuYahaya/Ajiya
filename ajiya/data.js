// Written by ajiya build. Do not edit: run 'ajiya build'.
window.AJIYA = {
  "schema": 1,
  "generated": "2026-09-29",
  "project": {
    "name": "Ajiya",
    "prefix": "AJ"
  },
  "milestones": [
    {
      "name": "v0-1",
      "targets": [
        "AJ-0034"
      ],
      "required": {
        "total": 37,
        "done": 37,
        "in_progress": 0,
        "pending": 0,
        "dropped": 0
      },
      "percent": 100,
      "ready": 0
    },
    {
      "name": "v0-2",
      "targets": [
        "AJ-0047"
      ],
      "required": {
        "total": 61,
        "done": 38,
        "in_progress": 0,
        "pending": 23,
        "dropped": 0
      },
      "percent": 62,
      "ready": 6
    },
    {
      "name": "v0-3",
      "targets": [
        "AJ-0058"
      ],
      "required": {
        "total": 71,
        "done": 38,
        "in_progress": 0,
        "pending": 33,
        "dropped": 0
      },
      "percent": 53,
      "ready": 6
    },
    {
      "name": "after-v0-3",
      "targets": [
        "AJ-0057"
      ],
      "required": {
        "total": 70,
        "done": 38,
        "in_progress": 0,
        "pending": 32,
        "dropped": 0
      },
      "percent": 54,
      "ready": 6
    }
  ],
  "phases": [
    {
      "slug": "commit-rule",
      "title": "Commit-rule",
      "goal": "Every commit names 1 to 3 tickets, enforced by hooks with a CI check as backup",
      "counts": {
        "total": 8,
        "done": 8,
        "in_progress": 0,
        "pending": 0,
        "dropped": 0
      },
      "percent": 100,
      "milestone": "v0-1"
    },
    {
      "slug": "apps-launch",
      "title": "Apps-launch",
      "goal": "Apps are detected and managed, launch readiness is defined, and every check in the spec runs",
      "counts": {
        "total": 14,
        "done": 14,
        "in_progress": 0,
        "pending": 0,
        "dropped": 0
      },
      "percent": 100,
      "milestone": "v0-1"
    },
    {
      "slug": "outputs",
      "title": "Outputs",
      "goal": "PROGRESS.md, data.js, the dashboard and serve show the plan and its activity",
      "counts": {
        "total": 12,
        "done": 12,
        "in_progress": 0,
        "pending": 0,
        "dropped": 0
      },
      "percent": 100,
      "milestone": "v0-1"
    },
    {
      "slug": "onboarding",
      "title": "Onboarding",
      "goal": "A new or existing project can be set up by an agent using only the installed kit and importers",
      "counts": {
        "total": 10,
        "done": 10,
        "in_progress": 0,
        "pending": 0,
        "dropped": 0
      },
      "percent": 100,
      "milestone": "v0-1"
    },
    {
      "slug": "release",
      "title": "Release",
      "goal": "v0.1 is ready to publish: binaries, Homebrew cask, README walkthroughs and changelog",
      "counts": {
        "total": 5,
        "done": 5,
        "in_progress": 0,
        "pending": 0,
        "dropped": 0
      },
      "percent": 100,
      "milestone": "v0-1"
    },
    {
      "slug": "agent-access",
      "title": "Agent-access",
      "goal": "Agents read and change the plan through ajiya status and an MCP server they register with in one step",
      "counts": {
        "total": 3,
        "done": 0,
        "in_progress": 0,
        "pending": 3,
        "dropped": 0
      },
      "percent": 0,
      "milestone": "v0-2"
    },
    {
      "slug": "install",
      "title": "Install",
      "goal": "Ajiya installs in one step, with a Claude Code plugin",
      "counts": {
        "total": 21,
        "done": 1,
        "in_progress": 0,
        "pending": 20,
        "dropped": 0
      },
      "percent": 4,
      "milestone": "v0-2"
    },
    {
      "slug": "decisions-and-changes",
      "title": "Decisions-and-changes",
      "goal": "Plans change without losing history: tags, decisions, superseded work and approved change requests",
      "counts": {
        "total": 11,
        "done": 0,
        "in_progress": 0,
        "pending": 11,
        "dropped": 0
      },
      "percent": 0,
      "milestone": "v0-3"
    }
  ],
  "apps": [
    {
      "name": "ajiya",
      "path": ".",
      "counts": {
        "total": 84,
        "done": 50,
        "in_progress": 0,
        "pending": 34,
        "dropped": 0
      }
    }
  ],
  "tickets": [
    {
      "id": "AJ-0001",
      "phase": "commit-rule",
      "app": "ajiya",
      "title": "Parse Ajiya trailers with git's own rules",
      "done_when": "Refs, chore, merge and revert exemptions agree between message files and history; table tests pass",
      "depends": [],
      "status": {
        "state": "done",
        "text": "🟩 Done · 4cc2b3f · 2026-09-28",
        "commit": "4cc2b3f",
        "date": "2026-09-28"
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0002",
        "AJ-0003",
        "AJ-0004",
        "AJ-0007",
        "AJ-0013",
        "AJ-0017",
        "AJ-0032"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0002",
      "phase": "commit-rule",
      "app": "ajiya",
      "title": "ajiya check --commits \u003crange\u003e",
      "done_when": "Reports commits with 0 or more than 3 references and unknown IDs; fixture per case",
      "depends": [
        "AJ-0001"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 4a49882 · 2026-09-28",
        "commit": "4a49882",
        "date": "2026-09-28"
      },
      "milestone": "",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0006"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0003",
      "phase": "commit-rule",
      "app": "ajiya",
      "title": "commit-msg hook",
      "done_when": "Refuses a commit without a valid trailer; merges and reverts pass; script test in a temp repo",
      "depends": [
        "AJ-0001"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · ea005bf · 2026-09-28",
        "commit": "ea005bf",
        "date": "2026-09-28"
      },
      "milestone": "",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0005"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0004",
      "phase": "commit-rule",
      "app": "ajiya",
      "title": "prepare-commit-msg hook",
      "done_when": "Pre-fills the trailer from in-progress tickets and changed ajiya/ rows; leaves merge, squash and amend alone",
      "depends": [
        "AJ-0001"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 6f015d8 · 2026-09-28",
        "commit": "6f015d8",
        "date": "2026-09-28"
      },
      "milestone": "",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0005"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0005",
      "phase": "commit-rule",
      "app": "ajiya",
      "title": "ajiya hook install",
      "done_when": "Respects core.hooksPath, appends a marked block to existing hooks, works in Git for Windows, clear message when ajiya is not on PATH",
      "depends": [
        "AJ-0003",
        "AJ-0004"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 58e9736 · 2026-09-28",
        "commit": "58e9736",
        "date": "2026-09-28"
      },
      "milestone": "",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0006"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0006",
      "phase": "commit-rule",
      "app": "ajiya",
      "title": "CI commit check and hooks on this repo",
      "done_when": "CI runs ajiya check --commits; hooks installed here and a commit without a trailer is refused",
      "depends": [
        "AJ-0002",
        "AJ-0005"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 5c94eca · 2026-09-29 · Note: hooks installed here; unknown-ID commit refused",
        "note": "hooks installed here; unknown-ID commit refused",
        "commit": "5c94eca",
        "date": "2026-09-29"
      },
      "milestone": "",
      "ready": false,
      "waiting_on": [],
      "dependants": [],
      "unblocks": 0
    },
    {
      "id": "AJ-0007",
      "phase": "commit-rule",
      "app": "ajiya",
      "title": "ticket show lists referencing commits",
      "done_when": "ajiya ticket show prints hash, date and subject of each commit naming the ticket",
      "depends": [
        "AJ-0001"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 2f9d28a · 2026-09-29",
        "commit": "2f9d28a",
        "date": "2026-09-29"
      },
      "milestone": "",
      "ready": false,
      "waiting_on": [],
      "dependants": [],
      "unblocks": 0
    },
    {
      "id": "AJ-0008",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "App detection in init",
      "done_when": "Suggests apps from workspace files and manifests, ranked, skipping vendor folders; --yes accepts; fixture repos pass",
      "depends": [],
      "status": {
        "state": "done",
        "text": "🟩 Done · 83f20ca · 2026-09-29 · tests passed",
        "commit": "83f20ca",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0025"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0009",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "ajiya app add, remove and move",
      "done_when": "Refuses to orphan open tickets unless --to is given; script tests pass",
      "depends": [],
      "status": {
        "state": "done",
        "text": "🟩 Done · 57e4454 · 2026-09-29",
        "commit": "57e4454",
        "date": "2026-09-29"
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0012"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0010",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "ajiya launch set and show",
      "done_when": "Required set is the target's tickets plus everything they depend on; table tests pass",
      "depends": [],
      "status": {
        "state": "done",
        "text": "🟩 Done · 3349f60 · 2026-09-29",
        "commit": "3349f60",
        "date": "2026-09-29"
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0011",
        "AJ-0012",
        "AJ-0037"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0011",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "ajiya next --launch",
      "done_when": "Launch-required tickets are listed first",
      "depends": [
        "AJ-0010"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · d32c066 · 2026-09-29",
        "commit": "d32c066",
        "date": "2026-09-29"
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0038"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0012",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "Remaining error checks",
      "done_when": "Unregistered app, done without evidence and missing launch target each have a code and a fixture",
      "depends": [
        "AJ-0009",
        "AJ-0010"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 7bac540 · 2026-09-29 · tests passed",
        "commit": "7bac540",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0013"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0013",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "Warning checks and --strict",
      "done_when": "Every warning in spec section 4 has a W code and a fixture; --strict fails on warnings",
      "depends": [
        "AJ-0001",
        "AJ-0012"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · b3d5792 · 2026-09-29 · tests passed",
        "commit": "b3d5792",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0014",
        "AJ-0018"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0014",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "--json on every read command",
      "done_when": "Each read command prints stable JSON with --json; tests decode it",
      "depends": [
        "AJ-0013"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · e2b8505 · 2026-09-29 · tests passed",
        "commit": "e2b8505",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "",
      "ready": false,
      "waiting_on": [],
      "dependants": [],
      "unblocks": 0
    },
    {
      "id": "AJ-0015",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "ajiya ticket done --test",
      "done_when": "Runs the configured test command, records tests passed, refuses on failure",
      "depends": [],
      "status": {
        "state": "done",
        "text": "🟩 Done · c0239ef · 2026-09-29 · tests passed",
        "commit": "c0239ef",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "",
      "ready": false,
      "waiting_on": [],
      "dependants": [],
      "unblocks": 0
    },
    {
      "id": "AJ-0016",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "ajiya ticket block and ticket drop",
      "done_when": "block sets a reason; drop needs --reason and --by and counts as closed",
      "depends": [],
      "status": {
        "state": "done",
        "text": "🟩 Done · 1ad6b45 · 2026-09-29",
        "commit": "1ad6b45",
        "date": "2026-09-29"
      },
      "milestone": "",
      "ready": false,
      "waiting_on": [],
      "dependants": [],
      "unblocks": 0
    },
    {
      "id": "AJ-0017",
      "phase": "outputs",
      "app": "ajiya",
      "title": "Activity feed from git",
      "done_when": "Last 30 days of commits with tickets, status changes from diffs and agent co-authors; deterministic",
      "depends": [
        "AJ-0001"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 8c2dea8 · 2026-09-29 · tests passed",
        "commit": "8c2dea8",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0018"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0018",
      "phase": "outputs",
      "app": "ajiya",
      "title": "ajiya build",
      "done_when": "Writes PROGRESS.md and data.js with a schema version, only when content changes, dated by the newest source change",
      "depends": [
        "AJ-0013",
        "AJ-0017",
        "AJ-0038",
        "AJ-0040"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 7b6a1fb · 2026-09-29 · tests passed",
        "commit": "7b6a1fb",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0019"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0019",
      "phase": "outputs",
      "app": "ajiya",
      "title": "Dashboard",
      "done_when": "Adapted from reference/index.html for phases and tickets, with no RO work packages or MedInformer values: Overview with progress per milestone in order, Next up, Board, Phases, Tickets, Apps, Checks and Activity views, phases in display order, CSS variable themes for light and dark, embedded in the binary. The Dependencies view is its own ticket",
      "depends": [
        "AJ-0018",
        "AJ-0038"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 994ad17 · 2026-09-29 · tests passed",
        "commit": "994ad17",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0020",
        "AJ-0036",
        "AJ-0057",
        "AJ-0060",
        "AJ-0061",
        "AJ-0062",
        "AJ-0064"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0020",
      "phase": "outputs",
      "app": "ajiya",
      "title": "ajiya serve",
      "done_when": "Serves on 127.0.0.1 only, answers 403 to other Host headers, rebuilds when files change",
      "depends": [
        "AJ-0019"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 28b8618 · 2026-09-29 · tests passed",
        "commit": "28b8618",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0021",
        "AJ-0060"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0021",
      "phase": "outputs",
      "app": "ajiya",
      "title": "Milestones v0-1, v0-2 and v0-3 set and the dashboard in use on this repo",
      "done_when": "Milestones v0-1, v0-2 and v0-3 target the three version sign-off tickets; ajiya milestone list shows them in order; ajiya serve shows this repo's plan",
      "depends": [
        "AJ-0020",
        "AJ-0036",
        "AJ-0038",
        "AJ-0060",
        "AJ-0061",
        "AJ-0062",
        "AJ-0063",
        "AJ-0064"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 61d1f54 · 2026-09-29 · tests passed · Note: milestone list shows v0-1, v0-2, v0-3 in order; ajiya serve shows this repo's plan live",
        "note": "milestone list shows v0-1, v0-2, v0-3 in order; ajiya serve shows this repo's plan live",
        "commit": "61d1f54",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0022",
        "AJ-0023"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0022",
      "phase": "onboarding",
      "app": "ajiya",
      "title": "Setup guide",
      "done_when": ".ajiya/guide/setup.md covers planning from a goal and organising imported tickets, per spec section 6",
      "depends": [
        "AJ-0021"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 8d5cafe · 2026-09-29 · tests passed",
        "commit": "8d5cafe",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0024",
        "AJ-0041"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0023",
      "phase": "onboarding",
      "app": "ajiya",
      "title": "Daily guide",
      "done_when": ".ajiya/guide/daily.md covers the loop from ajiya next to ajiya check, per spec section 6",
      "depends": [
        "AJ-0021"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 8d5cafe · 2026-09-29 · tests passed",
        "commit": "8d5cafe",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0024"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0024",
      "phase": "onboarding",
      "app": "ajiya",
      "title": "Skill and agent instruction blocks",
      "done_when": "SKILL.md and the marked AGENTS.md and CLAUDE.md blocks point to the guides and list the rules",
      "depends": [
        "AJ-0022",
        "AJ-0023"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 39cb092 · 2026-09-28 · tests passed",
        "commit": "39cb092",
        "date": "2026-09-28",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0025"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0025",
      "phase": "onboarding",
      "app": "ajiya",
      "title": "init writes and updates the agent kit",
      "done_when": "Re-running init updates the kit and marked blocks without touching the user's text",
      "depends": [
        "AJ-0008",
        "AJ-0024"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 15cad8a · 2026-09-29 · tests passed",
        "commit": "15cad8a",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0029",
        "AJ-0066"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0026",
      "phase": "onboarding",
      "app": "ajiya",
      "title": "ajiya import todo",
      "done_when": "Checkbox and bullet items land in the inbox phase unchanged; ticked items are done before Ajiya",
      "depends": [],
      "status": {
        "state": "done",
        "text": "🟩 Done · 5cf99f1 · 2026-09-29 · tests passed",
        "commit": "5cf99f1",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "",
      "ready": false,
      "waiting_on": [],
      "dependants": [],
      "unblocks": 0
    },
    {
      "id": "AJ-0027",
      "phase": "onboarding",
      "app": "ajiya",
      "title": "ajiya import github",
      "done_when": "Uses gh to import open issues as pending and closed issues as done, each keeping its URL",
      "depends": [],
      "status": {
        "state": "done",
        "text": "🟩 Done · da34caf · 2026-09-29 · tests passed",
        "commit": "da34caf",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "",
      "ready": false,
      "waiting_on": [],
      "dependants": [],
      "unblocks": 0
    },
    {
      "id": "AJ-0028",
      "phase": "onboarding",
      "app": "ajiya",
      "title": "ajiya import legacy",
      "done_when": "Reads the collate.py format with new IDs, old IDs kept as aliases and a duplicate report; tested on reference/private if present",
      "depends": [],
      "status": {
        "state": "done",
        "text": "🟩 Done · 3b606af · 2026-09-29 · tests passed",
        "commit": "3b606af",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "",
      "ready": false,
      "waiting_on": [],
      "dependants": [],
      "unblocks": 0
    },
    {
      "id": "AJ-0029",
      "phase": "onboarding",
      "app": "ajiya",
      "title": "Re-run init on this repo",
      "done_when": "This repository's kit is the one users get",
      "depends": [
        "AJ-0025"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · c93b4be · 2026-09-29 · tests passed",
        "commit": "c93b4be",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0030",
        "AJ-0033"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0030",
      "phase": "release",
      "app": "ajiya",
      "title": "GoReleaser config",
      "done_when": "Snapshot build makes macOS, Linux and Windows binaries with the version set",
      "depends": [
        "AJ-0029"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 9a2c69c · 2026-09-29 · tests passed · Note: goreleaser release --snapshot --clean: 6 binaries, version stamped",
        "note": "goreleaser release --snapshot --clean: 6 binaries, version stamped",
        "commit": "9a2c69c",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0031",
        "AJ-0034",
        "AJ-0045",
        "AJ-0073",
        "AJ-0075"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0031",
      "phase": "release",
      "app": "ajiya",
      "title": "Homebrew cask config",
      "done_when": "homebrew_casks config with quarantine removal; not published",
      "depends": [
        "AJ-0030"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · bb6e0eb · 2026-09-29 · tests passed · Note: snapshot writes dist/homebrew/Casks/ajiya.rb with quarantine removal; skip_upload true",
        "note": "snapshot writes dist/homebrew/Casks/ajiya.rb with quarantine removal; skip_upload true",
        "commit": "bb6e0eb",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0034",
        "AJ-0045",
        "AJ-0077"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0032",
      "phase": "release",
      "app": "ajiya",
      "title": "Changelog from Ajiya trailers",
      "done_when": "Changelog groups commits by ticket and phase",
      "depends": [
        "AJ-0001"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · d84543c · 2026-09-29 · tests passed",
        "commit": "d84543c",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0034"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0033",
      "phase": "release",
      "app": "ajiya",
      "title": "README walkthroughs",
      "done_when": "README walks through a new project and an existing project",
      "depends": [
        "AJ-0029"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 6fb1bcc · 2026-09-29 · tests passed",
        "commit": "6fb1bcc",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0034"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0034",
      "phase": "release",
      "app": "ajiya",
      "title": "v0.1 definition of done",
      "done_when": "check --strict and check --commits since M2 pass, dashboard at 100 percent for the v0-1 milestone, both walkthroughs under ten minutes, every spec 01 ticket done",
      "depends": [
        "AJ-0030",
        "AJ-0031",
        "AJ-0032",
        "AJ-0033",
        "AJ-0037",
        "AJ-0038",
        "AJ-0039",
        "AJ-0040",
        "AJ-0041",
        "AJ-0059",
        "AJ-0065",
        "AJ-0066"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · by Aliyu Yahaya · 2026-09-29 · Note: v0.1 approved by the maintainer; check --strict and check --commits since M2 pass; agent-only walkthrough test runs after sign-off",
        "note": "v0.1 approved by the maintainer; check --strict and check --commits since M2 pass; agent-only walkthrough test runs after sign-off",
        "by": "Aliyu Yahaya",
        "date": "2026-09-29"
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0042",
        "AJ-0067",
        "AJ-0068",
        "AJ-0069",
        "AJ-0070",
        "AJ-0071",
        "AJ-0072"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0035",
      "phase": "outputs",
      "app": "ajiya",
      "title": "ajiya phase rename",
      "done_when": "ajiya phase rename \u003cold\u003e \u003cnew\u003e renames the file and moves every ticket; refused if the new slug exists; the launch target follows the rename",
      "depends": [],
      "status": {
        "state": "done",
        "text": "🟩 Done · 145c34f · 2026-09-29 · tests passed",
        "commit": "145c34f",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0059"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0036",
      "phase": "outputs",
      "app": "ajiya",
      "title": "Dependencies view in the dashboard",
      "done_when": "Phase graph by default, with an arrow where a ticket in one phase depends on a ticket in another; clicking a phase opens its ticket graph; Path to a milestone, with a milestone picker; hover traces a chain; stays readable with 500 tickets",
      "depends": [
        "AJ-0019"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · a0210e3 · 2026-09-29 · tests passed",
        "commit": "a0210e3",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0021",
        "AJ-0063"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0037",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "Milestone config and model",
      "done_when": "[[milestones]] in ajiya.toml, in file order, targets are phases or tickets; each ticket belongs to the earliest milestone that requires it, else unscheduled; old [launch] means a milestone named launch and both forms together is a config error; table tests cover several milestones, a shared ticket, unscheduled tickets and the old form",
      "depends": [
        "AJ-0010"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 8a327fa · 2026-09-29 · tests passed",
        "commit": "8a327fa",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0034",
        "AJ-0038",
        "AJ-0039",
        "AJ-0040"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0038",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "ajiya milestone commands and next by milestone",
      "done_when": "milestone add (launch stays last by default), list, show, move and remove, all with --json; launch set and show are shorthand for the launch milestone and their tests still pass; next sorts by earliest milestone then unblocks and shows it, with a milestone field in JSON; CLI tests cover each",
      "depends": [
        "AJ-0011",
        "AJ-0037"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · a36d139 · 2026-09-29 · tests passed",
        "commit": "a36d139",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0018",
        "AJ-0019",
        "AJ-0021",
        "AJ-0034",
        "AJ-0041",
        "AJ-0042",
        "AJ-0059"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0039",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "Milestone checks",
      "done_when": "Errors for a target that does not exist and a duplicate name; warnings for milestones in the wrong order and for open unscheduled tickets (up to 10 IDs); stable codes and a fixture each",
      "depends": [
        "AJ-0037"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 1828a44 · 2026-09-29 · tests passed",
        "commit": "1828a44",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0034"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0040",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "Phase display order",
      "done_when": "[phases] order in ajiya.toml, display only; phase move, phase list (--json) and phase order --suggest/--apply with a topological order, ties by milestone then creation, cycles kept together; phase rename updates it; a check warning for order against dependencies with a fixture; every listing uses the order",
      "depends": [
        "AJ-0037"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 999e8f1 · 2026-09-29 · tests passed",
        "commit": "999e8f1",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0018",
        "AJ-0034",
        "AJ-0041"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0041",
      "phase": "onboarding",
      "app": "ajiya",
      "title": "import legacy maps each RO package to a phase",
      "done_when": "Per spec 01 Part B: RO package to phase, area app per source, packages without scope, root-to-sink links, first package wins, cross-references dropped, inbox and archive, headings to milestones, area mode kept without a rollout file; fixture goldens for phases, ajiya.toml and report; check passes; package comment and the importing guide updated",
      "depends": [
        "AJ-0022",
        "AJ-0038",
        "AJ-0040"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 7d547a5 · 2026-09-29 · tests passed",
        "commit": "7d547a5",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0034"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0042",
      "phase": "agent-access",
      "app": "ajiya",
      "title": "ajiya status",
      "done_when": "Milestones in order, tickets in progress, what can start, Needs a human, check counts and the last five activity entries; --brief and --json; golden output for a fixture plan",
      "depends": [
        "AJ-0034",
        "AJ-0038"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": true,
      "waiting_on": [],
      "dependants": [
        "AJ-0043",
        "AJ-0046",
        "AJ-0047"
      ],
      "unblocks": 24
    },
    {
      "id": "AJ-0043",
      "phase": "agent-access",
      "app": "ajiya",
      "title": "ajiya mcp server",
      "done_when": "stdio MCP server in the same binary calling the CLI code, tools as listed in spec 02 with a test of the tools not offered; protocol tests for initialize, tools/list and tools/call; tools/list golden; parity with --json",
      "depends": [
        "AJ-0042"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0042"
      ],
      "dependants": [
        "AJ-0044",
        "AJ-0046",
        "AJ-0047"
      ],
      "unblocks": 23
    },
    {
      "id": "AJ-0044",
      "phase": "agent-access",
      "app": "ajiya",
      "title": "ajiya mcp install, uninstall and status",
      "done_when": "Registers ajiya mcp with Claude Code, Codex and Cursor from their current documentation, idempotent, backups, other entries kept, malformed files refused; init offers it; tests with a temporary home",
      "depends": [
        "AJ-0043"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0043"
      ],
      "dependants": [
        "AJ-0045",
        "AJ-0047",
        "AJ-0073",
        "AJ-0077"
      ],
      "unblocks": 21
    },
    {
      "id": "AJ-0045",
      "phase": "install",
      "app": "ajiya",
      "title": "install.sh for macOS and Linux",
      "done_when": "POSIX sh: detects OS and arch, resolves the latest tag or AJIYA_VERSION, downloads the archive and checksums.txt, refuses on a checksum mismatch, installs to AJIYA_INSTALL_DIR or ~/.local/bin without sudo, prints PATH advice, offers ajiya mcp install --all only on a terminal; --uninstall; installs from a local snapshot fixture in a test",
      "depends": [
        "AJ-0030",
        "AJ-0031",
        "AJ-0044"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0044"
      ],
      "dependants": [
        "AJ-0047",
        "AJ-0074",
        "AJ-0078"
      ],
      "unblocks": 16
    },
    {
      "id": "AJ-0046",
      "phase": "install",
      "app": "ajiya",
      "title": "Claude Code plugin",
      "done_when": "plugin/ with the skill, the MCP entry and a session-start hook running ajiya status --brief; validates against the documented schema; hook tested with and without ajiya on PATH",
      "depends": [
        "AJ-0042",
        "AJ-0043"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0042",
        "AJ-0043"
      ],
      "dependants": [
        "AJ-0047"
      ],
      "unblocks": 12
    },
    {
      "id": "AJ-0047",
      "phase": "install",
      "app": "ajiya",
      "title": "v0.2 definition of done",
      "done_when": "On a clean account in under ten minutes from the README: install, register with Claude Code and Codex, plan a project and finish one ticket through MCP tools",
      "depends": [
        "AJ-0042",
        "AJ-0043",
        "AJ-0044",
        "AJ-0045",
        "AJ-0046",
        "AJ-0082"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending · Needs a human: final sign-off before publishing v0.2",
        "human": "final sign-off before publishing v0.2"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0042",
        "AJ-0043",
        "AJ-0044",
        "AJ-0045",
        "AJ-0046",
        "AJ-0082"
      ],
      "dependants": [
        "AJ-0048"
      ],
      "unblocks": 11
    },
    {
      "id": "AJ-0048",
      "phase": "decisions-and-changes",
      "app": "ajiya",
      "title": "Tags column",
      "done_when": "Optional seventh Tags column with tech, dec and path tags only (no comp), written from flags; a ticket with no tags is valid and check never warns about missing tags; six-column files read and upgraded; ajiya tags; near-duplicate warning for tech values; round-trip goldens for both forms; MCP tools",
      "depends": [
        "AJ-0047"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-3",
      "ready": false,
      "waiting_on": [
        "AJ-0047"
      ],
      "dependants": [
        "AJ-0049",
        "AJ-0050",
        "AJ-0051",
        "AJ-0058"
      ],
      "unblocks": 10
    },
    {
      "id": "AJ-0049",
      "phase": "decisions-and-changes",
      "app": "ajiya",
      "title": "ajiya ticket files",
      "done_when": "Lists every file changed by commits referencing a ticket, from git; CLI test and MCP tool",
      "depends": [
        "AJ-0048"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-3",
      "ready": false,
      "waiting_on": [
        "AJ-0048"
      ],
      "dependants": [
        "AJ-0053",
        "AJ-0058"
      ],
      "unblocks": 6
    },
    {
      "id": "AJ-0050",
      "phase": "decisions-and-changes",
      "app": "ajiya",
      "title": "Decision records and commands",
      "done_when": "ajiya/decisions/DEC-xxxx.md written only by the CLI; decision add (with --tech, no --comp), show, list and accept with --json; ticket show lists linked decisions; checks for missing and superseded decisions; goldens and MCP tools",
      "depends": [
        "AJ-0048"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-3",
      "ready": false,
      "waiting_on": [
        "AJ-0048"
      ],
      "dependants": [
        "AJ-0053",
        "AJ-0058"
      ],
      "unblocks": 6
    },
    {
      "id": "AJ-0051",
      "phase": "decisions-and-changes",
      "app": "ajiya",
      "title": "Superseded status",
      "done_when": "Superseded by \u003cID\u003e · CHG · was: \u003cold status\u003e keeps the old evidence; excluded from totals and required sets; round-trip goldens",
      "depends": [
        "AJ-0048"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-3",
      "ready": false,
      "waiting_on": [
        "AJ-0048"
      ],
      "dependants": [
        "AJ-0052",
        "AJ-0053",
        "AJ-0058"
      ],
      "unblocks": 7
    },
    {
      "id": "AJ-0052",
      "phase": "decisions-and-changes",
      "app": "ajiya",
      "title": "Snapshots as git tags",
      "done_when": "ajiya snapshot \u003cname\u003e tags ajiya/plan/\u003cname\u003e, refused with uncommitted ajiya/ changes; snapshot list; CLI tests",
      "depends": [
        "AJ-0051"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-3",
      "ready": false,
      "waiting_on": [
        "AJ-0051"
      ],
      "dependants": [
        "AJ-0053",
        "AJ-0058"
      ],
      "unblocks": 6
    },
    {
      "id": "AJ-0053",
      "phase": "decisions-and-changes",
      "app": "ajiya",
      "title": "change open and change impact",
      "done_when": "Change files written only by the CLI with a snapshot; impact finds candidates by tag, decision, path (path tags and the files in each ticket's commits) and one level downstream, with reasons, and still finds tickets that have no tags; verdicts kept on rerun; goldens, a no-tags case and MCP tools",
      "depends": [
        "AJ-0049",
        "AJ-0050",
        "AJ-0051",
        "AJ-0052"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-3",
      "ready": false,
      "waiting_on": [
        "AJ-0049",
        "AJ-0050",
        "AJ-0051",
        "AJ-0052"
      ],
      "dependants": [
        "AJ-0054",
        "AJ-0058"
      ],
      "unblocks": 5
    },
    {
      "id": "AJ-0054",
      "phase": "decisions-and-changes",
      "app": "ajiya",
      "title": "change assess",
      "done_when": "Records unaffected, modify, replace, invalidate or investigate with a reason, and the new ticket text where needed; CLI tests and MCP tool",
      "depends": [
        "AJ-0053"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-3",
      "ready": false,
      "waiting_on": [
        "AJ-0053"
      ],
      "dependants": [
        "AJ-0055",
        "AJ-0058"
      ],
      "unblocks": 4
    },
    {
      "id": "AJ-0055",
      "phase": "decisions-and-changes",
      "app": "ajiya",
      "title": "change approve and change reject",
      "done_when": "approve applies every verdict in one write (new tickets, edits, superseded, moved dependencies, superseded decisions) and records who and when; refused with missing or investigate verdicts; reject leaves the plan alone; not offered over MCP; the Supabase scenario test passes",
      "depends": [
        "AJ-0054"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-3",
      "ready": false,
      "waiting_on": [
        "AJ-0054"
      ],
      "dependants": [
        "AJ-0056",
        "AJ-0057",
        "AJ-0058"
      ],
      "unblocks": 3
    },
    {
      "id": "AJ-0056",
      "phase": "decisions-and-changes",
      "app": "ajiya",
      "title": "Guides and skill for decisions and changes",
      "done_when": "Guides and skill say tags are optional, add tech only when a ticket depends on a specific technology or service (tech:aws-rds on Configure RDS backups, none on Add a booking form), use path only while a ticket has no commits, never store what the CLI can work out; how to record decisions and run the change procedure; agents never approve or reject",
      "depends": [
        "AJ-0055"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-3",
      "ready": false,
      "waiting_on": [
        "AJ-0055"
      ],
      "dependants": [
        "AJ-0058"
      ],
      "unblocks": 1
    },
    {
      "id": "AJ-0057",
      "phase": "decisions-and-changes",
      "app": "ajiya",
      "title": "Dashboard: open changes and decisions",
      "done_when": "Open changes with their impact table and a Decisions view in the dashboard",
      "depends": [
        "AJ-0019",
        "AJ-0055"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "after-v0-3",
      "ready": false,
      "waiting_on": [
        "AJ-0055"
      ],
      "dependants": [],
      "unblocks": 0
    },
    {
      "id": "AJ-0058",
      "phase": "decisions-and-changes",
      "app": "ajiya",
      "title": "v0.3 definition of done",
      "done_when": "The scenario test passes, and a person runs one real change through a sample repository from change open to approve and reviews the result",
      "depends": [
        "AJ-0048",
        "AJ-0049",
        "AJ-0050",
        "AJ-0051",
        "AJ-0052",
        "AJ-0053",
        "AJ-0054",
        "AJ-0055",
        "AJ-0056"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending · Needs a human: final sign-off before publishing v0.3",
        "human": "final sign-off before publishing v0.3"
      },
      "milestone": "v0-3",
      "ready": false,
      "waiting_on": [
        "AJ-0048",
        "AJ-0049",
        "AJ-0050",
        "AJ-0051",
        "AJ-0052",
        "AJ-0053",
        "AJ-0054",
        "AJ-0055",
        "AJ-0056"
      ],
      "dependants": [],
      "unblocks": 0
    },
    {
      "id": "AJ-0059",
      "phase": "apps-launch",
      "app": "ajiya",
      "title": "phase rename updates milestone targets",
      "done_when": "phase rename also renames the phase in every [[milestones]] targets list, keeping comments; script test",
      "depends": [
        "AJ-0035",
        "AJ-0038"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 57776dc · 2026-09-29 · tests passed",
        "commit": "57776dc",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0034"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0060",
      "phase": "outputs",
      "app": "ajiya",
      "title": "Dashboard live refresh under ajiya serve",
      "done_when": "The dashboard served by ajiya serve polls /version and reloads its data when it changes, without losing the current view or filters; opened from disk it does not poll; tested in the dashboard tests",
      "depends": [
        "AJ-0019",
        "AJ-0020"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 3ae5531 · 2026-09-29 · tests passed",
        "commit": "3ae5531",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0021"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0061",
      "phase": "outputs",
      "app": "ajiya",
      "title": "Brand logo in the dashboard",
      "done_when": "The sidebar shows ajiya/brand-logo.png scaled for 2x screens and embedded as a base64 data URI (no extra file or request), legible in light and dark themes; collapsed sidebar hides it like the reference",
      "depends": [
        "AJ-0019"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · d84de11 · 2026-09-29 · tests passed",
        "commit": "d84de11",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0021"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0062",
      "phase": "outputs",
      "app": "ajiya",
      "title": "Search and filters on every dashboard view",
      "done_when": "The filter bar (search, app, phase, milestone, state, and Can start now / Needs a human pills) shows on every view the reference shows it on, including Overview and Dependencies; search also narrows Checks and Activity; a test checks each view's filters flag",
      "depends": [
        "AJ-0019"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · d84de11 · 2026-09-29 · tests passed",
        "commit": "d84de11",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0021"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0063",
      "phase": "outputs",
      "app": "ajiya",
      "title": "Dependencies view: toggle between ticket links and phase links",
      "done_when": "A Tickets / Phases toggle on the Dependencies view; Tickets mode draws every ticket as a card like the reference, with All / Only linked and path to a milestone, and fades tickets that miss the filters; Phases mode keeps the phase graph and its click-through; the choice is remembered",
      "depends": [
        "AJ-0036"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · d84de11 · 2026-09-29 · tests passed",
        "commit": "d84de11",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0021"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0064",
      "phase": "outputs",
      "app": "ajiya",
      "title": "Ticket cards without the side border outside Dependencies",
      "done_when": "Ticket cards on Overview, Next up, Board, Phases and the rest have no coloured left border; only Dependencies graph nodes keep the state stripe",
      "depends": [
        "AJ-0019"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 40d743d · 2026-09-29 · tests passed",
        "commit": "40d743d",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0021"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0065",
      "phase": "commit-rule",
      "app": "ajiya",
      "title": "ticket done takes evidence only from commits after the ticket started",
      "done_when": "ticket start records the start time locally (in the git directory); ticket done refuses when every commit naming the ticket predates it and names the ignored commit; tests cover a pre-start commit alone, a pre-start plus a later commit, and a ticket with no start record",
      "depends": [],
      "status": {
        "state": "done",
        "text": "🟩 Done · 5c1af99 · 2026-09-29 · tests passed",
        "commit": "5c1af99",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0034"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0066",
      "phase": "onboarding",
      "app": "ajiya",
      "title": "Guides say how to set the test command",
      "done_when": "The kit no longer forbids editing ajiya.toml outright: ajiya/ is never edited by hand, and ajiya.toml changes through commands except the [test] command, which the guides say to set by hand; ticket done --test without one is explained",
      "depends": [
        "AJ-0025"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · 5a75b38 · 2026-09-29 · tests passed",
        "commit": "5a75b38",
        "date": "2026-09-29",
        "tests_passed": true
      },
      "milestone": "v0-1",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0034"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0067",
      "phase": "install",
      "app": "ajiya",
      "title": "Decide how the npm package ships the binary",
      "done_when": "Decision recorded: per-platform optional packages behind a wrapper (recommended: works with --ignore-scripts and registry mirrors) or a postinstall download as spec 02 words it",
      "depends": [
        "AJ-0034"
      ],
      "status": {
        "state": "done",
        "text": "🟩 Done · by Aliyu Yahaya · 2026-09-29 · Note: npm ships per-platform optional packages behind a small ajiya wrapper, not a postinstall download (agreed in session)",
        "note": "npm ships per-platform optional packages behind a small ajiya wrapper, not a postinstall download (agreed in session)",
        "by": "Aliyu Yahaya",
        "date": "2026-09-29"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [],
      "dependants": [
        "AJ-0075"
      ],
      "unblocks": 0
    },
    {
      "id": "AJ-0068",
      "phase": "install",
      "app": "ajiya",
      "title": "npm account and package names",
      "done_when": "npm account with 2FA; the name ajiya claimed; platform package names chosen (scoped @ajiya/* or unscoped ajiya-\u003cos\u003e-\u003carch\u003e) and written in the ticket note",
      "depends": [
        "AJ-0034"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending · Needs a human: needs the maintainer's npm account",
        "human": "needs the maintainer's npm account"
      },
      "milestone": "v0-2",
      "ready": true,
      "waiting_on": [],
      "dependants": [
        "AJ-0075"
      ],
      "unblocks": 18
    },
    {
      "id": "AJ-0069",
      "phase": "install",
      "app": "ajiya",
      "title": "Homebrew tap repository and token",
      "done_when": "AliyuYahaya/homebrew-tap exists; a fine-grained token with Contents read and write on it alone is stored as HOMEBREW_TAP_TOKEN in a homebrew-release environment",
      "depends": [
        "AJ-0034"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending · Needs a human: needs the maintainer's GitHub account",
        "human": "needs the maintainer's GitHub account"
      },
      "milestone": "v0-2",
      "ready": true,
      "waiting_on": [],
      "dependants": [
        "AJ-0077"
      ],
      "unblocks": 17
    },
    {
      "id": "AJ-0070",
      "phase": "install",
      "app": "ajiya",
      "title": "Choose the install script URL",
      "done_when": "Decision recorded: a custom domain served by GitHub Pages, or the release asset URL releases/latest/download/install.sh",
      "depends": [
        "AJ-0034"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending · Needs a human: a domain costs money and is the maintainer's choice",
        "human": "a domain costs money and is the maintainer's choice"
      },
      "milestone": "v0-2",
      "ready": true,
      "waiting_on": [],
      "dependants": [
        "AJ-0078"
      ],
      "unblocks": 15
    },
    {
      "id": "AJ-0071",
      "phase": "install",
      "app": "ajiya",
      "title": "Release environments with required reviewers",
      "done_when": "GitHub environments npm-release and homebrew-release exist with the maintainer as required reviewer",
      "depends": [
        "AJ-0034"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending · Needs a human: repository settings need the maintainer",
        "human": "repository settings need the maintainer"
      },
      "milestone": "v0-2",
      "ready": true,
      "waiting_on": [],
      "dependants": [
        "AJ-0079",
        "AJ-0083"
      ],
      "unblocks": 17
    },
    {
      "id": "AJ-0072",
      "phase": "install",
      "app": "ajiya",
      "title": "Release workflow: build to a draft release",
      "done_when": "release.yml on a v* tag writes release notes with ajiya changelog, runs GoReleaser to a draft release and keeps dist/ as an artifact; on pull requests it runs a snapshot build; a test run publishes nothing",
      "depends": [
        "AJ-0034"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": true,
      "waiting_on": [],
      "dependants": [
        "AJ-0079",
        "AJ-0084"
      ],
      "unblocks": 16
    },
    {
      "id": "AJ-0073",
      "phase": "install",
      "app": "ajiya",
      "title": "install.ps1 for Windows",
      "done_when": "Detects arch, downloads the zip and checksums.txt, refuses on a mismatch, installs to LOCALAPPDATA\\Programs\\ajiya\\bin and adds it to the user PATH; AJIYA_VERSION; installs from a local snapshot fixture on a Windows runner",
      "depends": [
        "AJ-0030",
        "AJ-0044"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0044"
      ],
      "dependants": [
        "AJ-0074",
        "AJ-0078"
      ],
      "unblocks": 16
    },
    {
      "id": "AJ-0074",
      "phase": "install",
      "app": "ajiya",
      "title": "Installer CI",
      "done_when": "shellcheck and PSScriptAnalyzer pass; a macOS, Linux and Windows matrix installs from a snapshot fixture served locally, runs ajiya version, and a tampered archive makes the script fail without installing",
      "depends": [
        "AJ-0045",
        "AJ-0073"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0045",
        "AJ-0073"
      ],
      "dependants": [
        "AJ-0081"
      ],
      "unblocks": 14
    },
    {
      "id": "AJ-0075",
      "phase": "install",
      "app": "ajiya",
      "title": "npm packages: a wrapper and one package per platform",
      "done_when": "A script verifies every release archive against checksums.txt, then writes six platform packages (os and cpu set, binary mode 755) and the ajiya wrapper with exact optionalDependencies, all versioned from the tag; npm pack --dry-run succeeds for all seven; a bad checksum fails it",
      "depends": [
        "AJ-0030",
        "AJ-0067",
        "AJ-0068"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0068"
      ],
      "dependants": [
        "AJ-0076",
        "AJ-0079",
        "AJ-0080"
      ],
      "unblocks": 17
    },
    {
      "id": "AJ-0076",
      "phase": "install",
      "app": "ajiya",
      "title": "npm launcher tests",
      "done_when": "node --test on Linux, macOS and Windows: the launcher runs the binary with its arguments and exit code, and a missing platform binary gives a clear error that points to the install script and Homebrew",
      "depends": [
        "AJ-0075"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0075"
      ],
      "dependants": [
        "AJ-0081"
      ],
      "unblocks": 14
    },
    {
      "id": "AJ-0077",
      "phase": "install",
      "app": "ajiya",
      "title": "Homebrew tap wiring",
      "done_when": "skip_upload: auto; the cask caveat prints the one command to register with agents (ajiya mcp install); brew install --cask from a local tap installs a snapshot build",
      "depends": [
        "AJ-0031",
        "AJ-0044",
        "AJ-0069"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0044",
        "AJ-0069"
      ],
      "dependants": [
        "AJ-0079",
        "AJ-0084"
      ],
      "unblocks": 16
    },
    {
      "id": "AJ-0078",
      "phase": "install",
      "app": "ajiya",
      "title": "Install scripts served at a stable URL",
      "done_when": "install.sh and install.ps1 are attached to every release, and served from the chosen URL; curl -I on a test release resolves them",
      "depends": [
        "AJ-0045",
        "AJ-0070",
        "AJ-0073"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0045",
        "AJ-0070",
        "AJ-0073"
      ],
      "dependants": [
        "AJ-0081"
      ],
      "unblocks": 14
    },
    {
      "id": "AJ-0079",
      "phase": "install",
      "app": "ajiya",
      "title": "Gated publish jobs",
      "done_when": "npm and tap jobs in release.yml run only after approval in their environments; a workflow_dispatch dry run reaches the gate and, approved, completes npm publish --dry-run without publishing",
      "depends": [
        "AJ-0071",
        "AJ-0072",
        "AJ-0075",
        "AJ-0077"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0071",
        "AJ-0072",
        "AJ-0075",
        "AJ-0077"
      ],
      "dependants": [
        "AJ-0081"
      ],
      "unblocks": 14
    },
    {
      "id": "AJ-0080",
      "phase": "install",
      "app": "ajiya",
      "title": "npm trusted publishing",
      "done_when": "Each npm package has had its first publish and a trusted-publisher entry for release.yml on npmjs.com, so releases need no npm token",
      "depends": [
        "AJ-0075"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending · Needs a human: needs the maintainer's npm account",
        "human": "needs the maintainer's npm account"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0075"
      ],
      "dependants": [
        "AJ-0081"
      ],
      "unblocks": 14
    },
    {
      "id": "AJ-0081",
      "phase": "install",
      "app": "ajiya",
      "title": "Release rehearsal on a release candidate",
      "done_when": "A v0.2.0-rc tag produces a draft release, npm packages under the next dist-tag with provenance, and a tap pull request; nothing is public until the maintainer publishes",
      "depends": [
        "AJ-0074",
        "AJ-0076",
        "AJ-0078",
        "AJ-0079",
        "AJ-0080",
        "AJ-0084"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending · Needs a human: publishing anything needs the maintainer's approval",
        "human": "publishing anything needs the maintainer's approval"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0074",
        "AJ-0076",
        "AJ-0078",
        "AJ-0079",
        "AJ-0080",
        "AJ-0084"
      ],
      "dependants": [
        "AJ-0082"
      ],
      "unblocks": 13
    },
    {
      "id": "AJ-0082",
      "phase": "install",
      "app": "ajiya",
      "title": "Install docs for every channel",
      "done_when": "README (and landing page if chosen) has one install line each for Homebrew, npm, the install script and Windows, plus the what-just-happened note from spec 02; each line matches a tested command",
      "depends": [
        "AJ-0081"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0081"
      ],
      "dependants": [
        "AJ-0047"
      ],
      "unblocks": 12
    },
    {
      "id": "AJ-0083",
      "phase": "install",
      "app": "ajiya",
      "title": "Apple signing credentials as release secrets",
      "done_when": "A Developer ID Application certificate (.p12) and its password, and an App Store Connect API key (.p8) with its key ID and issuer ID, are stored as secrets in a release environment",
      "depends": [
        "AJ-0071"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending · Needs a human: needs the maintainer's Apple Developer account",
        "human": "needs the maintainer's Apple Developer account"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0071"
      ],
      "dependants": [
        "AJ-0084"
      ],
      "unblocks": 15
    },
    {
      "id": "AJ-0084",
      "phase": "install",
      "app": "ajiya",
      "title": "Sign and notarise the macOS binaries",
      "done_when": "GoReleaser signs and notarises the darwin binaries (notarize.macos) in the release workflow; spctl -a -vv accepts a binary downloaded from a test release; the cask's quarantine-removal hook is removed",
      "depends": [
        "AJ-0072",
        "AJ-0077",
        "AJ-0083"
      ],
      "status": {
        "state": "pending",
        "text": "🟥 Pending"
      },
      "milestone": "v0-2",
      "ready": false,
      "waiting_on": [
        "AJ-0072",
        "AJ-0077",
        "AJ-0083"
      ],
      "dependants": [
        "AJ-0081"
      ],
      "unblocks": 14
    }
  ],
  "next": [
    "AJ-0042",
    "AJ-0068",
    "AJ-0069",
    "AJ-0071",
    "AJ-0072",
    "AJ-0070"
  ],
  "checks": [],
  "activity": [
    {
      "hash": "6a75e6037a1643a56ba57e1315f6852f5b676e43",
      "short": "6a75e60",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "v0.1 signed off",
      "refs": [
        "AJ-0034"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0034",
          "phase": "release",
          "from": "pending",
          "to": "done",
          "text": "🟩 Done · by Aliyu Yahaya · 2026-09-29 · Note: v0.1 approved by the maintainer; check --strict and check --commits since M2 pass; agent-only walkthrough test runs after sign-off"
        }
      ],
      "time": 1790716084
    },
    {
      "hash": "e803fc9d4044184273e19ce1779ab0f715a7ca3f",
      "short": "e803fc9",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a658cbd8dac8f5d6c'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790715883
    },
    {
      "hash": "e0873df13864a3cb5b805b76823a3162fa673aa6",
      "short": "e0873df",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Licence held by Yavid PTY Ltd; the name and logo are not licensed",
      "refs": [],
      "chore": true,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790715876
    },
    {
      "hash": "2603a19ee989dac6d4b0eea564038fa6e1011edc",
      "short": "2603a19",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "MIT licence",
      "refs": [],
      "chore": true,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790715377
    },
    {
      "hash": "110dcb2c7f6d12dd262bd60594a7920b1452346f",
      "short": "110dcb2",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Rebuild plan outputs after the merge",
      "refs": [],
      "chore": true,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790715129
    },
    {
      "hash": "cc33bdb225fb147343288c8a3ae9ef770c4f522b",
      "short": "cc33bdb",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a5a988be55680e087' into worktree-agent-a658cbd8dac8f5d6c",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790715103
    },
    {
      "hash": "2d65774cb12f8dd9d743e49544c397059b080c83",
      "short": "2d65774",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Plan: AJ-0041 done",
      "refs": [
        "AJ-0041"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0041",
          "phase": "onboarding",
          "from": "pending",
          "to": "done",
          "text": "🟩 Done · 7d547a5 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790714947
    },
    {
      "hash": "7d547a524020425284e59f1ad64d5112ffb32b62",
      "short": "7d547a5",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "import legacy: map each RO package to a phase",
      "refs": [
        "AJ-0041"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790714670
    },
    {
      "hash": "ff9cdca90b676672305c3228fcbfebd3927b8cee",
      "short": "ff9cdca",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Plan: sign and notarise the macOS binaries",
      "refs": [],
      "chore": true,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0083",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending · Needs a human: needs the maintainer's Apple Developer account"
        },
        {
          "id": "AJ-0084",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        }
      ],
      "time": 1790713405
    },
    {
      "hash": "810c3c0e1576565d83bac786416dfd1b850df279",
      "short": "810c3c0",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "npm packaging decided: per-platform packages behind a wrapper",
      "refs": [
        "AJ-0067",
        "AJ-0075"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0067",
          "phase": "install",
          "from": "pending",
          "to": "done",
          "text": "🟩 Done · by Aliyu Yahaya · 2026-09-29 · Note: npm ships per-platform optional packages behind a small ajiya wrapper, not a postinstall download (agreed in session)"
        }
      ],
      "time": 1790713318
    },
    {
      "hash": "2ba31d5c99eefb96618996039fbaca561d2612cf",
      "short": "2ba31d5",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Reflow the daily guide's done paragraph",
      "refs": [
        "AJ-0065"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790713300
    },
    {
      "hash": "622e67856c1e157906e1ff4aa92307fb22af33e5",
      "short": "622e678",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-af0c9a6c1037d2858' into worktree-agent-a658cbd8dac8f5d6c",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790713282
    },
    {
      "hash": "140a432e17f79b84388154d37fe04dad3d0221f9",
      "short": "140a432",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0065 done",
      "refs": [
        "AJ-0065"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0065",
          "phase": "commit-rule",
          "from": "pending",
          "to": "done",
          "text": "🟩 Done · 5c1af99 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790713202
    },
    {
      "hash": "5c1af9978623670e6ad96f84e17d4d641d1871bf",
      "short": "5c1af99",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Ticket done counts only commits made after the ticket started",
      "refs": [
        "AJ-0065"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790713063
    },
    {
      "hash": "619df03ca1896569f885249708a822acd3e92868",
      "short": "619df03",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Plan the install channels: npm, Homebrew and the install script",
      "refs": [],
      "chore": true,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0067",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending · Needs a human: changes spec 02's wording; the maintainer decides"
        },
        {
          "id": "AJ-0068",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending · Needs a human: needs the maintainer's npm account"
        },
        {
          "id": "AJ-0069",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending · Needs a human: needs the maintainer's GitHub account"
        },
        {
          "id": "AJ-0070",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending · Needs a human: a domain costs money and is the maintainer's choice"
        },
        {
          "id": "AJ-0071",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending · Needs a human: repository settings need the maintainer"
        },
        {
          "id": "AJ-0072",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0073",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0074",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0075",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0076",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0077",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0078",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0079",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0080",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending · Needs a human: needs the maintainer's npm account"
        },
        {
          "id": "AJ-0081",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending · Needs a human: publishing anything needs the maintainer's approval"
        },
        {
          "id": "AJ-0082",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        }
      ],
      "time": 1790713032
    },
    {
      "hash": "7748eae54cb34604773a6deb55f3fe8648bc4411",
      "short": "7748eae",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a658cbd8dac8f5d6c'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790712885
    },
    {
      "hash": "d0c5d61ab6f51a54a306aafb9392197706f50f5f",
      "short": "d0c5d61",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0066 done",
      "refs": [
        "AJ-0066"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0066",
          "phase": "onboarding",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 5a75b38 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790712816
    },
    {
      "hash": "5a75b3826b94add680132a239f1d75fe68760be0",
      "short": "5a75b38",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Guides: how to set the test command",
      "refs": [
        "AJ-0066"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0066",
          "phase": "onboarding",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790712812
    },
    {
      "hash": "60ba1a327450c7a7a0776e57e57b702ee5b738cb",
      "short": "60ba1a3",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Plan: guides explain the test command",
      "refs": [],
      "chore": true,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0066",
          "phase": "onboarding",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        }
      ],
      "time": 1790712583
    },
    {
      "hash": "c6f6b5aed1691036f5c3eead0c714d27435fc3df",
      "short": "c6f6b5a",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0033 done",
      "refs": [
        "AJ-0033"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0033",
          "phase": "release",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 6fb1bcc · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790712549
    },
    {
      "hash": "6fb1bccd2d6c3aefd697b8377e509dcb96d9e4d8",
      "short": "6fb1bcc",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "README: how it works, install, and the two walkthroughs",
      "refs": [
        "AJ-0033"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0033",
          "phase": "release",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790712545
    },
    {
      "hash": "edf9e224c4ea6f1ec9723cb61d4391c5277ea7cc",
      "short": "edf9e22",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0031 done",
      "refs": [
        "AJ-0031"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0031",
          "phase": "release",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · bb6e0eb · 2026-09-29 · tests passed · Note: snapshot writes dist/homebrew/Casks/ajiya.rb with quarantine removal; skip_upload true"
        }
      ],
      "time": 1790712371
    },
    {
      "hash": "bb6e0ebe1362d64f52b80eebbd4b8ffb71873da3",
      "short": "bb6e0eb",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Homebrew cask config",
      "refs": [
        "AJ-0031"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0031",
          "phase": "release",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790712367
    },
    {
      "hash": "03b432dbfb59b411171640fe293d0426d8f03235",
      "short": "03b432d",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0030 done",
      "refs": [
        "AJ-0030"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0030",
          "phase": "release",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 9a2c69c · 2026-09-29 · tests passed · Note: goreleaser release --snapshot --clean: 6 binaries, version stamped"
        }
      ],
      "time": 1790711870
    },
    {
      "hash": "9a2c69c5d9375fd7c2e659f341b18b58af2bef32",
      "short": "9a2c69c",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "GoReleaser config",
      "refs": [
        "AJ-0030"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0030",
          "phase": "release",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790711866
    },
    {
      "hash": "d18ad9e34ebeee7270a29884ed7027a63f7678f4",
      "short": "d18ad9e",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0029 done",
      "refs": [
        "AJ-0029"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0029",
          "phase": "onboarding",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · c93b4be · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790710883
    },
    {
      "hash": "c93b4bea8c2ba7d8762fb91374e7000502028c8f",
      "short": "c93b4be",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "This repository runs the kit it ships",
      "refs": [
        "AJ-0029"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0029",
          "phase": "onboarding",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790710879
    },
    {
      "hash": "298f56f505a7eefbcb1bef83c21c2b8aa897cce7",
      "short": "298f56f",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0025 done",
      "refs": [
        "AJ-0025"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0025",
          "phase": "onboarding",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 15cad8a · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790710670
    },
    {
      "hash": "15cad8ae073d3244eebcac01ed58c626596c16d7",
      "short": "15cad8a",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "init writes and updates the agent kit",
      "refs": [
        "AJ-0025"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0025",
          "phase": "onboarding",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790710637
    },
    {
      "hash": "4f9b1a62134d072f3fdc39df716c45ac117a708e",
      "short": "4f9b1a6",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Plan: ticket done evidence must come after the ticket started",
      "refs": [],
      "chore": true,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0065",
          "phase": "commit-rule",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        }
      ],
      "time": 1790709594
    },
    {
      "hash": "5556d64f91d22caf0ff3a4c8813039604afc0afd",
      "short": "5556d64",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "CLAUDE.md with the Ajiya block",
      "refs": [
        "AJ-0024"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790709574
    },
    {
      "hash": "0c78c499fce55618c2773391840721d7f94f91b7",
      "short": "0c78c49",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0024 done",
      "refs": [
        "AJ-0024"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0024",
          "phase": "onboarding",
          "from": "pending",
          "to": "done",
          "text": "🟩 Done · 39cb092 · 2026-09-28 · tests passed"
        }
      ],
      "time": 1790709475
    },
    {
      "hash": "f6b5387ed38098cd7703d5d13952581affef7db7",
      "short": "f6b5387",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0022 and AJ-0023 done",
      "refs": [
        "AJ-0022",
        "AJ-0023"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0022",
          "phase": "onboarding",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 8d5cafe · 2026-09-29 · tests passed"
        },
        {
          "id": "AJ-0023",
          "phase": "onboarding",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 8d5cafe · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790709280
    },
    {
      "hash": "8d5cafea01fae6e046b5c0108394b3c41ee970cd",
      "short": "8d5cafe",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Setup and daily guides for agents",
      "refs": [
        "AJ-0022",
        "AJ-0023"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0022",
          "phase": "onboarding",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        },
        {
          "id": "AJ-0023",
          "phase": "onboarding",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790709236
    },
    {
      "hash": "47117c5f6f586ee6d1fffab898eb34e0f750d684",
      "short": "47117c5",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a658cbd8dac8f5d6c'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790708945
    },
    {
      "hash": "4b321be611a5f9f1cf6ed8de1260a5d3cbb4a2bf",
      "short": "4b321be",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0021 done",
      "refs": [
        "AJ-0021"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0021",
          "phase": "outputs",
          "from": "pending",
          "to": "done",
          "text": "🟩 Done · 61d1f54 · 2026-09-29 · tests passed · Note: milestone list shows v0-1, v0-2, v0-3 in order; ajiya serve shows this repo's plan live"
        }
      ],
      "time": 1790708918
    },
    {
      "hash": "dcb35011bec7eae072c99a4be9cb170514b2472c",
      "short": "dcb3501",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0061 to AJ-0064 done",
      "refs": [],
      "chore": true,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0061",
          "phase": "outputs",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · d84de11 · 2026-09-29 · tests passed"
        },
        {
          "id": "AJ-0062",
          "phase": "outputs",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · d84de11 · 2026-09-29 · tests passed"
        },
        {
          "id": "AJ-0063",
          "phase": "outputs",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · d84de11 · 2026-09-29 · tests passed"
        },
        {
          "id": "AJ-0064",
          "phase": "outputs",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 40d743d · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790708649
    },
    {
      "hash": "40d743defd65b40583bf2303fb4cc1929b43056f",
      "short": "40d743d",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Ticket cards without the side stripe",
      "refs": [
        "AJ-0064"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790708324
    },
    {
      "hash": "d84de11fc0f5cffbc6776d492f36efadde05a5ce",
      "short": "d84de11",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Dashboard: logo, filters on every view, ticket and phase dependency graphs",
      "refs": [
        "AJ-0061",
        "AJ-0062",
        "AJ-0063"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0061",
          "phase": "outputs",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        },
        {
          "id": "AJ-0062",
          "phase": "outputs",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        },
        {
          "id": "AJ-0063",
          "phase": "outputs",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        },
        {
          "id": "AJ-0064",
          "phase": "outputs",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790708324
    },
    {
      "hash": "72a0c1f34578f06dda176955567cd9556f4fe2cf",
      "short": "72a0c1f",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Plan dashboard polish: logo, filters everywhere, dependency toggle, card borders",
      "refs": [],
      "chore": true,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0061",
          "phase": "outputs",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0062",
          "phase": "outputs",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0063",
          "phase": "outputs",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0064",
          "phase": "outputs",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        }
      ],
      "time": 1790707833
    },
    {
      "hash": "8ced23318e816cc1367c7bc6ca7277194585b40b",
      "short": "8ced233",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a658cbd8dac8f5d6c'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790706988
    },
    {
      "hash": "f84a39b75381ace7177b72a7e9fca7f9f1c14ed4",
      "short": "f84a39b",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0036 and AJ-0060 done",
      "refs": [
        "AJ-0036",
        "AJ-0060"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0036",
          "phase": "outputs",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · a0210e3 · 2026-09-29 · tests passed"
        },
        {
          "id": "AJ-0060",
          "phase": "outputs",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 3ae5531 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790706850
    },
    {
      "hash": "3ae5531716610ec103b17e7c24e4909bbed44123",
      "short": "3ae5531",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Refill filter menus on live refresh",
      "refs": [
        "AJ-0060"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790706832
    },
    {
      "hash": "7207ba63f5060b34d665cfa6ee724bd4bc39fcbf",
      "short": "7207ba6",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Dashboard live refresh under ajiya serve",
      "refs": [
        "AJ-0060"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790706600
    },
    {
      "hash": "a0210e3923b9d5790ba827ff3e8c73b4a5e2d143",
      "short": "a0210e3",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Dependencies view in the dashboard",
      "refs": [
        "AJ-0036"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790691355
    },
    {
      "hash": "61d1f5477f38e4fa2a5ce359e8a453209b815dd8",
      "short": "61d1f54",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Milestones v0-1, v0-2, v0-3 for this repository",
      "refs": [
        "AJ-0021"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790690491
    },
    {
      "hash": "3aa35723493ecf7b5c285dc6959eace07a3c5a74",
      "short": "3aa3572",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Plan dashboard live refresh; AJ-0021 waits for the dependencies view again",
      "refs": [
        "AJ-0021",
        "AJ-0036",
        "AJ-0060"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0036",
          "phase": "outputs",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        },
        {
          "id": "AJ-0060",
          "phase": "outputs",
          "from": "",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        }
      ],
      "time": 1790690441
    },
    {
      "hash": "1a7db969d60d9bf9b1f7eb4b637c63fc02d224cc",
      "short": "1a7db96",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0019 and AJ-0020 done",
      "refs": [
        "AJ-0019",
        "AJ-0020"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0019",
          "phase": "outputs",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 994ad17 · 2026-09-29 · tests passed"
        },
        {
          "id": "AJ-0020",
          "phase": "outputs",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 28b8618 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790690406
    },
    {
      "hash": "d7460c014521727273a1116bdeefc70c40581f10",
      "short": "d7460c0",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-acd57cf3aab521d30'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790690389
    },
    {
      "hash": "994ad178f88893af8326629994fda168f9b4e3a3",
      "short": "994ad17",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Dashboard: overview by milestone, next up, board, phases, tickets, apps, checks and activity",
      "refs": [
        "AJ-0019"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790690340
    },
    {
      "hash": "b7778308492f2ffe099418b70aee10fd689e7d6e",
      "short": "b777830",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a236161fb3f4ab13d'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790690327
    },
    {
      "hash": "28b8618e51bd805c183763b88d54a82c0bfa877f",
      "short": "28b8618",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Add ajiya serve: dashboard on 127.0.0.1 with Host check and rebuild on change",
      "refs": [
        "AJ-0020"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790689552
    },
    {
      "hash": "9d3407229e86fe3001c36741af1fc86f6d514091",
      "short": "9d34072",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0059 done",
      "refs": [
        "AJ-0059"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0059",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 57776dc · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790670902
    },
    {
      "hash": "57776dcc4dd367fcca69ecf281808a2b6955830f",
      "short": "57776dc",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "phase rename updates milestone targets",
      "refs": [
        "AJ-0059"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790670894
    },
    {
      "hash": "4ddeb1f1dd8573be8619e23916a291f93c1ec3ed",
      "short": "4ddeb1f",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Start AJ-0019, AJ-0020 and AJ-0059",
      "refs": [
        "AJ-0019",
        "AJ-0020",
        "AJ-0059"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0019",
          "phase": "outputs",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        },
        {
          "id": "AJ-0020",
          "phase": "outputs",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        },
        {
          "id": "AJ-0059",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790670796
    },
    {
      "hash": "29d61fc33dc2e1a2c7d223ca7b7d2ab8863facdb",
      "short": "29d61fc",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Embed the dashboard page; build writes ajiya/index.html",
      "refs": [
        "AJ-0019",
        "AJ-0059"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0059",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        }
      ],
      "time": 1790670795
    },
    {
      "hash": "2eae85df7d5635a6aa25c5106c9d872047be4cfd",
      "short": "2eae85d",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0018 done",
      "refs": [
        "AJ-0018"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0018",
          "phase": "outputs",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 7b6a1fb · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790670718
    },
    {
      "hash": "7b6a1fbc94bfe6b777f849a86a2635f92c2b88f0",
      "short": "7b6a1fb",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "ajiya build lists phases in display order",
      "refs": [
        "AJ-0018"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790670699
    },
    {
      "hash": "7d048ab11d1462f9592d90f729a646a6a8ac615a",
      "short": "7d048ab",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0040 done",
      "refs": [
        "AJ-0040"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0040",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 999e8f1 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790670662
    },
    {
      "hash": "a939eb50ad3d0f9e7ad1d9ed935bfda276816803",
      "short": "a939eb5",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a2dda545491c21789'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790670658
    },
    {
      "hash": "2d78fb09f4747033768f3f0f38399d58247560b0",
      "short": "2d78fb0",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "v0.3 spec update: keep ticket context light",
      "refs": [
        "AJ-0048",
        "AJ-0053",
        "AJ-0056"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790670622
    },
    {
      "hash": "999e8f1803e7e81d6b67d3f47fed34432cd19550",
      "short": "999e8f1",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Phase display order",
      "refs": [
        "AJ-0040"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790670603
    },
    {
      "hash": "4cdb51839c3a5d48ff0d14e3bf64a896f4214360",
      "short": "4cdb518",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0038 done",
      "refs": [
        "AJ-0038"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0038",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · a36d139 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790670579
    },
    {
      "hash": "04df5bfdf5a1fd211eb751090993e9d1f81289ec",
      "short": "04df5bf",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a2ad244a3eb91f538'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790670505
    },
    {
      "hash": "899acd5b84ee9ce71f792a5c75739248aee72cfb",
      "short": "899acd5",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "ajiya build: PROGRESS.md and data.js",
      "refs": [
        "AJ-0018"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0018",
          "phase": "outputs",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790670505
    },
    {
      "hash": "37cbb474ceaa89d4f1636370a8d27dbb7c94efb2",
      "short": "37cbb47",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Edit [[milestones]] tables in ajiya.toml, keeping comments",
      "refs": [
        "AJ-0038"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790670230
    },
    {
      "hash": "a36d139820d1f46799b55aa1929778a4b77919cf",
      "short": "a36d139",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Milestone commands and next by milestone",
      "refs": [
        "AJ-0038"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790670230
    },
    {
      "hash": "9afaf59a5e40c06bb0be36e33ff43e08f7a05369",
      "short": "9afaf59",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0039 done",
      "refs": [
        "AJ-0039"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0039",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 1828a44 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790669526
    },
    {
      "hash": "bf06c27cf371187d485211943c606a106c64680f",
      "short": "bf06c27",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-ac8546a83ced264ba'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790669519
    },
    {
      "hash": "1828a4447ddbe6e4c4dce7aaf210e1a9da56c7fc",
      "short": "1828a44",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Milestone checks: E012 per milestone, W009, W010",
      "refs": [
        "AJ-0039"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790669492
    },
    {
      "hash": "4892da452015c820eebe2232642eb5fe88c3402d",
      "short": "4892da4",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0037 done",
      "refs": [
        "AJ-0037"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0037",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 8a327fa · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790669238
    },
    {
      "hash": "e9b12ad458117acc0b9446ede5d223b61dfc853b",
      "short": "e9b12ad",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Start AJ-0038 to AJ-0040 in parallel",
      "refs": [
        "AJ-0038",
        "AJ-0039",
        "AJ-0040"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0038",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        },
        {
          "id": "AJ-0039",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        },
        {
          "id": "AJ-0040",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        }
      ],
      "time": 1790669238
    },
    {
      "hash": "8a327fa27dcb4022d2f17782c07751df4683d067",
      "short": "8a327fa",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Milestone config and model",
      "refs": [
        "AJ-0037"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0037",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790669234
    },
    {
      "hash": "8c39354299dc3760994832adfcaba7895167507c",
      "short": "8c39354",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Plan the roadmap: rest of v0.1, v0.2 and v0.3",
      "refs": [],
      "chore": true,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0037",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0038",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0039",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0040",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0041",
          "phase": "onboarding",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0042",
          "phase": "agent-access",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0043",
          "phase": "agent-access",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0044",
          "phase": "agent-access",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0045",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0046",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0047",
          "phase": "install",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending · Needs a human: final sign-off before publishing v0.2"
        },
        {
          "id": "AJ-0048",
          "phase": "decisions-and-changes",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0049",
          "phase": "decisions-and-changes",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0050",
          "phase": "decisions-and-changes",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0051",
          "phase": "decisions-and-changes",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0052",
          "phase": "decisions-and-changes",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0053",
          "phase": "decisions-and-changes",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0054",
          "phase": "decisions-and-changes",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0055",
          "phase": "decisions-and-changes",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0056",
          "phase": "decisions-and-changes",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0057",
          "phase": "decisions-and-changes",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0058",
          "phase": "decisions-and-changes",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending · Needs a human: final sign-off before publishing v0.3"
        }
      ],
      "time": 1790668900
    },
    {
      "hash": "0e258c373b2521f8093fcfbfd2ceee37141fd436",
      "short": "0e258c3",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Plan the dependencies view",
      "refs": [
        "AJ-0019",
        "AJ-0021",
        "AJ-0036"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0036",
          "phase": "outputs",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        }
      ],
      "time": 1790668673
    },
    {
      "hash": "6ad11f0fe49db3cd4062e614d8d3b243f875bb59",
      "short": "6ad11f0",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0028 done",
      "refs": [
        "AJ-0028"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0028",
          "phase": "onboarding",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 3b606af · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790635217
    },
    {
      "hash": "12bd215689b920dc9f1bb879950a1ac1461b49ca",
      "short": "12bd215",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a728b082f65808850'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790635162
    },
    {
      "hash": "6626d8e85dc24e96a741a67caed8daa09559c591",
      "short": "6626d8e",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0017 done",
      "refs": [
        "AJ-0017"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0017",
          "phase": "outputs",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 8c2dea8 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790635162
    },
    {
      "hash": "3b606af3c4c04d3dd580d79c0da8c74afd00385c",
      "short": "3b606af",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "ajiya import legacy: read the collate.py WBS format",
      "refs": [
        "AJ-0028"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790635042
    },
    {
      "hash": "4dd5a1485e0b2dbb9954acb0bd5ace4204da026f",
      "short": "4dd5a14",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-ad61081c23c6db22b'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790634828
    },
    {
      "hash": "2b36aa2edbcf8ac50e89be79b99034701e15eac0",
      "short": "2b36aa2",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0032 done",
      "refs": [
        "AJ-0032"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0032",
          "phase": "release",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · d84543c · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790634712
    },
    {
      "hash": "e01fe5d52f4baaa5486390d22e0001902b20f865",
      "short": "e01fe5d",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a4a414bf5472fec0e'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790634702
    },
    {
      "hash": "8c2dea89736b5f50698bbf2f88ee0351a385f3e0",
      "short": "8c2dea8",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Add activity feed read from git history",
      "refs": [
        "AJ-0017"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790634688
    },
    {
      "hash": "d84543cb56c00a2f51010ae4a1dae32573bab1d4",
      "short": "d84543c",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Add 'ajiya changelog': commits grouped by phase and ticket",
      "refs": [
        "AJ-0032"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790634676
    },
    {
      "hash": "2eef7c9abb06be5786733fb00f25514cdb1b0aee",
      "short": "2eef7c9",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0027 done",
      "refs": [
        "AJ-0027"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0027",
          "phase": "onboarding",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · da34caf · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790634664
    },
    {
      "hash": "392e452d15e8d3d6c113405d4a9299963168cf31",
      "short": "392e452",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a97739c83a4b8bd21'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790634661
    },
    {
      "hash": "c4f8fb3cc837984484a37feade45a65a91ec34e5",
      "short": "c4f8fb3",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0026 done",
      "refs": [
        "AJ-0026"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0026",
          "phase": "onboarding",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 5cf99f1 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790634641
    },
    {
      "hash": "d4be704082cbb89ebd178d09f71de73134ea459d",
      "short": "d4be704",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a6ded39c29dcd9bfb'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790634632
    },
    {
      "hash": "da34caf635bb88e05b6bd89c9c856524df6b1542",
      "short": "da34caf",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "ajiya import github: import issues through gh into the inbox",
      "refs": [
        "AJ-0027"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790634626
    },
    {
      "hash": "5cf99f147f3a05b4acbc19c4de3dbda4cb4472f1",
      "short": "5cf99f1",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Add ajiya import todo: markdown list items into the inbox phase",
      "refs": [
        "AJ-0026"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790634615
    },
    {
      "hash": "d892d73d6b1fc68159840afa802dcf5e1cceb8d5",
      "short": "d892d73",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0035 done",
      "refs": [
        "AJ-0035"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0035",
          "phase": "outputs",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 145c34f · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790634613
    },
    {
      "hash": "fd56edf146762bcd5e46f2cc330f980351a25484",
      "short": "fd56edf",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Merge branch 'worktree-agent-a188dc52edb942989'",
      "refs": [],
      "chore": false,
      "merge": true,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790634604
    },
    {
      "hash": "145c34fc9656a3c982d38f8402da404003940f20",
      "short": "145c34f",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Add ajiya phase rename",
      "refs": [
        "AJ-0035"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790634560
    },
    {
      "hash": "71e1a1dd4a7efb82f96a1df8451c603b45649c63",
      "short": "71e1a1d",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Start six tickets in parallel",
      "refs": [],
      "chore": true,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0017",
          "phase": "outputs",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        },
        {
          "id": "AJ-0026",
          "phase": "onboarding",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        },
        {
          "id": "AJ-0027",
          "phase": "onboarding",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        },
        {
          "id": "AJ-0028",
          "phase": "onboarding",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        },
        {
          "id": "AJ-0032",
          "phase": "release",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        },
        {
          "id": "AJ-0035",
          "phase": "outputs",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: in a parallel worktree"
        }
      ],
      "time": 1790634384
    },
    {
      "hash": "e6fa03476058339d2fc7c938a2d24f6abb28e24f",
      "short": "e6fa034",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Ticket model for imports: issue and before-Ajiya evidence, links, aliases",
      "refs": [
        "AJ-0026",
        "AJ-0027",
        "AJ-0028"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790634382
    },
    {
      "hash": "75087a497a1630a51afe2f9adfea45f2c7ec8581",
      "short": "75087a4",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0014 done",
      "refs": [
        "AJ-0014"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0014",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · e2b8505 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790634110
    },
    {
      "hash": "e2b8505d7fc9c09f623f201cbb36958b2f8c5d71",
      "short": "e2b8505",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "--json on every read command",
      "refs": [
        "AJ-0014"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0014",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790634107
    },
    {
      "hash": "8ec110f3c4002cdf92f841c3708842b7e7ba6156",
      "short": "8ec110f",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0008 done",
      "refs": [
        "AJ-0008"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0008",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 83f20ca · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790633981
    },
    {
      "hash": "83f20ca34eebd39e8424e52c3449603b91df4a21",
      "short": "83f20ca",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "App detection in init",
      "refs": [
        "AJ-0008"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0008",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790633978
    },
    {
      "hash": "6a7c515786d0f04d772a2770ac077a1a1bdb1c19",
      "short": "6a7c515",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0013 done",
      "refs": [
        "AJ-0013"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0013",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · b3d5792 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790633818
    },
    {
      "hash": "b3d5792f273b20fecb5ced4cc5e6d21acc812542",
      "short": "b3d5792",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Warning checks W001-W008 and --strict",
      "refs": [
        "AJ-0013"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0013",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790633815
    },
    {
      "hash": "401b31e12e828e3e1e2e1d132b53e8ad67eb311a",
      "short": "401b31e",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0012 done",
      "refs": [
        "AJ-0012"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0012",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 7bac540 · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790633541
    },
    {
      "hash": "7bac5403db16c08d333d03d094c55e02a16cf0c7",
      "short": "7bac540",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Error checks for apps, evidence and the launch target",
      "refs": [
        "AJ-0012"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0012",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790633539
    },
    {
      "hash": "c0ca509777a6589aac787397944bc2bf13848800",
      "short": "c0ca509",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0015 done",
      "refs": [
        "AJ-0015"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0015",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · c0239ef · 2026-09-29 · tests passed"
        }
      ],
      "time": 1790633506
    },
    {
      "hash": "c0239ef016e396ab618cc413a1d1e4ac0bc84301",
      "short": "c0239ef",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "ajiya ticket done --test",
      "refs": [
        "AJ-0015"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0015",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790633503
    },
    {
      "hash": "03297289de6e83927056e4ccbbb20e8be498a69f",
      "short": "0329728",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0016 done",
      "refs": [
        "AJ-0016"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0016",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 1ad6b45 · 2026-09-29"
        }
      ],
      "time": 1790633466
    },
    {
      "hash": "1ad6b45ed195f6e5a78e58b15f6132b99d4f1959",
      "short": "1ad6b45",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "ajiya ticket block and ticket drop",
      "refs": [
        "AJ-0016"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0016",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790633465
    },
    {
      "hash": "21bcd17ccbc1ed48ce3b9bf0818b7bf7a66cb96a",
      "short": "21bcd17",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0009 done",
      "refs": [
        "AJ-0009"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0009",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 57e4454 · 2026-09-29"
        }
      ],
      "time": 1790633423
    },
    {
      "hash": "57e44540e563a4db760a52074879922e67f8fcb9",
      "short": "57e4454",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "ajiya app add, move and remove",
      "refs": [
        "AJ-0009"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0009",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790633422
    },
    {
      "hash": "f5a5aa2552c82f061d6525550ca3d0e145ddd4c4",
      "short": "f5a5aa2",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0011 done",
      "refs": [
        "AJ-0011"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0011",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · d32c066 · 2026-09-29"
        }
      ],
      "time": 1790633231
    },
    {
      "hash": "d32c0667e684382af10f9d3d5cbea02930186dfb",
      "short": "d32c066",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "ajiya next lists launch-required tickets first",
      "refs": [
        "AJ-0011"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0011",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790633230
    },
    {
      "hash": "802b01317bfc21c0db2518b43011ed014c819df9",
      "short": "802b013",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0010 done",
      "refs": [
        "AJ-0010"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0010",
          "phase": "apps-launch",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 3349f60 · 2026-09-29"
        }
      ],
      "time": 1790633197
    },
    {
      "hash": "3349f60eb586db783e313acce2bc2f458f0ddff3",
      "short": "3349f60",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "ajiya launch set and show",
      "refs": [
        "AJ-0010"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0010",
          "phase": "apps-launch",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790633196
    },
    {
      "hash": "c2c4fa4d734e02990c7e5ee38167880a407dddb1",
      "short": "c2c4fa4",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0006 done",
      "refs": [
        "AJ-0006"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0006",
          "phase": "commit-rule",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 5c94eca · 2026-09-29 · Note: hooks installed here; unknown-ID commit refused"
        }
      ],
      "time": 1790633052
    },
    {
      "hash": "2f9d28a8e40b7f6fea7d3ba900e17ea33eedee08",
      "short": "2f9d28a",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "ticket show lists referencing commits",
      "refs": [
        "AJ-0007"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0007",
          "phase": "commit-rule",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790632959
    },
    {
      "hash": "6e080164a4370b46efff2a6e6a4adbc4766a4d6f",
      "short": "6e08016",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0007 done",
      "refs": [
        "AJ-0007"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0007",
          "phase": "commit-rule",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 2f9d28a · 2026-09-29"
        }
      ],
      "time": 1790632959
    },
    {
      "hash": "5c94eca9e638472d1e8d3d8bc6bca2a671010192",
      "short": "5c94eca",
      "date": "2026-09-29",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Run the test editor through sh so Windows paths work",
      "refs": [
        "AJ-0006"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790632862
    },
    {
      "hash": "f8d4b01d75aa56b51424b25f75baaa97c2e6c372",
      "short": "f8d4b01",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "CI: check the commit rule on every push",
      "refs": [
        "AJ-0006"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0006",
          "phase": "commit-rule",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: CI job done; hooks need ajiya on PATH"
        }
      ],
      "time": 1790632727
    },
    {
      "hash": "58e9736015104a7ab5e9fe3461f46e43a2bf608b",
      "short": "58e9736",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "ajiya hook install",
      "refs": [
        "AJ-0005"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0005",
          "phase": "commit-rule",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790632700
    },
    {
      "hash": "800c2260f43485cd9f407ebcb0feff80d2b56d8f",
      "short": "800c226",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0005 done",
      "refs": [
        "AJ-0005"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0005",
          "phase": "commit-rule",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 58e9736 · 2026-09-28"
        }
      ],
      "time": 1790632700
    },
    {
      "hash": "7ad90ad6ad2a97129341d0324918eb37b4d76253",
      "short": "7ad90ad",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0004 done",
      "refs": [
        "AJ-0004"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0004",
          "phase": "commit-rule",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 6f015d8 · 2026-09-28"
        }
      ],
      "time": 1790632586
    },
    {
      "hash": "6f015d82c5b418dfaffb198d3b0b36fc14542e57",
      "short": "6f015d8",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "prepare-commit-msg hook",
      "refs": [
        "AJ-0004"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0004",
          "phase": "commit-rule",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790632585
    },
    {
      "hash": "bc3bdbea64e36d884fd727b45d9c9744f1811344",
      "short": "bc3bdbe",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0003 done",
      "refs": [
        "AJ-0003"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0003",
          "phase": "commit-rule",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · ea005bf · 2026-09-28"
        }
      ],
      "time": 1790632504
    },
    {
      "hash": "ea005bf26d08721af18152b9cf32da2c7a0f97c7",
      "short": "ea005bf",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "commit-msg hook",
      "refs": [
        "AJ-0003"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0003",
          "phase": "commit-rule",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790632503
    },
    {
      "hash": "39cb0929f312d2b0c392f17475023c9fc7ba85a6",
      "short": "39cb092",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Reference: real SKILL.md and AGENTS.md templates",
      "refs": [
        "AJ-0024"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790632365
    },
    {
      "hash": "a5019a2a2887e5548c72b100d870198b90c70df9",
      "short": "a5019a2",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0002 done",
      "refs": [
        "AJ-0002"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0002",
          "phase": "commit-rule",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 4a49882 · 2026-09-28"
        }
      ],
      "time": 1790632305
    },
    {
      "hash": "4a49882b57e35fd3b754f273f6f6bae8f595e88b",
      "short": "4a49882",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "ajiya check --commits \u003crange\u003e",
      "refs": [
        "AJ-0002"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0002",
          "phase": "commit-rule",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress"
        }
      ],
      "time": 1790632304
    },
    {
      "hash": "3e3674caa1f9ae5be1f93527da484c90efc4dc3b",
      "short": "3e3674c",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "AJ-0001 done",
      "refs": [
        "AJ-0001"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0001",
          "phase": "commit-rule",
          "from": "in_progress",
          "to": "done",
          "text": "🟩 Done · 4cc2b3f · 2026-09-28"
        }
      ],
      "time": 1790632211
    },
    {
      "hash": "4cc2b3fe50c9a6fad1eba45a09a0fcf2e2c3a22c",
      "short": "4cc2b3f",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Parse Ajiya trailers with git's own rules",
      "refs": [
        "AJ-0001"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0001",
          "phase": "commit-rule",
          "from": "pending",
          "to": "in_progress",
          "text": "🟨 In progress: message parsing, exemptions, agreement tests"
        }
      ],
      "time": 1790632210
    },
    {
      "hash": "125cbc47f8ec7f6420fb1c55f056a1b4a7d109ee",
      "short": "125cbc4",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Plan: launch target is the v0-1 phase",
      "refs": [
        "AJ-0035",
        "AJ-0021",
        "AJ-0034"
      ],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0035",
          "phase": "outputs",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        }
      ],
      "time": 1790632075
    },
    {
      "hash": "f9ece851adce684711adbae45ae2489d11272367",
      "short": "f9ece85",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Add the ajiya skill to the repository",
      "refs": [],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790631833
    },
    {
      "hash": "0db8b0d5abc5f8180c12372ec628dfeba8ecc967",
      "short": "0db8b0d",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "M1 done: Ajiya now tracks its own work",
      "refs": [],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [
        {
          "id": "AJ-0001",
          "phase": "commit-rule",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0002",
          "phase": "commit-rule",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0003",
          "phase": "commit-rule",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0004",
          "phase": "commit-rule",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0005",
          "phase": "commit-rule",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0006",
          "phase": "commit-rule",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0007",
          "phase": "commit-rule",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0008",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0009",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0010",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0011",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0012",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0013",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0014",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0015",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0016",
          "phase": "apps-launch",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0017",
          "phase": "outputs",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0018",
          "phase": "outputs",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0019",
          "phase": "outputs",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0020",
          "phase": "outputs",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0021",
          "phase": "outputs",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0022",
          "phase": "onboarding",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0023",
          "phase": "onboarding",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0024",
          "phase": "onboarding",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0025",
          "phase": "onboarding",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0026",
          "phase": "onboarding",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0027",
          "phase": "onboarding",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0028",
          "phase": "onboarding",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0029",
          "phase": "onboarding",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0030",
          "phase": "release",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0031",
          "phase": "release",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0032",
          "phase": "release",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0033",
          "phase": "release",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending"
        },
        {
          "id": "AJ-0034",
          "phase": "release",
          "from": "",
          "to": "pending",
          "text": "🟥 Pending · Needs a human: final sign-off before publishing"
        }
      ],
      "time": 1790631806
    },
    {
      "hash": "9bf8ffb79cc4d3be121a1fea057df83732edad16",
      "short": "9bf8ffb",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "M1: commands, error checks and CLI tests",
      "refs": [],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790631705
    },
    {
      "hash": "0587829ba1692a70dbc7eec398e8bdef0b34f973",
      "short": "0587829",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Apply research: fsync before rename, refine plan items",
      "refs": [],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790631371
    },
    {
      "hash": "78e5ceeb60c87738fa5833e704cc7f696447fcf6",
      "short": "78e5cee",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "M1: config loading, phase file reader and writer, dependency graph",
      "refs": [],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790631288
    },
    {
      "hash": "4ea5d42c178bac013aca4cde19e2559ac7b3ff5a",
      "short": "4ea5d42",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Add sanitised reference material",
      "refs": [],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790630951
    },
    {
      "hash": "16548d77d9650f75b748f88c0501466ef647be88",
      "short": "16548d7",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "M0: project skeleton",
      "refs": [],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "Claude",
      "changes": [],
      "time": 1790630482
    },
    {
      "hash": "4f97764d49da87c93c4f32f5b82d00f56cc4ce7e",
      "short": "4f97764",
      "date": "2026-09-28",
      "author": "Aliyu Yahaya - ST10440112",
      "subject": "Chore: match module path to repo name, update CI actions",
      "refs": [],
      "chore": false,
      "merge": false,
      "revert": false,
      "agent": "",
      "changes": [],
      "time": 1790630482
    }
  ]
};
