<!-- Written by ajiya. Do not edit by hand: use ajiya commands. -->
# Dashboard views

Goal: The dashboard shows each thing once, explains every phase, and shows how the plan changed over time

| ID | App | Ticket | Done when | Depends | Tags | Status |
|---|---|---|---|---|---|---|
| AJ-0107 | ajiya | Phase descriptions | phase add takes --description and a new phase describe <slug> "<text>" sets it; the description is kept as a paragraph under Goal in the phase file, round-trips with goldens, appears in phase list --json, ajiya build data and an MCP tool | AJ-0048 | - | 🟩 Done · 3e35786 · 2026-10-16 · tests passed |
| AJ-0108 | ajiya | Dashboard: one view per way of seeing tickets | Next is folded into Board's Can start now column, and Phases and Tickets become one Tickets view with a group-by-phase toggle; no two tabs list the same tickets in the same grouping; old #next and #phases links land on the new view; dashboard tests updated | AJ-0057 | - | 🟩 Done · 611bd42 · 2026-10-16 · tests passed |
| AJ-0109 | ajiya | Dashboard: a page per phase | Clicking a phase opens its page with goal, description, milestone, progress, tickets grouped by state, and the phases it depends on and that depend on it; the page has its own link (#phase/<slug>); dashboard test covers a phase with and without a description | AJ-0107, AJ-0108 | - | 🟨 In progress |
| AJ-0110 | ajiya | Dashboard: plan history from decisions and changes | A History view lists decisions and approved or rejected changes in date order; each change shows what it superseded and created and links to those tickets, and each decision shows what it replaced or was replaced by; ajiya build data carries the history; dashboard test covers a plan with one approved change | AJ-0057, AJ-0108 | - | 🟨 In progress |
