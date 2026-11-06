# Item Format

Items in `tasksd` are stored as individual Markdown files inside `items/` (or `tickets/` in software repositories).

## File Naming

Files are named `<id>.md`, where `<id>` matches the frontmatter ID (e.g., `T-0042.md`, `P-0007.md`).

## Frontmatter Schema

Each item starts with a YAML frontmatter block enclosed by `---`:

```markdown
---
id: T-0042
title: Draft sync section 3
level: task
parent: P-0007
status: active
priority: A
urgent: false
important: true
moscow: should
tags:
  - frontier
remind: 2026-11-12
waiting_on: Joseph
reason: Deprecated in favor of T-0055
links:
  - type: blocks
    to: T-0050
created: 2026-11-05
---
Free-form notes or structured sections.
```

## Fields and Validation Rules

| Field | Type | Required | Allowed Values / Format | Notes |
|---|---|---|---|---|
| `id` | string | Yes | `^[A-Za-z]+-[0-9]+$` | Unique identifier, e.g. `T-0042` |
| `title` | string | Yes | Non-empty string | Summary of the item |
| `level` | enum | Yes | `target`, `project`, `task`, `subtask` | Hierarchy level |
| `parent` | string | Conditional | Valid item ID | `target` cannot have parent; `subtask` must have parent |
| `status` | enum | Yes | `inbox`, `active`, `waiting`, `done`, `dropped` | Current lifecycle status |
| `priority` | enum | No | `A`, `B`, `C`, `D`, `E` | ABCDE method priority |
| `urgent` | bool | No | `true`, `false` | Eisenhower matrix urgency (default false) |
| `important` | bool | No | `true`, `false` | Eisenhower matrix importance (default false) |
| `moscow` | enum | No | `must`, `should`, `could`, `wont` | MoSCoW prioritization |
| `tags` | list | No | List of strings | Categorization tags |
| `remind` | string | No | `YYYY-MM-DD` | Date when item should reappear |
| `waiting_on` | string | Conditional | Non-empty string | Required if `status: waiting`; prohibited otherwise |
| `reason` | string | Conditional | Non-empty string | Required if `status: dropped` |
| `links` | list | No | `[{type, to}]` | Link `type` must be `blocks`, `related`, or `duplicate` |
| `created` | string | Yes | `YYYY-MM-DD` | Creation date |

## Body

The body contains Markdown text following the closing `---`.
- Personal items: Free-form notes.
- Software tickets: Fixed headings (`Objective`, `Context`, `Acceptance Criteria`, `Plan`, `Work Log`).
