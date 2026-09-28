<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Bookings

Goal: A patient can book a slot end to end

| ID | App | Ticket | Done when | Depends | Status |
|---|---|---|---|---|---|
| CB-0007 | api | Add slots table and GET /slots | Returns free slots for a day; tests pass | CB-0002 | 🟩 Done · 4f2a91c · 2026-09-28 · tests passed |
| CB-0008 | api | Add POST /bookings with validation | Rejects past dates with 422 | CB-0007 | 🟨 In progress: validation done, tests to do |
| CB-0009 | web | Booking screen with slot picker | A patient can pick and confirm a slot | CB-0007 | 🟥 Pending |
| CB-0010 | infra | Set up production hosting | App reachable on the clinic's domain | - | 🟥 Pending · Needs a human: hosting account |
