<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Decisions-and-changes

Goal: Plans change without losing history: tags, decisions, superseded work and approved change requests

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| AJ-0048 | ajiya | Tags column | Seventh Tags column with tech, comp, dec and path tags written from flags; six-column files read and upgraded; ajiya tags; near-duplicate warning; round-trip goldens for both forms; MCP tools | AJ-0047 | 🟥 Pending |
| AJ-0049 | ajiya | ajiya ticket files | Lists every file changed by commits referencing a ticket, from git; CLI test and MCP tool | AJ-0048 | 🟥 Pending |
| AJ-0050 | ajiya | Decision records and commands | ajiya/decisions/DEC-xxxx.md written only by the CLI; decision add, show, list and accept with --json; ticket show lists linked decisions; checks for missing and superseded decisions; goldens and MCP tools | AJ-0048 | 🟥 Pending |
| AJ-0051 | ajiya | Superseded status | Superseded by <ID> · CHG · was: <old status> keeps the old evidence; excluded from totals and required sets; round-trip goldens | AJ-0048 | 🟥 Pending |
| AJ-0052 | ajiya | Snapshots as git tags | ajiya snapshot <name> tags ajiya/plan/<name>, refused with uncommitted ajiya/ changes; snapshot list; CLI tests | AJ-0051 | 🟥 Pending |
| AJ-0053 | ajiya | change open and change impact | Change files written only by the CLI with a snapshot; impact finds candidates by tag, decision, path and one level downstream with reasons, keeping verdicts on rerun; goldens and MCP tools | AJ-0049, AJ-0050, AJ-0051, AJ-0052 | 🟥 Pending |
| AJ-0054 | ajiya | change assess | Records unaffected, modify, replace, invalidate or investigate with a reason, and the new ticket text where needed; CLI tests and MCP tool | AJ-0053 | 🟥 Pending |
| AJ-0055 | ajiya | change approve and change reject | approve applies every verdict in one write (new tickets, edits, superseded, moved dependencies, superseded decisions) and records who and when; refused with missing or investigate verdicts; reject leaves the plan alone; not offered over MCP; the Supabase scenario test passes | AJ-0054 | 🟥 Pending |
| AJ-0056 | ajiya | Guides and skill for decisions and changes | Guides and skill say how to tag tickets, record decisions and run the change procedure, and that agents never approve or reject | AJ-0055 | 🟥 Pending |
| AJ-0057 | ajiya | Dashboard: open changes and decisions | Open changes with their impact table and a Decisions view in the dashboard | AJ-0019, AJ-0055 | 🟥 Pending |
| AJ-0058 | ajiya | v0.3 definition of done | The scenario test passes, and a person runs one real change through a sample repository from change open to approve and reviews the result | AJ-0048, AJ-0049, AJ-0050, AJ-0051, AJ-0052, AJ-0053, AJ-0054, AJ-0055, AJ-0056 | 🟥 Pending · Needs a human: final sign-off before publishing v0.3 |
