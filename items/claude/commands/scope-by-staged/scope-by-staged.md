---
description: "Show staged files grouped by area, then optionally run a follow-on slash command scoped to them. Example: /scope-by-staged /i18n"
---

Show all staged (git index) files and optionally run a follow-on command scoped to them.

## Steps

### 1. Run the discovery script

```bash
bash .claude/scripts/scope-by-staged.sh
```

If there are no staged files, inform the user and stop.

### 2. Present the results

Display the grouped output from the script clearly. The output includes:
- A human-readable grouped section (files organised by feature/component area)
- A `--- file list ---` section with the raw file list

### 3. Check for a follow-on command

If `$ARGUMENTS` contains a slash command (e.g. `/i18n`, `/walkthrough`), invoke that skill immediately — the `--- file list ---` block is already in context so the follow-on command can consume it via its "Mode B" / context-based file discovery.

If `$ARGUMENTS` is empty, show follow-on options instead:

```
Follow-on options:
  /scope-by-staged /i18n          — i18n all staged files
  /scope-by-staged /walkthrough   — create walkthroughs for staged routes
  bun run test                    — run static analysis checks
```

Do not automatically run any follow-on action unless one was specified in `$ARGUMENTS`.
