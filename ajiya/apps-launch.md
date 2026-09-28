<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Apps-launch

Goal: Apps are detected and managed, launch readiness is defined, and every check in the spec runs

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| AJ-0008 | ajiya | App detection in init | Suggests apps from workspace files and manifests, ranked, skipping vendor folders; --yes accepts; fixture repos pass | - | 🟥 Pending |
| AJ-0009 | ajiya | ajiya app add, remove and move | Refuses to orphan open tickets unless --to is given; script tests pass | - | 🟩 Done · 57e4454 · 2026-09-29 |
| AJ-0010 | ajiya | ajiya launch set and show | Required set is the target's tickets plus everything they depend on; table tests pass | - | 🟩 Done · 3349f60 · 2026-09-29 |
| AJ-0011 | ajiya | ajiya next --launch | Launch-required tickets are listed first | AJ-0010 | 🟩 Done · d32c066 · 2026-09-29 |
| AJ-0012 | ajiya | Remaining error checks | Unregistered app, done without evidence and missing launch target each have a code and a fixture | AJ-0009, AJ-0010 | 🟨 In progress |
| AJ-0013 | ajiya | Warning checks and --strict | Every warning in spec section 4 has a W code and a fixture; --strict fails on warnings | AJ-0001, AJ-0012 | 🟥 Pending |
| AJ-0014 | ajiya | --json on every read command | Each read command prints stable JSON with --json; tests decode it | AJ-0013 | 🟥 Pending |
| AJ-0015 | ajiya | ajiya ticket done --test | Runs the configured test command, records tests passed, refuses on failure | - | 🟩 Done · c0239ef · 2026-09-29 · tests passed |
| AJ-0016 | ajiya | ajiya ticket block and ticket drop | block sets a reason; drop needs --reason and --by and counts as closed | - | 🟩 Done · 1ad6b45 · 2026-09-29 |
