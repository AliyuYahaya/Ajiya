# Agent kit templates

Real templates for the files `ajiya init` installs (tickets AJ-0022 to AJ-0025).
Collected 2026-09-28.

## SKILL.md

Sources:
- Agent Skills specification: https://agentskills.io/specification
- Anthropic's template: https://github.com/anthropics/skills/blob/main/template/SKILL.md
- A real skill with helper files: https://github.com/anthropics/skills/blob/main/skills/webapp-testing/SKILL.md
- Claude Code skill docs: https://code.claude.com/docs/en/skills

Anthropic's template, in full:

```markdown
---
name: template-skill
description: Replace with description of the skill and when Claude should use it.
---

# Insert instructions below
```

Rules from the specification:

| Field | Rule |
|---|---|
| `name` | Required. 1 to 64 characters: `a-z`, `0-9`, `-`; no leading, trailing or double hyphen; must match the folder name |
| `description` | Required. 1 to 1024 characters. Says what the skill does **and when to use it**, with the words a user would say |
| `license`, `compatibility`, `metadata`, `allowed-tools` | Optional. Most skills need none of them |

Claude Code specifics: the `---` must be the first line of the file; the key use
case goes first in the description (description plus `when_to_use` is capped at
1,536 characters in the listing).

Layout: `SKILL.md` plus optional `scripts/`, `references/`, `assets/`. Loading is
progressive: name and description always (about 100 tokens), the body when the
skill is used (under 5,000 tokens, under 500 lines), other files only when read.
Keep references one level deep from `SKILL.md`.

Good description, from the specification:
> Extracts text and tables from PDF files, fills PDF forms, and merges multiple
> PDFs. Use when working with PDF documents or when the user mentions PDFs, forms,
> or document extraction.

Poor: "Helps with PDFs."

What the real skills do well: tell the agent to run helper commands with `--help`
before reading their source; give a short decision tree for choosing an approach;
show exact commands to copy.

### Shape for Ajiya's skill

```markdown
---
name: ajiya
description: Plans, tracks and proves work in this repository with the ajiya CLI: picks the next ticket, starts it, commits with the Ajiya trailer and marks it done with evidence. Use before starting any task in this repository, when deciding what to work on, before committing, and when work is finished.
---

# Ajiya

<one paragraph: the plan lives in ajiya/, changed only through commands>

## Rules
<the five rules, one line each>

## Guides
- Planning a project or organising imported tickets: `.ajiya/guide/setup.md`
- The loop for every task: `.ajiya/guide/daily.md`
```

## AGENTS.md

Sources:
- Convention: https://agents.md
- The agents.md project's own file (MIT): https://github.com/openai/agents.md/blob/main/AGENTS.md
- A large real one: https://github.com/openai/codex/blob/main/AGENTS.md

The format has no required fields: it is plain markdown. The nearest `AGENTS.md`
to a file wins, so monorepos can nest one per app. Common sections: project
overview, build and test commands, code style, testing, security, commit message
format, pull request rules.

What the real files do well:
- Numbered sections with an imperative heading ("Use the development server, not
  `npm run build`") and the reason underneath.
- A closing table of commands and what each is for.
- Concrete rules with the exact command to run after a change ("run `just fmt`
  after code changes; do not ask for approval").
- Say what the agent may run without asking and what needs the user.

### Shape for Ajiya's block

`init` writes only between markers, so the user's own text is never touched:

```markdown
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
```

The same block goes into `CLAUDE.md`, between the same markers.
