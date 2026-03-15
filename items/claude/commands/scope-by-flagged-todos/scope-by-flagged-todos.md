---
description: "Show files containing TODO/FIXME/HACK comments, grouped by area, then optionally run a follow-on slash command scoped to them. Use --dir=<path> to narrow scope. Example: /scope-by-flagged-todos /ideate"
---

Show files with TODO/FIXME/HACK comments and optionally run a follow-on command scoped to them.

## Steps

### 1. Parse `$ARGUMENTS`

Split `$ARGUMENTS` into:
- **Flags** — tokens like `--dir=ui/src` (passed to the script)
- **Follow-on command** — any token starting with `/` (e.g. `/i18n`, `/ideate`)

### 2. Run the discovery script

```bash
bash .claude/scripts/scope-by-flagged-todos.sh <flags>
```

If no files are found, inform the user and stop.

### 3. Present the results

Display the grouped output from the script clearly. The output includes:
- A human-readable grouped section (files organised by feature/component area)
- A `--- file list ---` section with the raw file list

### 4. Check for a follow-on command

If a follow-on slash command was found in `$ARGUMENTS`, invoke that skill immediately — the `--- file list ---` block is already in context so the follow-on command can consume it via its "Mode B" / context-based file discovery.

If no follow-on command was specified, show follow-on options instead:

```
Follow-on options:
  /scope-by-flagged-todos /ideate       — brainstorm improvements for flagged areas
  /scope-by-flagged-todos /i18n         — i18n flagged files
  /scope-by-flagged-todos /simplify     — review flagged files for quality
```

Do not automatically run any follow-on action unless one was specified in `$ARGUMENTS`.
