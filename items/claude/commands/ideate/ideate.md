---
description: "Brainstorm feature ideas based on the current codebase scope. Pass a route/area path, or leave empty to use staged/recent files from context. Example: /ideate analytics/data-engine"
---

Brainstorm actionable feature ideas based on the current scope.

## Steps

### 1. Determine scope

**Mode A — explicit path** (when `$ARGUMENTS` is non-empty and does not start with `/`):

Treat `$ARGUMENTS` as a feature area (e.g. `analytics/data-engine`, `features/monitoring`). Read the route pages, components, and any related APIs/services to understand what currently exists.

**Mode B — context file list** (when `$ARGUMENTS` is empty):

Look in the conversation context for a `--- file list ---` block (injected by the `UserPromptSubmit` hook, or from a prior `/scope-by-staged` or `/scope-by-recent-changes` invocation). Use those files as the scope. If no file list is found, run:

```bash
bash .claude/scripts/scope-by-staged.sh
```

If still empty, fall back to recently changed files:

```bash
bash .claude/scripts/scope-by-recent-changes.sh 1w
```

**Mode C — follow-on from scope command** (when `$ARGUMENTS` starts with `/`):

This shouldn't normally happen — inform the user to use `/scope-by-staged /ideate` or `/scope-by-recent-changes 1w /ideate` instead.

### 2. Analyse the scope

Read the key files in scope (route pages, components, APIs, state files). Understand:
- What the feature currently does
- What UI elements and interactions exist
- What data is available (API endpoints, state stores, types)
- What patterns/components from the broader codebase could be reused

### 3. Generate ideas

Produce **5–8 concrete feature ideas** that would enhance the scoped area. For each idea:

- **Title** — short, descriptive name
- **What** — 1–2 sentence description of the feature
- **Why** — the user problem or opportunity it addresses
- **Reuses** — existing components, APIs, or patterns from the codebase it would build on
- **Effort** — rough T-shirt size (S / M / L) based on what already exists vs what's new

### 4. Prioritise

Rank the ideas by **impact / effort ratio** (best bang-for-buck first). Group them into:

- **Quick wins** (S effort, clear value)
- **High impact** (M–L effort, significant value)
- **Exploratory** (interesting but needs more research or design)

### 5. Present

Format the output as a clean numbered list with the groupings above. End with:

```
Pick a number to explore further, or ask me to draft a plan for any idea.
```
