<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Outputs

Goal: PROGRESS.md, data.js, the dashboard and serve show the plan and its activity

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| AJ-0017 | ajiya | Activity feed from git | Last 30 days of commits with tickets, status changes from diffs and agent co-authors; deterministic | AJ-0001 | 🟩 Done · 8c2dea8 · 2026-09-29 · tests passed |
| AJ-0018 | ajiya | ajiya build | Writes PROGRESS.md and data.js with a schema version, only when content changes, dated by the newest source change | AJ-0013, AJ-0017, AJ-0038, AJ-0040 | 🟩 Done · 7b6a1fb · 2026-09-29 · tests passed |
| AJ-0019 | ajiya | Dashboard | Adapted from reference/index.html for phases and tickets, with no RO work packages or MedInformer values: Overview with progress per milestone in order, Next up, Board, Phases, Tickets, Apps, Checks and Activity views, phases in display order, CSS variable themes for light and dark, embedded in the binary. The Dependencies view is its own ticket | AJ-0018, AJ-0038 | 🟨 In progress: in a parallel worktree |
| AJ-0020 | ajiya | ajiya serve | Serves on 127.0.0.1 only, answers 403 to other Host headers, rebuilds when files change | AJ-0019 | 🟨 In progress: in a parallel worktree |
| AJ-0021 | ajiya | Milestones v0-1, v0-2 and v0-3 set and the dashboard in use on this repo | Milestones v0-1, v0-2 and v0-3 target the three version sign-off tickets; ajiya milestone list shows them in order; ajiya serve shows this repo's plan | AJ-0020, AJ-0038 | 🟥 Pending |
| AJ-0035 | ajiya | ajiya phase rename | ajiya phase rename <old> <new> renames the file and moves every ticket; refused if the new slug exists; the launch target follows the rename | - | 🟩 Done · 145c34f · 2026-09-29 · tests passed |
| AJ-0036 | ajiya | Dependencies view in the dashboard | Phase graph by default, with an arrow where a ticket in one phase depends on a ticket in another; clicking a phase opens its ticket graph; Path to a milestone, with a milestone picker; hover traces a chain; stays readable with 500 tickets | AJ-0019 | 🟥 Pending |
