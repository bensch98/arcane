---
description: "Flag heavy/large package imports in frontend files — bundle size candidates. Use --dir=<path> to narrow scope. Example: /check-bundle-impact /simplify"
---

Find frontend files importing known heavy packages and optionally run a follow-on command scoped to them.

## Steps

### 1. Parse `$ARGUMENTS`

Split `$ARGUMENTS` into:
- **Flags** — `--dir=<path>` (default: `ui/src`)
- **Follow-on command** — any token starting with `/` (e.g. `/simplify`, `/ideate`)

### 2. Run the discovery script

```bash
bash .claude/scripts/check-bundle-impact.sh <flags>
```

If no heavy imports are found, inform the user and stop.

### 3. Present the results

Display the grouped output from the script clearly. The output includes:
- A human-readable grouped section with the specific import lines per file
- A `--- file list ---` section with the raw file list

### 4. Analyse and advise

For each flagged import, briefly note:
- Whether it's **justified** (core to the feature, e.g. echarts in a chart component, duckdb in the data engine)
- Whether it could be **lazy-loaded** (dynamic `import()` behind user interaction)
- Whether a **lighter alternative** exists (e.g. `dayjs` over `moment`, tree-shaken lodash-es)

### 5. Check for a follow-on command

If a follow-on slash command was found in `$ARGUMENTS`, invoke that skill immediately — the `--- file list ---` block is already in context so the follow-on command can consume it via its "Mode B" / context-based file discovery.

If no follow-on command was specified, show follow-on options instead:

```
Follow-on options:
  /check-bundle-impact /simplify     — review flagged files for optimisation
  /check-bundle-impact /ideate       — brainstorm lighter alternatives
  /check-bundle-impact --dir=ui/src/routes  — narrow to route files only
```

Do not automatically run any follow-on action unless one was specified in `$ARGUMENTS`.
