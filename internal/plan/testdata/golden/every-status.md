<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Every status

Goal: One row for each status form, pipes \| and all

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| CB-0001 | api | Parse a \| b | Output is a \| b | - | 🟥 Pending |
| CB-0002 | api | Blocked and human | - | CB-0001 | 🟥 Pending · Needs a human: DNS access · Blocked: waiting on vendor |
| CB-0003 | api | Started, no note | x | CB-0001, CB-0002 | 🟨 In progress |
| CB-0004 | api | Started, human | x | - | 🟨 In progress: half done · Needs a human: sign-off |
| CB-0005 | infra | Done by a person | x | - | 🟩 Done · by Amina Bello · 2026-09-01 · Note: hosting account opened |
| CB-0006 | api | Done with commit and note | x | - | 🟩 Done · 4f2a91cd · 2026-09-02 · Note: edge cases left for CB-0009 |
| CB-0007 | web | Dropped | x | - | 🟩 Dropped: not needed after all, decided by Amina Bello |
| CB-10000 | web | Five-digit ID | x | CB-0007 | 🟥 Pending |
| CB-10001 | web | Imported, closed issue | x | - | 🟩 Done · issue #12 · Link: https://github.com/o/r/issues/12 |
| CB-10002 | web | Imported from a todo list | x | - | 🟩 Done · Done before Ajiya |
| CB-10003 | web | Legacy ticket | x | CB-10002 | 🟨 In progress: half · Needs a human: keys · Link: https://x.test/a · Was: MI3-185, RO-053 |
| CB-10004 | web | Dropped with link | x | - | 🟩 Dropped: duplicate, decided by Amina · Link: https://x.test/b |
