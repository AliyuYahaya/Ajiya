#!/usr/bin/env python3
"""
Collates every WBS in docs/ and drives the rollout tracker.

  python3 rollout/collate.py            collate and write every output
  python3 rollout/collate.py --check    check only: write nothing, exit 1 on errors
  python3 rollout/collate.py --check --strict
                                        as --check, but warnings also fail

Agents should run --check before they finish a piece of work. A git
pre-commit hook can run it too.

It reads each source WBS below (each file stays the source of truth for
its own tickets), then:

  1. writes rollout/all-tickets.md: every ticket, grouped by area and epic,
     with its status, plus a list of IDs used by more than one ticket;
  2. fills the Status column of rollout/rollout-wbs.md for every work
     package that has a Scope, from the statuses of the tickets in scope
     (work packages with no Scope keep the status typed into them);
  3. writes rollout/PROGRESS.md: overall numbers, progress per rollout
     phase, what can start next, and any problems found (Checks);
  4. writes rollout/data.js: the same data for the index.html dashboard.

Status is read from the last table cell containing 🟩 (done), 🟨
(partial) or 🟥 (pending). A ticket with none of them counts as unknown.
A work package with no Scope may also use the words Done, In progress,
Partial, Not started or Pending; the mark is still preferred.

Scope syntax in rollout-wbs.md (comma separated):
  WEB-013               one ticket
  WEB-037..WEB-044       an inclusive range
  OPS[Phase 0]          every ticket under a heading starting "Phase 0"
  WEB[EPIC G]           every WEB ticket under a heading starting "EPIC G"

Cross-references: a ticket whose title or status says "tracked as
API-185", "cross-reference only: API-185", "duplicate of API-185" or
"same as API-185" is the same work as API-185. If both are in one work
package, it is counted once.

Checks (printed, and listed in PROGRESS.md and on the dashboard):
  errors    a source file is missing; a dependency names an RO item that
            doesn't exist; dependencies form a loop; an RO ID is used twice
  warnings  a source file yields no tickets; a row uses an unknown ID
            prefix; tickets with no status mark; a status taken from a
            column other than the last; scope entries not found; a ticket
            in more than one work package; an RO item named in a package's
            text but missing from its Depends column; a typed status with
            no mark
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import re
import sys
from collections import defaultdict
from dataclasses import dataclass, field
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent  # apps/
ROLLOUT = ROOT / "rollout"

SOURCES = [
    ("API", "docs/api/wbs.md"),
    ("Web app", "docs/web/wbs.md"),
    ("Operations", "docs/ops/wbs.md"),
    ("Priority gaps", "rollout/priority-plan.md"),
]

# Every ticket ID prefix. Add a new one here only: all patterns below use it.
PREFIXES = ("API", "WEB", "OPS", "PRI")
PFX = "(?:" + "|".join(re.escape(p) for p in PREFIXES) + ")"

MENTION_RE = re.compile(rf"\b({PFX}-\d{{3,}})\b")
ID_RE = re.compile(rf"^\|\s*({PFX}-\d+)\s*\|")
ANY_ID_ROW_RE = re.compile(r"^\|\s*([A-Z][A-Z0-9]*)-\d+\s*\|")
GROUP_RE = re.compile(rf"^({PFX})\[(.+)\]$")
RANGE_RE = re.compile(rf"^({PFX}-\d+)\.\.({PFX}-\d+)$")
SINGLE_RE = re.compile(rf"^{PFX}-\d+$")
ALIAS_RE = re.compile(
    rf"(?:tracked\s+as|cross-reference\s+only\W+|duplicate\s+of|same\s+as)\s+({PFX}-\d+)", re.I
)
RO_RE = re.compile(r"\bRO-\d+\b")
CELL_SPLIT = re.compile(r"(?<!\\)\|")
STATUS_MARKS = {"🟩": "done", "🟨": "partial", "🟥": "pending"}
ICON = {"done": "🟩", "partial": "🟨", "pending": "🟥", "unknown": "⬜"}
# Typed statuses on rollout-only rows, checked in order.
STATUS_WORDS = [
    (re.compile(r"^\s*(done|complete|completed)\b", re.I), "done"),
    (re.compile(r"^\s*(in progress|partial|started)\b", re.I), "partial"),
    (re.compile(r"^\s*(not started|pending|to do|todo|blocked)\b", re.I), "pending"),
]


# ----------------------------------------------------------------- checks

class Issues:
    def __init__(self) -> None:
        self.items: list[dict[str, str]] = []

    def error(self, where: str, msg: str) -> None:
        self.items.append({"level": "error", "where": where, "msg": msg})

    def warn(self, where: str, msg: str) -> None:
        self.items.append({"level": "warning", "where": where, "msg": msg})

    def note(self, where: str, msg: str) -> None:
        self.items.append({"level": "info", "where": where, "msg": msg})

    def count(self, level: str) -> int:
        return sum(i["level"] == level for i in self.items)


def plural(n: int, word: str) -> str:
    return f"{n} {word}{'' if n == 1 else 's'}"


def short_list(ids: list[str], n: int = 12) -> str:
    return ", ".join(ids[:n]) + (f" and {len(ids) - n} more" if len(ids) > n else "")


# ---------------------------------------------------------------- parsing

@dataclass
class Ticket:
    id: str
    title: str
    status: str
    area: str
    source: str
    epic: str
    status_text: str = ""
    mentions: set[str] = field(default_factory=set)
    alias_of: str = ""

    @property
    def key(self) -> str:
        return f"{self.id}|{self.source}"


@dataclass
class WorkPackage:
    id: str
    name: str
    scope: str
    depends: list[str]
    launch: str
    phase: str
    manual_status: str
    line_no: int
    resolved: list[Ticket] = field(default_factory=list)  # counted tickets
    aliased: list[Ticket] = field(default_factory=list)   # in scope, counted as another ticket
    missing: list[str] = field(default_factory=list)
    ambiguous: list[str] = field(default_factory=list)    # IDs in scope that name more than one ticket


def cells(line: str) -> list[str]:
    parts = CELL_SPLIT.split(line.strip())
    return [p.strip() for p in parts[1:-1]]


def status_of(row: list[str]) -> tuple[str, str, int]:
    for idx in range(len(row) - 1, -1, -1):
        for mark, name in STATUS_MARKS.items():
            if mark in row[idx]:
                return name, row[idx], idx
    return "unknown", "", -1


def short_title(row: list[str]) -> str:
    text = row[1] if len(row) > 1 else ""
    text = re.sub(r"^`[^`]*`:\s*", "", text)  # drop a leading `feat(x)`: prefix
    text = re.sub(r"\*\*|`", "", text)
    # Source cells may already escape pipes; store them plain, escape once on output.
    return text.replace("\\|", "|")


def parse_source(area: str, rel: str, issues: Issues) -> list[Ticket]:
    path = ROOT / rel
    if not path.exists():
        issues.error(rel, f"Source file missing, so {area} contributes no tickets. Fix the path in SOURCES.")
        return []
    tickets, epic = [], ""
    unknown_prefix: dict[str, list[str]] = defaultdict(list)
    no_mark: list[str] = []
    not_last: list[str] = []
    for line in path.read_text(encoding="utf-8").splitlines():
        heading = re.match(r"^#{2,3}\s+(.*)", line)
        if heading:
            epic = heading.group(1).strip()
            continue
        m = ID_RE.match(line)
        if not m:
            other = ANY_ID_ROW_RE.match(line)
            if other and other.group(1) not in PREFIXES and other.group(1) != "RO":
                unknown_prefix[other.group(1)].append(cells(line)[0])
            continue
        row = cells(line)
        status, text, idx = status_of(row)
        if status == "unknown":
            no_mark.append(m.group(1))
        elif idx != len(row) - 1 and row[-1]:
            not_last.append(m.group(1))
        mentions = set(MENTION_RE.findall(" ".join(row[1:]))) - {m.group(1)}
        alias = ALIAS_RE.search(" ".join(row[1:]))
        tickets.append(Ticket(
            m.group(1), short_title(row), status, area, rel, epic, text, mentions,
            alias.group(1) if alias and alias.group(1) != m.group(1) else "",
        ))

    if not tickets:
        issues.warn(rel, "No ticket rows found. Has the table layout changed? The ticket ID must be in the first column.")
    for prefix, ids in unknown_prefix.items():
        issues.warn(rel, f"Rows skipped because the prefix {prefix}- isn't in PREFIXES ({len(ids)}): {short_list(ids)}")
    if no_mark:
        issues.warn(rel, f"Tickets with no 🟩/🟨/🟥 mark, counted as unknown ({len(no_mark)}): {short_list(no_mark)}")
    if not_last:
        issues.warn(rel, f"Tickets whose status comes from a column other than the last ({len(not_last)}): {short_list(not_last)}")
    return tickets


def id_key(ticket_id: str) -> tuple[str, int]:
    prefix, num = ticket_id.rsplit("-", 1)
    return prefix, int(num)


def resolve_scope(scope: str, tickets: list[Ticket]) -> tuple[list[Ticket], list[Ticket], list[str], list[str]]:
    by_id: dict[str, list[Ticket]] = defaultdict(list)
    for t in tickets:
        by_id[t.id].append(t)

    found: list[Ticket] = []
    missing: list[str] = []
    ambiguous: list[str] = []
    for token in [s.strip() for s in scope.split(",") if s.strip()]:
        group, rng = GROUP_RE.match(token), RANGE_RE.match(token)
        if group:
            prefix, heading = group.groups()
            hits = [t for t in tickets if t.id.startswith(prefix + "-") and t.epic.lower().startswith(heading.lower())]
            found.extend(hits) if hits else missing.append(token)
        elif rng:
            (p1, n1), (p2, n2) = id_key(rng.group(1)), id_key(rng.group(2))
            if p1 != p2 or n2 < n1:
                missing.append(token)
                continue
            for n in range(n1, n2 + 1):
                tid = f"{p1}-{n:03d}"
                found.extend(by_id[tid]) if tid in by_id else missing.append(tid)
                if len(by_id.get(tid, [])) > 1:
                    ambiguous.append(tid)
        elif SINGLE_RE.match(token):
            found.extend(by_id[token]) if token in by_id else missing.append(token)
            if len(by_id.get(token, [])) > 1:
                ambiguous.append(token)
        else:
            missing.append(token)

    seen, unique = set(), []
    for t in found:
        if t.key not in seen:
            seen.add(t.key)
            unique.append(t)

    # Count a cross-reference once when the ticket it points at is also in scope.
    in_scope = {t.id for t in unique}
    counted = [t for t in unique if not (t.alias_of and t.alias_of in in_scope)]
    aliased = [t for t in unique if t not in counted]
    return counted, aliased, missing, ambiguous


def manual_kind(status: str) -> tuple[str, bool]:
    """Kind of a typed status, and whether it carried a mark."""
    for mark, name in STATUS_MARKS.items():
        if status.startswith(mark):
            return name, True
    for pattern, name in STATUS_WORDS:
        if pattern.match(status):
            return name, False
    return "unknown", False


def parse_rollout(tickets: list[Ticket], issues: Issues) -> tuple[list[str], list[WorkPackage]]:
    path = ROLLOUT / "rollout-wbs.md"
    if not path.exists():
        issues.error("rollout/rollout-wbs.md", "Rollout WBS missing.")
        return [], []
    lines = path.read_text(encoding="utf-8").splitlines()
    packages, phase = [], ""
    for i, line in enumerate(lines):
        heading = re.match(r"^##\s+(.*)", line)
        if heading:
            phase = heading.group(1).strip()
            continue
        if not re.match(r"^\|\s*RO-\d+\s*\|", line):
            continue
        row = cells(line)
        if len(row) < 6:
            issues.warn(f"rollout-wbs.md line {i + 1}", f"{row[0]} has {len(row)} columns, needs 6 (ID, name, scope, depends, launch, status). Skipped.")
            continue
        rid, name, scope, deps, launch, status = row[:6]
        scope = "" if scope in ("", "-", "none") else scope
        wp = WorkPackage(
            id=rid,
            name=name,
            scope=scope,
            depends=RO_RE.findall(deps),
            launch=launch,
            phase=phase,
            manual_status=status,
            line_no=i,
        )
        if scope:
            wp.resolved, wp.aliased, wp.missing, wp.ambiguous = resolve_scope(scope, tickets)
        packages.append(wp)
    return lines, packages


def find_cycles(packages: list[WorkPackage]) -> list[list[str]]:
    graph = {wp.id: [d for d in wp.depends] for wp in packages}
    state: dict[str, int] = {}  # 1 visiting, 2 done
    stack: list[str] = []
    cycles: list[list[str]] = []
    seen_sets: set[frozenset[str]] = set()

    def visit(node: str) -> None:
        state[node] = 1
        stack.append(node)
        for nxt in graph.get(node, []):
            if nxt not in graph or nxt == node:
                continue  # unknown and self dependencies are reported separately
            if state.get(nxt) == 1:
                loop = stack[stack.index(nxt):] + [nxt]
                key = frozenset(loop)
                if key not in seen_sets:
                    seen_sets.add(key)
                    cycles.append(loop)
            elif nxt not in state:
                visit(nxt)
        stack.pop()
        state[node] = 2

    for wp in packages:
        if wp.id not in state:
            visit(wp.id)
    return cycles


def check_packages(packages: list[WorkPackage], owners: dict[tuple[str, str], list[str]], issues: Issues) -> None:
    ids = [wp.id for wp in packages]
    known = set(ids)
    for rid in sorted({r for r in ids if ids.count(r) > 1}):
        issues.error(rid, "This RO ID is used by more than one row. Give each work package its own ID.")

    for wp in packages:
        for d in wp.depends:
            if d == wp.id:
                issues.error(wp.id, "Depends on itself.")
            elif d not in known:
                issues.error(wp.id, f"Depends on {d}, which doesn't exist, so {wp.id} can never start. Typo?")
        prose = [r for r in dict.fromkeys(RO_RE.findall(wp.name)) if r != wp.id and r not in wp.depends]
        if prose:
            issues.warn(wp.id, f"The name mentions {', '.join(prose)} but the Depends column doesn't, "
                               "so the dashboard may show it as ready too early. Add it, or split the package.")
        if wp.missing:
            issues.warn(wp.id, f"Scope entries not found: {', '.join(wp.missing)}")
        if wp.ambiguous:
            issues.warn(wp.id, f"Scope names {short_list(wp.ambiguous)}, used by tickets in more than one file, "
                               "so every one of them is counted. Use a heading scope such as API[EPIC 18] instead.")
        if wp.scope and not wp.resolved:
            issues.warn(wp.id, "Scope is set but matched no tickets, so its status can't be worked out.")
        if not wp.scope:
            kind, marked = manual_kind(wp.manual_status)
            if kind == "unknown":
                issues.warn(wp.id, f"Typed status '{wp.manual_status}' isn't recognised; start it with 🟩, 🟨 or 🟥.")
            elif not marked:
                issues.warn(wp.id, f"Typed status '{wp.manual_status}' has no mark; read as {kind}. Start it with {ICON[kind]}.")
        for t in wp.aliased:
            issues.note(wp.id, f"{t.id} ({t.source}) is a cross-reference to {t.alias_of}, so it is counted once.")

    for loop in find_cycles(packages):
        issues.error(loop[0], f"Dependency loop: {' → '.join(loop)}. None of these can start.")

    for (tid, src), rids in sorted(owners.items()):
        if len(rids) > 1:
            issues.warn(tid, f"In more than one work package ({', '.join(rids)}), so it counts in each. Source: {src}")


# ----------------------------------------------------------------- status

def derived_status(ts: list[Ticket]) -> str:
    if not ts:
        return "⬜ No tickets resolved"
    done = sum(t.status == "done" for t in ts)
    partial = sum(t.status == "partial" for t in ts)
    total = len(ts)
    if done == total:
        return f"🟩 Done ({done}/{total})"
    if done or partial:
        extra = f", {partial} partial" if partial else ""
        return f"🟨 In progress ({done}/{total} done{extra})"
    return f"🟥 Not started (0/{total})"


def status_for(wp: WorkPackage) -> str:
    if not wp.scope:
        return wp.manual_status
    status = derived_status(wp.resolved)
    if wp.missing:
        status += f" (not found: {', '.join(wp.missing)})"
    return status


def kind_for(wp: WorkPackage) -> str:
    if not wp.scope:
        return manual_kind(wp.manual_status)[0]
    return manual_kind(status_for(wp))[0]


def strip_mark(text: str) -> str:
    for mark in list(STATUS_MARKS) + ["⬜"]:
        text = text.replace(mark, "")
    return text.strip().lstrip(":").strip()


def clip(text: str, n: int) -> str:
    return text if len(text) <= n else text[: n - 1].rsplit(" ", 1)[0] + "…"


class State:
    """Status, readiness and priority for every work package, worked out once."""

    def __init__(self, packages: list[WorkPackage]) -> None:
        self.packages = packages
        self.text = {wp.id: status_for(wp) for wp in packages}
        self.kind = {wp.id: kind_for(wp) for wp in packages}

    def done(self, rid: str) -> bool:
        return self.kind.get(rid) == "done"

    def waiting_on(self, wp: WorkPackage) -> list[str]:
        return [d for d in wp.depends if not self.done(d)]

    def ready(self, wp: WorkPackage) -> bool:
        return not self.done(wp.id) and not self.waiting_on(wp)

    @staticmethod
    def priority_gap(wp: WorkPackage) -> bool:
        return any(t.id.startswith("PRI-") for t in wp.resolved + wp.aliased)

    @staticmethod
    def required(wp: WorkPackage) -> bool:
        return wp.launch.lower().startswith("required")

    def priority(self, wp: WorkPackage) -> tuple[int, int]:
        # Priority rule (rollout-wbs.md): priority gaps first, then other
        # launch-required work, then everything else.
        en, req = self.priority_gap(wp), self.required(wp)
        return (0 if en and req else 1 if req else 2 if en else 3, self.packages.index(wp))


# ---------------------------------------------------------------- outputs

def write_if_changed(path: Path, text: str) -> None:
    if not path.exists() or path.read_text(encoding="utf-8") != text:
        path.write_text(text, encoding="utf-8")


def write_rollout(lines: list[str], packages: list[WorkPackage]) -> None:
    lines = list(lines)
    for wp in packages:
        if not wp.scope:
            continue  # typed by hand: leave the row as written
        row = cells(lines[wp.line_no])
        row[5] = status_for(wp)
        lines[wp.line_no] = "| " + " | ".join(row) + " |"
    write_if_changed(ROLLOUT / "rollout-wbs.md", "\n".join(lines) + "\n")


def linked_groups(tickets: list[Ticket]) -> list[list[str]]:
    """Tickets that name each other (overlap, supersedes, carries forward,
    or depends on). Likely the same piece of work tracked in two places,
    or work that has to ship together. Union-find over the mentions."""
    known = {t.id for t in tickets}
    parent: dict[str, str] = {}

    def find(x: str) -> str:
        parent.setdefault(x, x)
        while parent[x] != x:
            parent[x] = parent[parent[x]]
            x = parent[x]
        return x

    for t in tickets:
        for other in t.mentions & known:
            # Only link across areas: same-file mentions are usually plain
            # ordering notes, not the same task tracked twice.
            if other.split("-")[0] != t.id.split("-")[0]:
                parent[find(t.id)] = find(other)

    groups: dict[str, list[str]] = defaultdict(list)
    for x in list(parent):
        groups[find(x)].append(x)
    return sorted((sorted(g, key=id_key) for g in groups.values() if len(g) > 1), key=lambda g: id_key(g[0]))


def id_clashes(tickets: list[Ticket]) -> dict[str, list[Ticket]]:
    dupes: dict[str, list[Ticket]] = defaultdict(list)
    for t in tickets:
        dupes[t.id].append(t)
    return {k: v for k, v in dupes.items() if len(v) > 1}


def write_all_tickets(tickets: list[Ticket], today: str, owners: dict[tuple[str, str], list[str]]) -> None:
    out = [
        "# All tickets",
        "",
        f"Generated by `rollout/collate.py` on {today}. Don't edit by hand: change the ticket in its home file and re-run.",
        "",
        "Legend: 🟩 done · 🟨 partial · 🟥 pending · ⬜ no status in its home file",
        "",
    ]

    clashes = id_clashes(tickets)
    if clashes:
        out += [
            "## ID clashes",
            "",
            "These IDs are used by more than one ticket. Scope references to them match every one listed.",
            "",
            "| ID | Where |",
            "|---|---|",
        ]
        for tid in sorted(clashes, key=id_key):
            where = "; ".join(f"{t.source} ({t.epic[:40]})" for t in clashes[tid])
            out.append(f"| {tid} | {where} |")
        out.append("")

    groups = linked_groups(tickets)
    if groups:
        out += [
            "## Linked across WBS files",
            "",
            "Tickets in different WBS files that name each other: the same work tracked twice, "
            "or work that has to ship together. The rollout WBS gives each group one RO item.",
            "",
        ]
        out += ["- " + ", ".join(g) for g in groups]
        out.append("")

    for area, rel in SOURCES:
        area_tickets = [t for t in tickets if t.source == rel]
        if not area_tickets:
            continue
        out += [f"## {area}", "", f"Source: [`{rel}`](../{rel})", ""]
        epics: dict[str, list[Ticket]] = defaultdict(list)
        for t in area_tickets:
            epics[t.epic].append(t)
        for epic, ts in epics.items():
            out += [f"### {epic}", "", "| ID | Ticket | Status | Rollout |", "|---|---|---|---|"]
            for t in ts:
                title = t.title.replace("|", "\\|")
                rollout = ", ".join(owners.get((t.id, t.source), [])) or ""
                out.append(f"| {t.id} | {title} | {ICON[t.status]} | {rollout} |")
            out.append("")
    write_if_changed(ROLLOUT / "all-tickets.md", "\n".join(out) + "\n")


def bar(done: int, partial: int, total: int, width: int = 20) -> str:
    if total == 0:
        return ""
    d = round(width * done / total)
    p = round(width * partial / total)
    return "█" * d + "▒" * p + "·" * (width - d - p)


def issues_md(issues: Issues) -> list[str]:
    out = ["## Checks", ""]
    shown = [i for i in issues.items if i["level"] != "info"]
    if not shown:
        return out + ["No problems found.", ""]
    out += [
        f"{plural(issues.count('error'), 'error')}, {plural(issues.count('warning'), 'warning')}. "
        "Errors make the numbers above wrong; fix them first.",
        "",
        "| Level | Where | Problem |",
        "|---|---|---|",
    ]
    for i in sorted(shown, key=lambda i: i["level"] != "error"):
        out.append(f"| {i['level']} | {i['where']} | {i['msg'].replace('|', '/')} |")
    return out + [""]


def write_progress(tickets: list[Ticket], st: State, today: str,
                   owners: dict[tuple[str, str], list[str]], issues: Issues) -> None:
    packages = st.packages
    out = [
        "# Rollout progress",
        "",
        f"Generated by `rollout/collate.py` on {today}. Re-run after updating any WBS.",
        "",
        "Plan: [`rollout-wbs.md`](rollout-wbs.md) · Every ticket: [`all-tickets.md`](all-tickets.md)",
        "",
    ]
    if issues.count("error"):
        out += [f"> **{plural(issues.count('error'), 'error')} found.** See Checks at the end.", ""]

    required = [wp for wp in packages if st.required(wp)]
    req_done = sum(st.done(wp.id) for wp in required)
    out += [
        "## Launch readiness",
        "",
        f"**{req_done} of {len(required)} launch-required work packages done.**",
        "",
        "| Phase | Work packages | Done | In progress | Not started | No status |",
        "|---|---|---|---|---|---|",
    ]
    phases: dict[str, list[WorkPackage]] = defaultdict(list)
    for wp in packages:
        phases[wp.phase].append(wp)
    for phase, wps in phases.items():
        k = [st.kind[w.id] for w in wps]
        out.append(f"| {phase} | {len(wps)} | {k.count('done')} | {k.count('partial')} | {k.count('pending')} | {k.count('unknown')} |")
    out.append("")

    ready = sorted((wp for wp in packages if st.ready(wp)), key=st.priority)
    blocked = [wp for wp in packages if not st.done(wp.id) and not st.ready(wp)]
    out += [
        "## Can start or continue now",
        "",
        "Not done, and everything they depend on is done. In priority order: priority gaps first "
        "(marked **PRI**), then other launch-required work, then the rest.",
        "",
    ]
    out += [
        f"- {'**PRI** ' if st.priority_gap(wp) else ''}**{wp.id}** {wp.name} ({wp.launch}): {st.text[wp.id]}"
        for wp in ready
    ] or ["- Nothing."]
    out += ["", "## Waiting on a dependency", ""]
    for wp in blocked:
        out.append(f"- **{wp.id}** {wp.name}: waiting on {', '.join(st.waiting_on(wp))}")
    if not blocked:
        out.append("- Nothing.")
    out.append("")

    out += [
        "## All tickets by area",
        "",
        "| Area | Done | Partial | Pending | No status | Total | |",
        "|---|---|---|---|---|---|---|",
    ]
    totals = defaultdict(int)
    for area, rel in SOURCES:
        ts = [t for t in tickets if t.source == rel]
        c = {k: sum(t.status == k for t in ts) for k in ("done", "partial", "pending", "unknown")}
        for k, v in c.items():
            totals[k] += v
        out.append(
            f"| [{area}](../{rel}) | {c['done']} | {c['partial']} | {c['pending']} | {c['unknown']} | {len(ts)} | "
            f"`{bar(c['done'], c['partial'], len(ts))}` |"
        )
    n = len(tickets)
    out.append(
        f"| **Total** | **{totals['done']}** | **{totals['partial']}** | **{totals['pending']}** | "
        f"**{totals['unknown']}** | **{n}** | `{bar(totals['done'], totals['partial'], n)}` |"
    )
    out += [
        "",
        "`█` done · `▒` partial · `·` pending or no status. Ticket counts aren't weighted by size: "
        "the rollout table above is the better measure of how close launch is.",
        "",
    ]
    unplaced = [t for t in tickets if t.status != "done" and (t.id, t.source) not in owners]
    out += [
        "## Open tickets not in the rollout plan",
        "",
        f"{len(unplaced)} open tickets aren't in any RO work package yet. Either add them to one, "
        "or decide they're out of scope and close them in their home WBS.",
        "",
        "| Area | Open, not in a work package |",
        "|---|---|",
    ]
    for area, rel in SOURCES:
        n_open = sum(t.source == rel for t in unplaced)
        if n_open:
            ids = ", ".join(t.id for t in unplaced if t.source == rel)
            out.append(f"| {area} | {n_open}: {ids} |")
    out.append("")
    out += issues_md(issues)
    write_if_changed(ROLLOUT / "PROGRESS.md", "\n".join(out) + "\n")


def write_data(tickets: list[Ticket], st: State, today: str,
               owners: dict[tuple[str, str], list[str]], issues: Issues) -> None:
    """data.js for index.html. A script (not JSON) so the page also works
    opened straight from disk, where fetch() is blocked."""
    clashes = id_clashes(tickets)
    data = {
        "generated": today,
        "areas": [
            {"name": area, "source": rel,
             **{k: sum(t.source == rel and t.status == k for t in tickets) for k in ("done", "partial", "pending", "unknown")},
             "total": sum(t.source == rel for t in tickets)}
            for area, rel in SOURCES
        ],
        "packages": [
            {
                "id": wp.id,
                "name": wp.name,
                "phase": wp.phase,
                "launch": wp.launch,
                "depends": wp.depends,
                "status": st.kind[wp.id],
                "statusText": strip_mark(st.text[wp.id]),
                "priority": st.priority_gap(wp),
                "ready": st.ready(wp),
                "waitingOn": st.waiting_on(wp),
                "tickets": [t.key for t in wp.resolved],
                "aliases": [t.key for t in wp.aliased],
            }
            for wp in st.packages
        ],
        "tickets": [
            {
                "key": t.key,
                "id": t.id,
                "title": t.title,
                "status": t.status,
                "note": clip(strip_mark(t.status_text), 320),
                "area": t.area,
                "epic": t.epic,
                "rollout": owners.get((t.id, t.source), []),
                **({"aliasOf": t.alias_of} if t.alias_of else {}),
                **({"clash": True} if t.id in clashes else {}),
            }
            for t in tickets
        ],
        "issues": [i for i in issues.items if i["level"] != "info"],
    }
    body = json.dumps(data, ensure_ascii=False, indent=1)
    write_if_changed(
        ROLLOUT / "data.js",
        f"// Generated by rollout/collate.py on {today}. Don't edit: change the home WBS and re-run.\n"
        f"window.ROLLOUT = {body};\n",
    )


# ------------------------------------------------------------------- main

def main() -> int:
    ap = argparse.ArgumentParser(description="Collate the WBS files into the rollout tracker.")
    ap.add_argument("--check", action="store_true", help="check only; write nothing; exit 1 on errors")
    ap.add_argument("--strict", action="store_true", help="with --check, warnings also fail")
    ap.add_argument("--quiet", action="store_true", help="print errors and warnings only")
    args = ap.parse_args()

    issues = Issues()
    today = dt.date.today().isoformat()
    tickets = [t for area, rel in SOURCES for t in parse_source(area, rel, issues)]
    lines, packages = parse_rollout(tickets, issues)
    owners: dict[tuple[str, str], list[str]] = defaultdict(list)
    for wp in packages:
        for t in wp.resolved + wp.aliased:
            owners[(t.id, t.source)].append(wp.id)
    check_packages(packages, owners, issues)
    clashes = id_clashes(tickets)
    if clashes:
        issues.warn("IDs", f"{len(clashes)} ticket IDs are used by more than one ticket: {short_list(sorted(clashes, key=id_key))}. "
                           "See ID clashes in all-tickets.md; renumber one side when you can.")
    st = State(packages)

    if not args.check and packages:
        write_rollout(lines, packages)
        write_all_tickets(tickets, today, owners)
        write_progress(tickets, st, today, owners, issues)
        write_data(tickets, st, today, owners, issues)

    if not args.quiet:
        print(f"{len(tickets)} tickets from {len(SOURCES)} WBS files, {len(packages)} rollout work packages.")
        print(f"{len(owners)} source tickets are covered by a work package.")
    order = {"error": 0, "warning": 1, "info": 2}
    for i in sorted(issues.items, key=lambda i: order[i["level"]]):
        if i["level"] == "info" and args.quiet:
            continue
        print(f"  {i['level']}: {i['where']}: {i['msg']}")
    errors, warnings = issues.count("error"), issues.count("warning")
    if not args.quiet or errors or warnings:
        print(f"{plural(errors, 'error')}, {plural(warnings, 'warning')}." + ("" if args.check else " Outputs written." if packages else ""))

    if args.check and (errors or (args.strict and warnings)):
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
