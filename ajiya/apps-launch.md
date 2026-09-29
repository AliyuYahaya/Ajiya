<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Apps-launch

Goal: Apps are detected and managed, launch readiness is defined, and every check in the spec runs

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| AJ-0008 | ajiya | App detection in init | Suggests apps from workspace files and manifests, ranked, skipping vendor folders; --yes accepts; fixture repos pass | - | 🟩 Done · 83f20ca · 2026-09-29 · tests passed |
| AJ-0009 | ajiya | ajiya app add, remove and move | Refuses to orphan open tickets unless --to is given; script tests pass | - | 🟩 Done · 57e4454 · 2026-09-29 |
| AJ-0010 | ajiya | ajiya launch set and show | Required set is the target's tickets plus everything they depend on; table tests pass | - | 🟩 Done · 3349f60 · 2026-09-29 |
| AJ-0011 | ajiya | ajiya next --launch | Launch-required tickets are listed first | AJ-0010 | 🟩 Done · d32c066 · 2026-09-29 |
| AJ-0012 | ajiya | Remaining error checks | Unregistered app, done without evidence and missing launch target each have a code and a fixture | AJ-0009, AJ-0010 | 🟩 Done · 7bac540 · 2026-09-29 · tests passed |
| AJ-0013 | ajiya | Warning checks and --strict | Every warning in spec section 4 has a W code and a fixture; --strict fails on warnings | AJ-0001, AJ-0012 | 🟩 Done · b3d5792 · 2026-09-29 · tests passed |
| AJ-0014 | ajiya | --json on every read command | Each read command prints stable JSON with --json; tests decode it | AJ-0013 | 🟩 Done · e2b8505 · 2026-09-29 · tests passed |
| AJ-0015 | ajiya | ajiya ticket done --test | Runs the configured test command, records tests passed, refuses on failure | - | 🟩 Done · c0239ef · 2026-09-29 · tests passed |
| AJ-0016 | ajiya | ajiya ticket block and ticket drop | block sets a reason; drop needs --reason and --by and counts as closed | - | 🟩 Done · 1ad6b45 · 2026-09-29 |
| AJ-0037 | ajiya | Milestone config and model | [[milestones]] in ajiya.toml, in file order, targets are phases or tickets; each ticket belongs to the earliest milestone that requires it, else unscheduled; old [launch] means a milestone named launch and both forms together is a config error; table tests cover several milestones, a shared ticket, unscheduled tickets and the old form | AJ-0010 | 🟥 Pending |
| AJ-0038 | ajiya | ajiya milestone commands and next by milestone | milestone add (launch stays last by default), list, show, move and remove, all with --json; launch set and show are shorthand for the launch milestone and their tests still pass; next sorts by earliest milestone then unblocks and shows it, with a milestone field in JSON; CLI tests cover each | AJ-0011, AJ-0037 | 🟥 Pending |
| AJ-0039 | ajiya | Milestone checks | Errors for a target that does not exist and a duplicate name; warnings for milestones in the wrong order and for open unscheduled tickets (up to 10 IDs); stable codes and a fixture each | AJ-0037 | 🟥 Pending |
| AJ-0040 | ajiya | Phase display order | [phases] order in ajiya.toml, display only; phase move, phase list (--json) and phase order --suggest/--apply with a topological order, ties by milestone then creation, cycles kept together; phase rename updates it; a check warning for order against dependencies with a fixture; every listing uses the order | AJ-0037 | 🟥 Pending |
