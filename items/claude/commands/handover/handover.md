# Generate Module Handover Document

You are generating a professional handover document for a codebase module, section, or group of files. The goal is to give any engineer — frontend, backend, new team member, or external collaborator — a complete mental model of the target without having to read all the source code themselves.

**Arguments:** `$ARGUMENTS`

Parse the arguments:

- The first token is the target path (relative to the repo root, e.g. `ui/src/components/computer-vision`, `src/api/auth`, `services/billing`). It may also be a glob or a comma-separated list of paths.
- The optional flag `--audience=<role>` (e.g. `--audience=backend`, `--audience=frontend`, `--audience=new-hire`) tilts emphasis toward what that role needs most. Default: general engineering audience.
- The optional flag `--out=<path>` sets the output file. If omitted, derive it: use the last path segment of the target as the filename and write to `docs/<segment>/overview.md`.

---

## Your task

### Step 1 — Read the target thoroughly

Find and read, in this order:

1. Type/interface/schema files — these define the contracts
2. Registry, manifest, or configuration files — these define the domain model
3. Core logic files — algorithms, state machines, executors, services
4. Persistence/serialization files — what gets stored or sent over the wire
5. Test files — they document invariants and non-obvious edge cases
6. Any existing docs in or near the target directory

Use `find`, `grep`, or directory listing to discover files if the structure isn't obvious. Read the full content of every file that looks load-bearing, not just excerpts.

### Step 2 — Form a mental model

Before writing, answer these questions internally:

- What is this module's single responsibility?
- What are the 3–5 most important concepts a reader must understand?
- What are the entry points (where does execution/data flow start)?
- What are the external boundaries (what does it consume from, or expose to, other modules)?
- What would trip up a new engineer in the first week?

### Step 3 — Write the document

Adapt the section structure to what the module actually contains. Not every section applies to every module — skip or merge sections that don't fit, and add domain-specific ones that do. The structure below is a starting point, not a rigid template.

---

## Document structure

```markdown
# <Module Name> — Overview

> **TL;DR** (2–4 sentences): what this module does, why it exists, and the one mental model
> the reader must hold before reading further.

## 1. Repo layout

File/directory tree with one-line annotations. Mark which files are:

- "contract" (types/schemas other modules depend on)
- "core logic" (the important algorithms)
- "UI-only" / "infra-only" / "test-only" (safe to skip for most audiences)

## 2. Core concepts

One subsection per key concept (3–6 concepts). Each subsection:

- Names the concept and defines it precisely
- Shows the TypeScript type or data shape (inline comment every field)
- Explains the invariants and constraints
- Links to the file:line where it's defined

## 3. Data model / schema

The primary data structure(s) in full. Use annotated TypeScript blocks.
Mark required vs. optional fields. Note discriminated unions and how to interpret the `kind`/`type` field.
If there's a version field, explain what bumping it means.

## 4. How it works — execution flow

Walk through the happy path end-to-end:
input → [step] → [step] → output

Use numbered steps. For each step, name the file and function responsible.
Include the error path: how failures surface, what errors look like, who handles them.

## 5. Key algorithms

For each non-trivial algorithm:

- **Name** and one-line description
- File:line reference
- Pseudocode or a short excerpt for the non-obvious part
- Edge cases and gotchas

## 6. External boundaries

What this module imports from other modules, and what it exports to them.
For each boundary: the data shape, the direction, and any contracts or version constraints.
If it talks to an API or backend service: request/response shape, auth model, error contract.

## 7. State and persistence

What state is kept (in-memory, local storage, cache, DB).
Cache key construction (if any). Invalidation strategy.
What survives a page reload / process restart vs. what doesn't.

## 8. Extension points

How to add a new function/handler/node/type to the registry (if applicable).
What hooks or interfaces to implement. What tests to write.

## 9. Known limitations and gotchas

Bullet list: design decisions that look wrong but aren't, known gaps, performance traps,
temporary workarounds still in the code, and anything that regularly surprises new contributors.

## 10. Getting started checklist

A concrete checklist for someone picking up this module for the first time:

- [ ] Files to read first, in order
- [ ] Tests to run to see the module in action
- [ ] First safe change to make as a warm-up
- [ ] Questions to answer before making any significant change
```

---

## Quality bar

- **Be specific.** Every non-trivial claim cites a file path and line number.
- **Be complete on contracts.** The reader should not need to open TypeScript files to know the full data shape.
- **Be proportional.** Spend more words on complex/surprising parts, fewer on obvious ones.
- **Be honest about gaps.** If something is unfinished, undocumented, or poorly understood, say so explicitly.
- **Write for the stated audience.** If `--audience=backend` was given, emphasize the wire format, server-side replication needs, and API contracts. If `--audience=new-hire`, emphasize mental models and onboarding order. Default to a general senior-engineer audience.
- **No padding.** Every sentence should carry information the reader needs.

Write the document now, then report the output path to the user.
