<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Outputs

Goal: PROGRESS.md, data.js, the dashboard and serve show the plan and its activity

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| AJ-0017 | ajiya | Activity feed from git | Last 30 days of commits with tickets, status changes from diffs and agent co-authors; deterministic | AJ-0001 | 🟩 Done · 8c2dea8 · 2026-09-29 · tests passed |
| AJ-0018 | ajiya | ajiya build | Writes PROGRESS.md and data.js with a schema version, only when content changes, dated by the newest source change | AJ-0013, AJ-0017 | 🟥 Pending |
| AJ-0019 | ajiya | Dashboard | Adapted from reference/index.html for phases and tickets, with no RO work packages or MedInformer values: Overview, Next up, Board, Phases, Tickets, Apps, Checks and Activity views, CSS variable themes for light and dark, embedded in the binary. The Dependencies view is its own ticket | AJ-0018 | 🟥 Pending |
| AJ-0020 | ajiya | ajiya serve | Serves on 127.0.0.1 only, answers 403 to other Host headers, rebuilds when files change | AJ-0019 | 🟥 Pending |
| AJ-0021 | ajiya | Launch target v0-1 set and dashboard in use on this repo | The release phase is renamed v0-1, ajiya launch show names it, and ajiya serve shows this repo's plan | AJ-0020, AJ-0035, AJ-0036 | 🟥 Pending |
| AJ-0035 | ajiya | ajiya phase rename | ajiya phase rename <old> <new> renames the file and moves every ticket; refused if the new slug exists; the launch target follows the rename | - | 🟩 Done · 145c34f · 2026-09-29 · tests passed |
| AJ-0036 | ajiya | Dependencies view in the dashboard | Phase graph by default, with an arrow where a ticket in one phase depends on a ticket in another; clicking a phase opens its ticket graph; Path to launch uses the launch target from ajiya.toml; hover traces a chain; stays readable with 500 tickets | AJ-0019 | 🟥 Pending |
