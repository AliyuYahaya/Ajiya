---
name: ajiya
description: Plans, tracks and proves work in this repository with the ajiya CLI. Picks the next ticket, starts it, commits with the Ajiya trailer and marks it done with evidence. Use before starting any task in this repository, when deciding what to work on, before committing, and when work is finished.
---

# Ajiya

This repository tracks its work with Ajiya. The plan is in `ajiya/`, one markdown
file per phase, and changes only through `ajiya` commands. Run any command with
`--help` for its arguments.

## Rules

1. Never edit `ajiya/` by hand: use `ajiya` commands (in `ajiya.toml`, only `[test]` is set by hand).
2. Before a task: `ajiya next`, then `ajiya ticket start <ID>`. No ticket? Add one first.
3. Commit with a last paragraph `Ajiya: <ID>` (1 to 3 IDs, or `Ajiya: chore`).
4. Finish with `ajiya ticket done <ID> --test`, then `ajiya check`. Never mark done any other way.
5. Found extra work? `ajiya ticket add`; do not widen the current ticket.
6. Tags are optional. Run `ajiya tags` first and reuse values. Add `--tech` only when
   a ticket depends on a specific technology or service (`tech:aws-rds` on
   "Configure RDS backups", none on "Add a booking form"). Use `--paths` only
   while a ticket has no commits: after that, `ajiya ticket files <ID>` reads git.
   Never store what the CLI can work out (files, commits, readiness, milestones).
7. A choice the plan rests on? `ajiya decision add`. A person accepts it
   (`ajiya decision accept --by`). Link tickets with `--decisions`.
8. A choice is being replaced? Use a change request, never edit the old tickets:
   `change open`, `change impact`, `change assess` every candidate, `change show`,
   commit, then wait. **Never run `ajiya change approve` or `ajiya change reject`:**
   only a person does.

## Guides

- The loop for every task, tags, decisions and the change procedure: `.ajiya/guide/daily.md`
- Planning a project or organising imported tickets: `.ajiya/guide/setup.md`
