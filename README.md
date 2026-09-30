# Ajiya

*Ajiya* is Hausa for safekeeping: something held in trust.

Ajiya is a command-line tool that keeps a project's work breakdown (phases and
tickets) in plain markdown inside the repository. Coding agents plan, update and
prove their work through `ajiya` commands, and people can see launch readiness at
a glance. Everything is local: no server, no account, no network calls.

Status: v0.2. Prebuilt binaries, an install script, an npm package and a
Homebrew cask are available.

## How it works

```
project   the repository; it may hold several apps
  phase   one whole feature ("Accounts", "Bookings", "Go-live"): one file in ajiya/
    ticket  one piece of work an agent can finish and prove in one session
```

- **Agents decide the work; the CLI keeps the books.** IDs, files, statuses,
  evidence and order all come from `ajiya` commands. Nobody edits `ajiya/` by
  hand, and `ajiya check` notices when someone does.
- **Order comes from dependencies.** `ajiya next` lists the tickets whose
  dependencies are all done, earliest milestone first.
- **Every commit names its tickets.** A last paragraph `Ajiya: CB-0007` (1 to 3
  IDs, or `Ajiya: chore` for upkeep). Git hooks enforce it, and
  `ajiya check --commits` backs them up in CI.
- **Done means proven.** `ajiya ticket done` refuses without a commit that names
  the ticket, and records it as evidence. With `--test` it also runs your tests.
- **Agents learn it from the repository.** `ajiya init` installs two short guides
  in `.ajiya/guide/`, a Claude Code skill, and a block in `AGENTS.md` and
  `CLAUDE.md`, so Claude Code, Codex and others follow the same rules.

## Install

Pick one. Each line installs the same `ajiya` binary.

macOS or Linux (install script):

```sh
curl -fsSL https://github.com/AliyuYahaya/Ajiya/releases/latest/download/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://github.com/AliyuYahaya/Ajiya/releases/latest/download/install.ps1 | iex
```

npm (any platform):

```sh
npm install -g @ajiya/cli
```

Homebrew (macOS):

```sh
brew install --cask AliyuYahaya/tap/ajiya
```

- **Install script and PowerShell script:** they check the download against
  `checksums.txt` and need no administrator rights or `sudo`. The shell script
  installs to `~/.local/bin` (or `AJIYA_INSTALL_DIR`); the PowerShell script to
  `%LOCALAPPDATA%\Programs\ajiya\bin` and adds that folder to your user `PATH`.
  `AJIYA_VERSION` pins a version. On a terminal, `install.sh` offers to run
  `ajiya mcp install` at the end.
- **npm:** installs only the package for your machine's platform, and runs no
  install scripts. To try it without installing: `npx @ajiya/cli init`.
- **Homebrew:** the binaries are signed and notarised, so macOS shows no
  Gatekeeper prompt. The cask tells you to run `ajiya mcp install` and
  `ajiya init`.
- **From source**, with Go 1.26 or later:
  `go install github.com/AliyuYahaya/Ajiya/cmd/ajiya@latest` (puts `ajiya` in
  `$(go env GOPATH)/bin`; make sure that folder is on your `PATH`).
- **Pre-releases:** `releases/latest` never points at a pre-release. To try one,
  use that release's own URL, for example
  `curl -fsSL https://github.com/AliyuYahaya/Ajiya/releases/download/v0.2.0-rc.4/install.sh | AJIYA_VERSION=v0.2.0-rc.4 sh`,
  or `npm install -g @ajiya/cli@next`.

### What just happened?

- **Once per computer:** install Ajiya (above).
- **Once per project:** run `ajiya init`. It sets up the project and registers
  Ajiya's MCP server with the coding agents it finds (Claude Code, Codex,
  Cursor). It shows the change, keeps backups and ends with an
  `Undo with: ajiya mcp uninstall ...` line. Register without a project with
  `ajiya mcp install`.
- **Every day:** open your agent. It starts `ajiya mcp` for that session and
  closes it afterwards; nothing runs while the agent is closed.

### Cloud agents

Claude Code on the web and Codex cloud read the committed `CLAUDE.md` and
`AGENTS.md`, but `ajiya` must exist in their sandbox. Add
`npm install -g @ajiya/cli` to the environment's setup script.

### Tell your agent

Paste one of these into your agent. New project:

```
Set up Ajiya in this repository and plan: <what you want to build>. If the ajiya command is missing, install it with npm install -g @ajiya/cli (or the install script from https://github.com/AliyuYahaya/Ajiya). Then run ajiya init --yes and ajiya hook install, and follow .ajiya/guide/setup.md: ask me at most three questions, then hand over a summary. Don't start building until I say so.
```

Existing project:

```
Set up Ajiya in this repository and bring in the work we already track. If the ajiya command is missing, install it with npm install -g @ajiya/cli (or the install script from https://github.com/AliyuYahaya/Ajiya). Run ajiya init --yes and ajiya hook install, import our GitHub issues with ajiya import github (or our TODO.md with ajiya import todo TODO.md), then organise them as .ajiya/guide/setup.md describes and hand over a summary. List duplicates for me instead of dropping anything.
```

Claude Code users can also add the plugin (skill and a session-start
status line): `/plugin marketplace add AliyuYahaya/Ajiya`, then
`/plugin install ajiya@ajiya`. The plugin does not register the MCP server; run
`ajiya mcp install` for that. See [plugin/README.md](plugin/README.md).

Cursor users: [![Add to Cursor](https://cursor.com/deeplink/mcp-install-dark.svg)](https://cursor.com/install-mcp?name=ajiya&config=eyJjb21tYW5kIjoiYWppeWEiLCJhcmdzIjpbIm1jcCJdfQ%3D%3D) adds the `ajiya mcp` server (`ajiya` must be on your `PATH`).

## Walkthrough: a new project

Start in an empty folder, or a fresh repository:

```
mkdir clinic-booking && cd clinic-booking
git init
ajiya init --name 'Clinic Booking'
ajiya hook install
```

`init` writes `ajiya.toml`, the `ajiya/` folder and the agent kit, and picks the
ticket prefix from the name (`CB-0001`, `CB-0002`, ...). If you want
`ajiya ticket done --test` to run your tests, set the test command in
`ajiya.toml`:

```toml
[test]
command = "npm test"
```

Now ask your agent to plan it, for example: *"Plan an online booking system for
a small clinic with Ajiya."* The agent follows `.ajiya/guide/setup.md`: it asks
at most three questions, adds phases and tickets with dependencies, adds tickets
for work that is easy to forget (secrets, backups, monitoring, TLS, rollback),
marks decisions only you can make as `Needs a human`, and hands over a summary.

Or plan by hand. The same commands are what the agent runs:

```
ajiya phase add bookings "Patients book, move and cancel appointments"
ajiya ticket add --phase bookings --app infra "Booking form" \
  --done-when "A patient books a free slot; tests cover a taken slot"
ajiya milestone add launch --targets bookings
ajiya next
```

Then work one ticket at a time. Your agent follows `.ajiya/guide/daily.md`:

```
ajiya ticket start CB-0001
# ... write the code and tests ...
git commit -m "Booking form" -m "Ajiya: CB-0001"
ajiya ticket done CB-0001 --test
ajiya check
```

A commit without the trailer is refused:

```
ajiya hook run: commit refused: the message names no ticket. End it with a
paragraph 'Ajiya: <ID>' (1 to 3 IDs), or 'Ajiya: chore' for upkeep.
```

## Walkthrough: an existing project

In the repository:

```
cd shop
ajiya init --name Shop
ajiya hook install
```

`init` looks for apps (workspace files such as `pnpm-workspace.yaml` or
`go.work`, then folders with their own `package.json`, `go.mod`,
`pyproject.toml` and so on) and asks which to register:

```
Apps found:
  1.  web  apps/web    (pnpm-workspace.yaml; dev script)
  2.  api  apps/api    (pnpm-workspace.yaml)
Register which? all, none, or numbers such as 1,3 [all]:
```

Add `--yes` to register them all without asking. Missed one?
`ajiya app add <name> --path <dir>`.

Bring in the work you already track. Imports never reword anything:

```
ajiya import todo TODO.md --app web     # markdown checkboxes; ticked ones are done
ajiya import github --repo owner/shop   # issues, through the gh CLI
```

Imported tickets land in a phase called `inbox`. Ask your agent to organise
them: *"Organise the imported tickets with Ajiya."* Following the setup guide, it
creates a phase per feature, moves each ticket out of `inbox`, adds the
dependencies and the missing work, and sets milestones. It lists duplicates for
you rather than dropping them: only a person drops a ticket.

```
ajiya phase list
ajiya milestone list
ajiya next
```

From here, every task follows the same loop as a new project.

## See the plan

```
ajiya build     # writes ajiya/PROGRESS.md and the dashboard: open ajiya/index.html
ajiya serve     # serves the dashboard on 127.0.0.1 and follows changes live
```

The dashboard shows progress per milestone, what can start now, a board, the
dependency graph by ticket or by phase, the checks and recent activity from git.
`ajiya/PROGRESS.md` is the same summary in markdown.

## Commands

Run `ajiya help` for the list and `ajiya <command> --help` for one command.
Every read command takes `--json`.

| Command | Does |
|---|---|
| `ajiya init` | Sets up a project, or updates its agent kit |
| `ajiya next` | Tickets that can start now |
| `ajiya ticket add / start / done / block / show / edit` | Work with one ticket |
| `ajiya ticket drop` | Closes a ticket that is not needed (a person's decision) |
| `ajiya phase add / list / rename / move / order` | Work with phases |
| `ajiya milestone add / list / show / move / remove` | Named gates, in order |
| `ajiya app add / move / remove` | The apps in the repository |
| `ajiya import todo / github / legacy` | Bring existing work in |
| `ajiya check` | Problems in the plan and, with `--commits`, in commit messages |
| `ajiya build` / `ajiya serve` | The progress summary and the dashboard |
| `ajiya changelog` | The work in a range of commits, by ticket |
| `ajiya hook install` | The commit-message hooks |

## Developing Ajiya

Ajiya is built with Ajiya: the plan for this repository is in `ajiya/`, and the
roadmap is in `docs/roadmap/`.

```
go install ./cmd/ajiya    # not 'go build -o ajiya': ajiya/ is the plan folder
go test ./...
ajiya next
```

## License

MIT: see [LICENSE](LICENSE). Copyright Yavid PTY Ltd. The name "Ajiya" and the
Ajiya logo belong to Yavid PTY Ltd; the licence does not grant them. A fork or a
copy you share or sell must use a different name and logo.
