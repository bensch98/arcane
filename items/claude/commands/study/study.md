---
description: "Familiarise yourself with a set of files or a topic without making changes. Read-only context loading. Can be prefixed by a scoping command (e.g. /scope-by-staged /study) or given a topic/path (e.g. /study computer-vision/builder)."
---

Read and internalise the in-scope files (or topic) so that follow-up questions and edits in this conversation can be answered from real context — not guesses. **Do not modify any files.** This is a read-only command.

## Steps

### 1. Determine study scope

**Mode A — A `--- file list ---` block exists in the current conversation context** (provided by a scoping command like `/scope-by-staged`, `/scope-by-unstaged`, `/scope-by-recent-changes`, `/scope-by-large-files`, `/scope-by-flagged-todos`):
Use exactly those files. Ignore any topic in `$ARGUMENTS`.

**Mode B — `$ARGUMENTS` contains `--dir=<path>`:**
Use that directory as the root. Walk it and pick relevant source files (`.ts`, `.svelte`, `.svelte.ts`, `.js`, `.py`, `.md`). Skip `node_modules`, `.svelte-kit`, `__pycache__`, `dist`, `build`, `static`.

**Mode C — `$ARGUMENTS` is a path or topic** (e.g. `computer-vision/builder`, `data-engine`, `ISA95 namespace hashing`):

- If it resolves to a directory under `ui/src/` or `backend/`, treat it as Mode B.
- Otherwise treat it as a topic: use Grep/Glob to locate the most relevant files (route, components, services, types, tests, related docs), then study those.

**Mode D — No scope provided:**
Ask the user what they want you to study (a path, a feature area, or a topic). Do not guess.

### 2. Read the files

Read every in-scope file in full. For directories, also read sibling files that are clearly part of the same unit (`index.ts`, `types.ts`, `ctx.svelte.ts`, `+page.svelte`, `+layout.ts`, etc.) even if they aren't in the initial list — context matters more than a strict file boundary.

If the scope references types, APIs, or stores that live elsewhere, follow one hop outward to read those definitions too. Stop after one hop — don't crawl the whole codebase.

### 3. Summarise what you learned

Produce a short, structured summary so the user can confirm you understood correctly. Keep it tight — this is orientation, not a report.

```
## Studied
<list of files actually read, grouped by area if many>

## What this code does
2–5 sentences on the purpose and shape of the scope. Name the entry point(s).

## Key types / contracts
Bullet list of the main types, props, context shapes, or API signatures a future edit would need to respect.

## Notable patterns or gotchas
Anything non-obvious: invariants, ordering requirements, side effects, perf constraints, TODOs, known bugs, deviations from the project conventions in CLAUDE.md.

## Open questions
Things the code alone doesn't answer that the user may need to clarify before edits.
```

If nothing notable exists for a section, omit the section rather than padding it.

### 4. Stop

Do **not** propose edits, refactors, or follow-up commands unless the user asks. The point of `/study` is to load context, not to act on it. End by inviting the next instruction:

> Ready. What would you like to do with this?
