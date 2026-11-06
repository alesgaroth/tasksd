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
