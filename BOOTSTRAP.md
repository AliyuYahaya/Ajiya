# Bootstrap plan

Temporary. The only plan written by hand. At the end of M1 every unticked item
below is added with `ajiya ticket add` (phase, app, done-when and dependencies as
written here), and this file is deleted.

Item keys (`M1.3`) are only for dependencies inside this file. Each item is one
ticket: one session's work that can be proved.

## M1. Self-hosting core (phase `core`)

- [x] M1.1 Config loading: read `ajiya.toml` (project, launch, test, apps); `infra`
      always valid. Done when: table-driven tests cover valid, missing and malformed
      config, with exit 2 and a one-line error. Depends: -
- [x] M1.2 Phase file model and reader: parse header, goal, table, status cell
      (mark, note, evidence, `Needs a human`, `Blocked`, `Dropped`), `\|` escapes.
      Done when: every example row in the spec parses. Depends: -
- [x] M1.3 Canonical writer: golden-file round-trip tests (read then write gives the
      same bytes). Depends: M1.2
- [x] M1.4 ID allocation: one sequence per project, `<PREFIX>-<4 digits>`, never
      reused, including dropped. Done when: table tests pass. Depends: M1.1, M1.2
- [x] M1.5 Minimal `ajiya init`: writes `ajiya.toml` and `ajiya/`, no detection, no
      kit. Depends: M1.1
- [x] M1.6 `ajiya phase add <slug> "<goal>"`. Depends: M1.3, M1.5
- [x] M1.7 `ajiya ticket add`: validates app, phase, dependencies. Depends: M1.4, M1.6
- [x] M1.8 `ajiya ticket start` and `ticket show` (row, phase, dependants; commits
      come in M2). Depends: M1.7
- [x] M1.9 `ajiya ticket done`: requires a commit whose trailer references the ticket;
      records hash and date; `--by/--note` for `Needs a human`. Depends: M1.7
- [x] M1.10 `ajiya ticket edit`, refusing changes that create a loop. Depends: M1.7, M1.12
- [x] M1.11 `ajiya next [--app]`: startable tickets, those that unblock most first.
      Depends: M1.7
- [x] M1.12 Dependency graph: missing IDs, self-dependency, loops (printed).
      Depends: M1.2
- [x] M1.13 `ajiya check` with error-level checks for format, round-trip, IDs,
      dependencies and loops; stable codes. Depends: M1.3, M1.12
- [x] M1.14 CLI test harness with testscript (rogpeppe/go-internal, a test helper):
      txtar scripts run the binary in a temporary git repo.
      Depends: M1.5
- [ ] M1.15 Run `ajiya init` on this repository, add these items as tickets, delete
      this file, write a minimal `.claude/skills/ajiya/SKILL.md`. Depends: M1.1–M1.14

## M2. The commit rule (phase `commit-rule`)

- [ ] M2.1 Trailer parsing: `Ajiya: X, Y`, `Ajiya: chore`, merge and revert exemptions.
      Parse with git's own rules (`git interpret-trailers --parse` for a message file,
      `%(trailers:key=Ajiya)` for history) so the hook, `done` and `check` always agree.
      Depends: M1.15
- [ ] M2.2 `ajiya check --commits <range>`: 0 or >3 references, unknown IDs.
      Depends: M2.1
- [ ] M2.3 `commit-msg` hook; exempts merges (MERGE_HEAD present). `--no-verify` skips
      it, which is why the CI check exists. Depends: M2.1
- [ ] M2.4 `prepare-commit-msg` hook: pre-fill from in-progress tickets and from
      changed rows in `ajiya/`-only commits, via `git interpret-trailers --in-place`;
      leave `merge`, `squash` and `commit` (amend) sources alone. Depends: M2.1
- [ ] M2.5 `ajiya hook install`: respects `core.hooksPath`, appends a marked block to
      existing hooks, `#!/bin/sh` scripts that call `ajiya hook run <name>` (works in
      Git for Windows), and a clear message when `ajiya` is not on PATH.
      Depends: M2.3, M2.4
- [ ] M2.6 CI job running `ajiya check --commits`; hooks installed on this repo.
      Depends: M2.2, M2.5
- [ ] M2.7 `ticket show` lists commits that reference the ticket. Depends: M2.1

## M3. Apps, launch and full checks (phase `apps-launch`)

- [ ] M3.1 App detection for `init`: workspace files, manifests, ranking, skip list,
      `--yes`. Depends: M1.15
- [ ] M3.2 `ajiya app add/remove/move`, with `--to` for open tickets. Depends: M1.15
- [ ] M3.3 `ajiya launch set/show`: required set = target's tickets plus transitive
      dependencies. Depends: M1.15
- [ ] M3.4 `ajiya next --launch`, launch-required first. Depends: M3.3
- [ ] M3.5 Remaining error checks (unregistered app, done without evidence, missing
      launch target), a fixture per code. Depends: M3.2, M3.3
- [ ] M3.6 Warning checks W-codes and `--strict`, a fixture per code. Depends: M3.5, M2.1
- [ ] M3.7 `--json` on every read command. Depends: M3.6
- [ ] M3.8 `ajiya ticket done --test`: run test command, record pass/fail, refuse on
      fail. Depends: M1.15
- [ ] M3.9 `ajiya ticket block` and `ticket drop --reason --by`. Depends: M1.15

## M4. Outputs (phase `outputs`)

- [ ] M4.1 Activity from git: last 30 days, tickets, status changes from diffs, agent
      co-authors. Depends: M2.1
- [ ] M4.2 `ajiya build`: `PROGRESS.md` and `data.js` (schema version), write only on
      change, dated by newest source change. Depends: M3.6, M4.1
- [ ] M4.3 Dashboard adapted from `reference/index.html`: MedInformer specifics
      removed, CSS variable themes (light/dark), all views plus Activity, embedded.
      Depends: M4.2
- [ ] M4.4 `ajiya serve [--port]`: 127.0.0.1 only, answers 403 unless the Host header
      is 127.0.0.1, localhost or [::1] with the port (DNS rebinding), rebuilds on change. Depends: M4.3
- [ ] M4.5 Launch target `v0.1` set on this repo; dashboard in use. Depends: M4.4

## M5. Onboarding (phase `onboarding`)

- [ ] M5.1 `.ajiya/guide/setup.md`. Depends: M4.5
- [ ] M5.2 `.ajiya/guide/daily.md`. Depends: M4.5
- [ ] M5.3 `SKILL.md` and the `AGENTS.md`/`CLAUDE.md` marker blocks. Depends: M5.1, M5.2
- [ ] M5.4 `init` writes and updates the kit without touching user text. Depends: M5.3, M3.1
- [ ] M5.5 `ajiya import todo <file>`. Depends: M1.15
- [ ] M5.6 `ajiya import github` via `gh`. Depends: M1.15
- [ ] M5.7 `ajiya import legacy <config>` with aliases and duplicate report; tested on
      `reference/private/` if present. Depends: M1.15
- [ ] M5.8 Re-run `ajiya init` on this repo so its kit is the shipped one.
      Depends: M5.4

## M6. Release v0.1 (phase `release`)

- [ ] M6.1 GoReleaser config (macOS, Linux, Windows; version via ldflags). Depends: M5.8
- [ ] M6.2 Homebrew tap config with `homebrew_casks` (GoReleaser deprecated `brews`),
      including quarantine removal for unsigned binaries; not published. Depends: M6.1
- [ ] M6.3 Changelog built from `Ajiya:` trailers. Depends: M2.1
- [ ] M6.4 README walkthroughs: new project, existing project. Depends: M5.8
- [ ] M6.5 Definition of done: `check --strict` passes, `check --commits` since M2
      passes, dashboard at 100% for `v0.1`, both walkthroughs timed under ten
      minutes. Needs a human: sign-off. Depends: M6.1–M6.4
