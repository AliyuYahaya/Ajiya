# Working with Ajiya every day

Use this guide for every task in a project that has a plan (`ajiya.toml` and an
`ajiya/` folder). To plan a project or organise imported tickets, read
`setup.md` instead.

The plan lives in `ajiya/`. Never edit it by hand: change it only with `ajiya`
commands. The same goes for `ajiya.toml`, except its `[test]` command. Trust the
plan over your memory of earlier sessions.

## The loop

### 1. Pick a ticket

```
ajiya next
```

It lists the tickets that can start now: open, not blocked, every dependency
done. The earliest milestone comes first, then what unblocks the most work. Take
the first one unless the user asked for something else. Skip tickets marked
`Needs a human` unless the user asks for them: tell the user about them instead.

If the user asks for work that is not in the plan, find the ticket
(`ajiya ticket show <ID>` also finds old IDs) or add one first (step 6).

If the ticket the user wants is not in `ajiya next`, run `ajiya ticket show <ID>`
to see what it waits on, and **tell the user before you start it**. Starting
blocked work usually means building on something that is not there yet.

### 2. Start it

```
ajiya ticket start <ID>
```

Read its "done when". That is the check you must pass, not just the title.

### 3. Write the code and the tests

Keep to the ticket. Tests prove the "done when" where they can.

### 4. Commit with the trailer

Every commit ends with a trailer that names 1 to 3 tickets:

```
Add booking model and migration

Ajiya: CB-0142
```

- More than 3 tickets means the commit is too big: split it.
- `Ajiya: chore` is for upkeep with no ticket: dependency updates, formatting,
  config, plan-only changes.
- Merge commits and reverts need no trailer.

The git hooks from `ajiya hook install` check the trailer and pre-fill it from the
tickets in progress. If a hook refuses a commit, fix the message; never skip the
hook.

### 5. Mark it done

```
ajiya ticket done <ID> --test
```

`done` refuses unless a commit names the ticket, and records the newest such
commit as evidence. Only commits made after `ticket start` count, so commit the
work itself with the trailer before you close the ticket. `--test` runs the
project's test command and refuses if it fails. Never mark a ticket done any
other way, and never before its "done when" is met.

`done` changes files in `ajiya/` (the phase files, `PROGRESS.md`, `data.js` and
`index.html`). Commit them with the same ticket's trailer (`Ajiya: <ID>`) so the
tree is clean.

If `done --test` says there is no `[test]` command, the project has not set one
yet. Set it in `ajiya.toml` as `setup.md` section 2 describes (ask the user if you
cannot tell the command), then run `done --test` again. Leave out `--test` only
when the project has no tests at all, and say so in your summary.

### 6. Check

```
ajiya check
```

Fix every error before you finish. Then tell the user what you did, and which
ticket is next.

## On the way

**Found extra work?** Add a ticket; do not widen the current one.

```
ajiya ticket add --phase <slug> --app <app> "<title>" \
  --done-when "<check>" --depends <IDs>
```

If a milestone should wait for it, make a ticket that milestone needs depend on it.

**Stuck on something outside your control?** Block the ticket and say what it
waits on. `ajiya ticket start` clears the block later.

```
ajiya ticket block <ID> "waiting for the API key from the payment provider"
```

**Needs a person?** Some work only a person can do: choose a host, sign a
contract, approve a design. Such tickets are marked `Needs a human`. Do not do
them yourself. When the person has done it, they (or you, on their word) close it:

```
ajiya ticket done <ID> --by "<name>" --note "<what was done>"
```

**Not needed any more?** Tell the user. Only a person drops a ticket:
`ajiya ticket drop <ID> --reason "<why>" --by "<name>"`.

**Created a new app folder?** Register it in the same task:
`ajiya app add <name> --path <dir>` (`--library` for a shared package). Create the
folder first: `app add` only warns when the path is not a folder yet. For work outside app
folders (hosting, CI, DNS) use the app `infra` instead.

**Ticket has no "done when"?** Imported tickets show `-`. Set one before you start:
`ajiya ticket edit <ID> --done-when "<check>"`.

**Plan wrong?** Fix it with commands: `ajiya ticket edit <ID>` changes the title,
"done when", dependencies, app or phase. `--depends` replaces the whole list, so
include the ones already there.

## Commands

| Command | Use it to |
|---|---|
| `ajiya next [--app <app>] [--milestone <name>]` | See what can start now |
| `ajiya ticket show <ID>` | See a ticket, what it waits on, what waits on it, and its commits |
| `ajiya ticket start <ID>` | Start a ticket |
| `ajiya ticket done <ID> --test` | Close a ticket with evidence |
| `ajiya ticket block <ID> "<reason>"` | Park a ticket that waits on something outside the plan |
| `ajiya ticket list [--phase <slug>] [--app <app>] [--state <state>] [--milestone <name>]` | List tickets in plan order (`--state` takes pending, in_progress, done, dropped, ready or human) |
| `ajiya ticket add ...` | Add work found on the way |
| `ajiya ticket edit <ID> ...` | Correct a ticket |
| `ajiya milestone list` / `ajiya milestone show <name>` | See progress to each gate, and what blocks it |
| `ajiya launch show` | See what launch still needs, when the project uses a single launch target |
| `ajiya phase list` | See each phase's progress |
| `ajiya check` | Find problems in the plan and fix them |
| `ajiya build` | Write `ajiya/PROGRESS.md` and the dashboard data |
| `ajiya serve` | Open the dashboard on 127.0.0.1; it follows changes |
| `ajiya changelog [<range>]` | List the work in a range of commits, by ticket |

Every read command takes `--json`.
