# Vibe Coding Plan: Git-Backed Task System

Nov 5, 2026 · @Alan Linton

## What we're building

A local Go website that keeps your tasks as markdown files in a private git repo, so one tool can replace Simpleology at home and later the notebook at work.

It keeps what you like from each tool and fixes what you don't:

- **Hidden by default:** Today shows only what you chose; everything else stays findable but out of sight (Simpleology's best idea, and what Trello lacks).
- **Items change size:** promote a task to a project or a target, or demote it, without recreating it.
- **Simple items, real links:** small files with few fields, plus parent/child and typed links (blocks, related, duplicate) so blockers are first-class.
- **Drop a focus, drop its work:** marking a Target or Project dropped cascades to everything under it and logs why.
- **A record of each day:** what you worked on (your `=`), what you finished, and the links you used.
- **Nudges toward focus:** 3-3-3, the 25-5 rule and column limits, tuned by trial.

## Ground rules

The files are the database, and the program is a thin layer that reads them, applies pure logic, and writes them back.

- **Stack:** Go standard library (`net/http`, `html/template`), HTMX for the interactive parts, no web framework. Allow one small YAML library; call the `git` binary for commits.
- **Functional core, imperative shell:** `internal/core` holds all decisions as pure functions over plain values: parsing, status changes, promote/demote, carry-over, view selection, nudge rules. It never touches files, the clock or HTTP; today's date is a parameter.
- **Test-driven core:** write the failing test first, aim for 100% coverage of `core`, and have the build fail below an agreed threshold.
- **Shell stays thin:** `store` reads and writes files and runs git; `web` turns requests into core calls and renders templates. A later `tk` CLI or TUI is just another shell over the same core.
- **One commit per change:** every edit in the UI becomes a git commit with a plain message, so history is the audit trail.
- **Separate data, same program:** the program takes a repo path. Home data lives in one private repo; work data (later) in another, never shared. The app's own source is a third repo.
- **Dev tickets live with the code:** software repos hold their own `tickets/` folder in the same format, so a ticket and its change are committed together.

* **Typed operations, validated writes:** every change (set status, promote, drop, add a work-log note) is a named operation in `core` that validates the result before anything is written. Hand edits still work, and a `validate` command in a pre-commit hook catches mistakes.

## Data model

One item type with a `level` and a `parent`, so changing size is an edit, not a rebuild.

| Level | Meaning | Typical parent |
| --- | --- | --- |
| Target | An outcome you want, kept to the 25-5 list | none |
| Project | A finite effort toward a target | Target |
| Task | Something to do in a day or a few sittings | Project or Target |
| Subtask | A step inside a task | Task |

Promote or demote changes `level` and `parent` in one operation, and the item's children come along. The core enforces the four levels; the depth limit is a config value, so a fifth level later is a one-line change.

An item is one file, for example `items/T-0042.md`:

```markdown
---
id: T-0042
title: Draft sync section 3
level: task
parent: P-0007
status: active      # inbox | active | waiting | done | dropped
priority: A         # A-E, from the ABCDE method
urgent: false       # with important, places it in the Eisenhower matrix
important: true
moscow: should      # optional
tags: [frontier]
remind: 2026-11-12  # when it should reappear
waiting_on: Joseph  # only when status is waiting
links:
  - {type: blocks, to: T-0050}
created: 2026-11-05
---
Free-form notes.
```

Other record types:

- **Daily record** (`days/2026-11-05.md`): the items you planned (a `*` marks top priority), the items you worked on with a note of what you did and what it waits for (your `=`), the items finished, and a list of links for the day.
- **Carry-over rule:** a worked-on, unfinished item with no `remind` date appears at the start of the next day; with one, it appears on that date.
- **Not-doing log:** every dropped item keeps a `reason` and date, and the log lists them with a one-tap way to add a declined idea that never became an item.
- **Habits** (`habits/*.md`): recurring items with a cadence, shown in Today and in the weekly review.
- **Reviews** (`reviews/2026-W45.md`): written from editable prompt files, weekly on Sundays and yearly.

### Work log, archive, playbooks and forecasts

- **Work log, written twice:** each `=` note is saved on the item as a dated entry in its Work Log (its history, which you may correct) and in that day's record (immutable). A task or project shows when and what you did on it; a week view shows what you worked on each day.
- **Archive:** done and dropped items can be moved to `items/archive/`, or simply deleted and left in git history. Ids and links keep resolving across both folders, and active views ignore the archive unless you search.
- **Playbooks:** a file in `playbooks/` lists the steps of a repeatable process (a release, a trip, onboarding). Adding one creates a single task with a subtask per step.
- **Software tickets:** in dev repos an item body has fixed headings (Objective, Context, Acceptance Criteria, Plan, Work Log) so a tool can edit one section by name. Personal items such as "get a haircut" stay a title plus notes.
- **Forecast fields (optional):** `size` (xs to xl) and `due`. Config maps each size to a range of work days. Projected finish is computed from sizes, blocking links, priority and serial capacity, and never stored. Later, the core compares actual days worked on done items with the range for their size and suggests updating the config.

## Repo layout

Three kinds of repo: the app, your data, and any software project that carries its own tickets. Start with the app and one data repo.

```
tasks-app/                 # the program (its own tickets live in tickets/)
  CLAUDE.md
  cmd/tasksd/main.go       # web shell
  cmd/tk/                  # later: CLI or TUI shell
  internal/core/           # pure logic, ~100% test coverage
  internal/store/          # files + git, kept thin
  internal/web/            # handlers, templates, HTMX endpoints
  web/templates/  web/static/   # htmx.js vendored here
  testdata/

life/                      # private data repo (home)
  config.yaml4              # level names, caps, nudge mode
  items/                   # one file per item, status in frontmatter
  days/                    # daily records and link lists
  habits/
  reviews/
  prompts/                 # weekly and yearly review questions

some-software-repo/        # e.g. Mauve, Frontier
  tickets/                 # same item format, committed with the code
```

Two choices worth knowing:

- **Flat `items/`, status in frontmatter:** promoting an item or moving it across a board is an edit to one file, so its id and history stay put. Folders per status would turn every drag into a rename.
- **Several repos, one site:** `tasksd --repo life=~/life --repo mauve=~/src/mauve` shows both as workspaces, and each change is committed to the repo that owns the item. Commits reference tickets with a trailer such as `Ticket: T-0042`, so the site can show the commits for an item.

## Views

Every method from this week's reading is a different lens on the same items, chosen by tags rather than by separate lists.

| View | What it answers | Fields it reads |
| --- | --- | --- |
| Today | What am I doing now? Hides everything else | planned list, `remind`, carry-over |
| Inbox | What arrived and needs sorting? | status `inbox` |
| Eisenhower | Urgent or important? Drag between quadrants | `urgent`, `important` |
| ABCDE | Must, should, nice, delegate, eliminate | `priority` |
| MoSCoW | Scope for a project | `moscow` |
| Kanban | Where is each item in its flow? Column limits apply | `status` |
| Targets (25-5) | Five focus targets, the other twenty held back | level `target`, `focus` |
| Day record | What did I do, and which links did I use? | `days/` |
| Not doing | What did I decide against, and why? | status `dropped`, `reason` |
| Review | Weekly and yearly prompts with the week's data beside them | `reviews/`, `habits/` |

Dragging in the Eisenhower or Kanban view changes fields on the item and makes a commit. Pareto and the "yes to one thing is no to another" idea show up as nudges, not views.

Much later (a C): a calendar with Gantt-style bars, past work drawn from the day records and projected finishes from the forecast.

## Nudges and A/B testing

The core exposes the nudge rules as pure functions that return violations, and a config value picks how the shell reacts, so trying both versions costs almost nothing.

Rules to start with:

- Today holds at most 3 important, 3 maintenance and 3 urgent items (the 3-3-3 method).
- At most 5 focus Targets; the other 20 live in a held-back list (25-5).
- Each Kanban column has an item limit.

Three ways to react, set by `nudge:` in `config.yaml`:

1. **banner:** a visible warning when a rule is broken, nothing blocked.
2. **cap:** the add is refused until you choose what to bump, and the bumped item goes to the not-doing log with a reason.
3. **soft cap:** the add goes through only after you type why this beats something already there. This is a middle option for you to test, since a hard cap felt too hard and a banner too easy to ignore.

To compare them, the shell appends one line per event to `nudges.log` (rule broken, overridden, bumped). Run each mode for a week or two, then compare how long Today got, how many items carried over, and how often you overrode it. The weekly review prompt asks which felt better.

## Build order

Each slice ends with something you can use, and each starts with failing tests in `internal/core`. Stop after any slice and you still have a working tool.

1. **Hello world:** repo, `CLAUDE.md`, `go.mod`, one HTMX page served by `tasksd`, a `make test` target that checks coverage. Done when the page loads and the test target passes.
2. **Items and validation:** parse and write item files with a strict schema (round-trip tests), a `validate` command and pre-commit hook, and list items read-only from a data repo. All writes go through core operations from here on.
3. **Capture:** an inbox form that writes an item file and makes a git commit. Test the shell against a temporary repo.
4. **Today and the daily record:** pick today's list, mark done, write a `=` note to both the item's Work Log and the day record, apply the carry-over rule. You can now replace the notebook at home.
5. **Hierarchy and archive:** parent and level, promote and demote, cascading drop with a reason into the not-doing log, and archiving.
6. **Links and waiting:** typed links, blocked indicators, `waiting_on` with a `remind` date.
7. **Playbooks:** a template that creates one task with a subtask per step.
8. **Lenses:** Eisenhower, ABCDE, then Kanban with drag via HTMX.
9. **Nudges:** the three rules, the three modes, `nudges.log`.
10. **Habits and reviews:** recurring items, Sunday review prompts, due-today reminders.
11. **Forecast (optional):** sizes mapped to day ranges, serial capacity, projected finish, a what-if check, and calibration from your work logs.
12. **Second shell and second repo:** a `tk` CLI over the same core, the multi-repo flag for software tickets, and later the calendar with Gantt bars.

Prompt pattern for every slice: "Slice N. Write the failing core tests first and show me them. Then implement until green, then wire the thinnest possible shell. Keep `core` free of I/O and time. Don't add anything outside this slice."

## Kickoff prompt and CLAUDE.md

Put this in `CLAUDE.md` at the root of `tasks-app` before the first session, so every later session starts with the same rules:

```markdown
# tasks-app

Local Go web app for personal task planning. Items are markdown files with
YAML frontmatter in a private git repo; every change is one git commit.

## Architecture
- Functional core, imperative shell.
- internal/core: pure functions over plain values. No file, network, git or
  clock access; the current date is always a parameter.
- internal/store: reads/writes files and runs git. As thin as possible.
- internal/web: net/http handlers + html/template + HTMX. No framework.
- Keep logic out of handlers so a CLI/TUI can reuse core unchanged.

## Rules
- Test first. Write the failing test, show it, then implement.
- Target 100% coverage of internal/core; `make test` fails below the threshold.
- Minimal dependencies: stdlib plus one YAML library. Ask before adding more.
- Four levels only: target, project, task, subtask. Depth is a config value.
- All writes go through named core operations that validate first.
- Software tickets use fixed headings (Objective, Context, Acceptance
  Criteria, Plan, Work Log); personal items are a title plus notes.
- Work one slice at a time; do not build ahead.

## Data
- Item format: see docs/item-format.md (frontmatter fields and statuses).
- Dropping an item requires a reason and cascades to its children.
```

Then start the first session with:

```markdown
Read CLAUDE.md. We are doing Slice 1 only: a hello-world HTMX page served
by tasksd, plus a Makefile with test and coverage targets. Before writing
code, propose the directory layout and the first failing test. Do not add
features beyond the page and the build setup.
```

## Deferred and open questions

- **Reddit post:** the post summarizes two articles by the same author, which I read. Adopted: typed write operations, strict validation, fixed headings for software tickets, and ordering kept out of git.
- **Reminders:** on a laptop, the first version can be a "due today" list on load plus a browser notification while the page is open. Real push notifications wait for mobile.
- **Mobile (a C):** later options include reaching the laptop's site over a private network, or a quick-add that only appends to the inbox.
- **Deeper levels:** the `parent` pointer already allows them; the four-level limit is a validation rule, not a data-format limit.
- **Import:** manual and later, perhaps a `tk add` command. Nothing is needed from Simpleology, Trello or the notebook for the first slices.
- **Concurrent edits:** if you edit files by hand or Claude edits them while the site runs, the site should reread from disk and treat git as the source of truth.
- **Work copy:** the same binary pointed at a work-only repo, with no shared data, once the home version has proved itself.

* **Ordering within a Kanban column:** keep it in a gitignored local file, so dragging doesn't rewrite item files or create merge noise.
* **Branches in dev repos:** a ticket's status on a feature branch can differ from `main`; decide which one the site shows before Slice 12.
* **Forecast scope:** single person, serial work. Team capacity and Gantt charts wait.
