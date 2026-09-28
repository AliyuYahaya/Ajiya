---
name: ajiya
description: Plan, track and prove work in this repository with the ajiya CLI. Use before starting any task here, when picking what to do next, when committing, and when marking work done.
---

# Ajiya

This repository tracks its own work with Ajiya. The plan is in `ajiya/`, one file
per phase. Build the CLI with `go build -o ajiya ./cmd/ajiya` if it is not on PATH.

Rules:
1. Never edit `ajiya/` by hand. Change the plan only with `ajiya` commands.
2. Before writing code, find or add the ticket: `ajiya next`, or `ajiya ticket add`
   with `--depends` for what must come first.
3. `ajiya ticket start <ID>`, write the code and tests, then commit with the trailer
   `Ajiya: <ID>` (1 to 3 IDs, or `Ajiya: chore` for upkeep).
4. `ajiya ticket done <ID>` records the commit as evidence. Never mark work done any
   other way.
5. Run `go test ./...` and `ajiya check` before you finish. Extra work found on the
   way becomes a new ticket, not a wider current one.

Full guides arrive with the agent kit (tickets AJ-0022 to AJ-0025).
