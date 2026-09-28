<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Outputs

Goal: PROGRESS.md, data.js, the dashboard and serve show the plan and its activity

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| AJ-0017 | ajiya | Activity feed from git | Last 30 days of commits with tickets, status changes from diffs and agent co-authors; deterministic | AJ-0001 | 🟥 Pending |
| AJ-0018 | ajiya | ajiya build | Writes PROGRESS.md and data.js with a schema version, only when content changes, dated by the newest source change | AJ-0013, AJ-0017 | 🟥 Pending |
| AJ-0019 | ajiya | Dashboard | Adapted from reference/index.html with all views plus Activity, CSS variable themes for light and dark, embedded in the binary | AJ-0018 | 🟥 Pending |
| AJ-0020 | ajiya | ajiya serve | Serves on 127.0.0.1 only, answers 403 to other Host headers, rebuilds when files change | AJ-0019 | 🟥 Pending |
| AJ-0021 | ajiya | Launch target set and dashboard in use on this repo | ajiya launch show names the v0.1 target and ajiya serve shows this repo's plan | AJ-0020 | 🟥 Pending |
