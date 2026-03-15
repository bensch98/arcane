---
description: "Show files above a line threshold (default 800), grouped by area — refactoring candidates. Use --min=<lines> and --dir=<path> to tune. Example: /scope-by-large-files --min=500 /simplify"
---

Show large source files (refactoring candidates) and optionally run a follow-on command scoped to them.

## Steps

### 1. Parse `$ARGUMENTS`

Split `$ARGUMENTS` into:
- **Flags** — `--min=<lines>` (default: 800), `--dir=<path>` (default: entire repo)
- **Follow-on command** — any token starting with `/` (e.g. `/simplify`, `/ideate`)

### 2. Run the discovery script

```bash
bash .claude/scripts/scope-by-large-files.sh <flags>
```

If no files are found, inform the user and stop.

### 3. Present the results

Display the grouped output from the script clearly. The output includes:
- A human-readable grouped section with line counts per file
- A `--- file list ---` section with the raw file list

### 4. Check for a follow-on command

If a follow-on slash command was found in `$ARGUMENTS`, invoke that skill immediately — the `--- file list ---` block is already in context so the follow-on command can consume it via its "Mode B" / context-based file discovery.

If no follow-on command was specified, show follow-on options instead:

```
Follow-on options:
  /scope-by-large-files /simplify     — review large files for quality & refactoring
  /scope-by-large-files /ideate       — brainstorm improvements for large files
  /scope-by-large-files --min=300     — raise threshold to 300 lines
  /scope-by-large-files --dir=ui/src  — narrow to UI source
```

Do not automatically run any follow-on action unless one was specified in `$ARGUMENTS`.
