# Build Ajiya to v0.3, using Ajiya

You are continuing the build of Ajiya in this repository. This file is the plan
for the next three versions. Three specs sit next to it:

| File | Version | Contents |
|---|---|---|
| `01-v0.1-milestones-and-rollout-import.md` | v0.1 (remaining) | milestones, phase display order, RO packages imported as phases |
| `02-v0.2-install-and-mcp.md` | v0.2 | `ajiya status`, the MCP server, registering with agents, install channels, the Claude Code plugin |
| `03-v0.3-decisions-and-changes.md` | v0.3 | ticket tags, decisions, the superseded status, change requests |

The original build prompt (`ajiya-build-prompt.md`) still holds: its decisions,
its "ask me before" list and its rules for working. Where a spec here adds to it,
the spec wins; where they conflict, stop and ask me.

## The idea

Ajiya exists so that a whole plan can live in the repository and an agent can
work through it one proven ticket at a time, picking up in any new session from
`ajiya next`. Build Ajiya's own roadmap that way. The plan goes into Ajiya first;
the code follows it.

## Step 1: load the plan (no code yet)

1. If `git status` shows changes in `ajiya/`, commit them with the right trailer
   (`ajiya/outputs.md` currently holds changes to AJ-0019, AJ-0021 and AJ-0036).
2. Read the three specs in full.
3. Add every ticket from spec 01 to the existing plan, with dependencies. Then
   `ajiya ticket edit AJ-0034 --depends ...` so the v0.1 sign-off depends on them
   too. v0.1 is not done until spec 01 is.
4. Add spec 02 as phases and tickets. Add a v0.2 sign-off ticket marked
   `--human "final sign-off before publishing v0.2"`, depending on every v0.2
   ticket. Make the first v0.2 tickets depend on AJ-0034, so `ajiya next` never
   offers v0.2 work before v0.1 is signed off.
5. Add spec 03 the same way, depending on the v0.2 sign-off ticket, with its own
   v0.3 sign-off ticket.
6. Plan later versions more loosely than v0.1. One ticket per spec part is
   enough for v0.2 and v0.3; split them when you reach them. They will change
   once people use v0.1, and that is expected.
7. Run `ajiya check --strict`. Fix every finding.
8. Commit the plan (`Ajiya: chore` is fine for a plan-only commit if more than
   three tickets changed).
9. **Stop and report to me**: phases and ticket counts per version, the order
   `ajiya next` will take, anything in the specs you think is wrong or unclear,
   and any dependency you had to guess. Wait for my go-ahead.

## Step 2: build, one version at a time

Work the loop the skill describes: `ajiya next`, `ticket start`, tests, code,
commit with the trailer, `ticket done --test`, `ajiya check`.

- Start every session with `ajiya next` (and `ajiya status` once it exists).
  Trust the plan over your memory of earlier sessions.
- **When milestones exist** (spec 01), set them up for this repository:
  `v0-1`, `v0-2` and `v0-3`, each targeting its version's sign-off ticket. Keep
  the dependencies on the sign-off tickets: they are real (a version must be
  signed off before the next starts).
- **When the MCP server exists** (spec 02), register it for this repository and
  use its tools in later sessions where your client supports them. This is the
  best test v0.2 will get.
- Found extra work? Add a ticket in the right version. Never widen the current
  ticket.
- A spec turns out wrong or unworkable? Stop and tell me before working around
  it. For v0.2 and v0.3 tickets, you may rewrite a ticket's title or "done when"
  as you learn, but list every such edit in the next report.
- Parallel work in separate worktrees is fine. Each worktree follows the same
  rules, and every merge keeps the trailers.

## Step 3: stop at every sign-off

When `ajiya next` offers a sign-off ticket (AJ-0034, then the v0.2 and v0.3
ones), do not start the next version. Report instead:

- what the version contains, in a few lines per phase
- `go test ./...`, `ajiya check --strict` and `ajiya check --commits` results
- the definition of done from the spec, item by item: met, not met, or needs me
- every ticket you edited or added that was not in the spec
- open questions

Then wait. I close the sign-off ticket myself.

## Never, without asking me

Everything in the build prompt's "ask me before" list still applies. In short:
- publish anything: releases, npm, Homebrew, the install script URL, the plugin
- add dependencies beyond the TOML parser and test helpers (the MCP SDK included)
- send data off the machine
- rewrite git history
- commit anything from `reference/private/`, copy it into fixtures or quote it
- change the decisions listed at the end of each spec
