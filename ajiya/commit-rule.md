<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Commit-rule

Goal: Every commit names 1 to 3 tickets, enforced by hooks with a CI check as backup

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| AJ-0001 | ajiya | Parse Ajiya trailers with git's own rules | Refs, chore, merge and revert exemptions agree between message files and history; table tests pass | - | 🟩 Done · 4cc2b3f · 2026-09-28 |
| AJ-0002 | ajiya | ajiya check --commits <range> | Reports commits with 0 or more than 3 references and unknown IDs; fixture per case | AJ-0001 | 🟩 Done · 4a49882 · 2026-09-28 |
| AJ-0003 | ajiya | commit-msg hook | Refuses a commit without a valid trailer; merges and reverts pass; script test in a temp repo | AJ-0001 | 🟩 Done · ea005bf · 2026-09-28 |
| AJ-0004 | ajiya | prepare-commit-msg hook | Pre-fills the trailer from in-progress tickets and changed ajiya/ rows; leaves merge, squash and amend alone | AJ-0001 | 🟩 Done · 6f015d8 · 2026-09-28 |
| AJ-0005 | ajiya | ajiya hook install | Respects core.hooksPath, appends a marked block to existing hooks, works in Git for Windows, clear message when ajiya is not on PATH | AJ-0003, AJ-0004 | 🟩 Done · 58e9736 · 2026-09-28 |
| AJ-0006 | ajiya | CI commit check and hooks on this repo | CI runs ajiya check --commits; hooks installed here and a commit without a trailer is refused | AJ-0002, AJ-0005 | 🟩 Done · 5c94eca · 2026-09-29 · Note: hooks installed here; unknown-ID commit refused |
| AJ-0007 | ajiya | ticket show lists referencing commits | ajiya ticket show prints hash, date and subject of each commit naming the ticket | AJ-0001 | 🟩 Done · 2f9d28a · 2026-09-29 |
| AJ-0065 | ajiya | ticket done takes evidence only from commits after the ticket started | ticket start records the start time locally (in the git directory); ticket done refuses when every commit naming the ticket predates it and names the ignored commit; tests cover a pre-start commit alone, a pre-start plus a later commit, and a ticket with no start record | - | 🟩 Done · 5c1af99 · 2026-09-29 · tests passed |
