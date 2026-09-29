# Ajiya

Ajiya is a Go command-line tool that keeps a project's phases and tickets as
markdown in the repository. This repository builds Ajiya with Ajiya.

- Build: `go install ./cmd/ajiya` (puts `ajiya` in `$(go env GOPATH)/bin`). Do not
  use `go build -o ajiya`: `ajiya/` is the plan folder, so the binary would land in it.
- Test: `go test ./...`
- The roadmap is in `docs/roadmap/`: read `00-build-plan.md` first.
- Never commit anything from `reference/private/`, copy it into fixtures or quote it.

<!-- ajiya:start -->
## Work tracking (Ajiya)

This repository tracks work with Ajiya. The plan is in `ajiya/`; never edit it by
hand.

1. Before a task: `ajiya next`, then `ajiya ticket start <ID>`.
2. Commit with a last paragraph `Ajiya: <ID>` (1 to 3 IDs, or `Ajiya: chore`).
3. Finish with `ajiya ticket done <ID> --test`, then `ajiya check`.
4. Found extra work? `ajiya ticket add`, do not widen the current ticket.

| Command | Purpose |
|---|---|
| `ajiya next` | Tickets that can start now |
| `ajiya ticket show <ID>` | A ticket, its dependencies and commits |
| `ajiya check` | Problems in the plan |

Full guides: `.ajiya/guide/daily.md` and `.ajiya/guide/setup.md`.
<!-- ajiya:end -->
