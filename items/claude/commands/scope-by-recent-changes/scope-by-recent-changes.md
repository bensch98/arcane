---
description: "Show recently changed source files grouped by area, then optionally run a follow-on slash command scoped to them. Accepts a time range (24h, 1w, etc.) and/or a slash command. Example: /scope-by-recent-changes 1w /i18n"
---

Show recently changed source files and optionally run a follow-on command scoped to them.

## Steps

### 1. Parse `$ARGUMENTS`

Split `$ARGUMENTS` into two parts:
- **Time range** — any token that is NOT a slash command (default: `24h`)
- **Follow-on command** — any token starting with `/` (e.g. `/i18n`, `/walkthrough`)

Examples:
- `` (empty) → time range `24h`, no follow-on
- `1w` → time range `1w`, no follow-on
- `/i18n` → time range `24h`, follow-on `/i18n`
- `1w /i18n` → time range `1w`, follow-on `/i18n`
- `--dir=ui 2d /i18n` → flags `--dir=ui`, time range `2d`, follow-on `/i18n`

Accepted time range formats:
- `24h`, `48h`, `1d`, `2d` — hours/days
- `1w`, `2w`, `7d`, `14d` — weeks
- `2026-03-10` — since a specific date
- `2026-03-10T14:00` — since a specific datetime
- `since 2026-03-10` — with optional "since" prefix

### 2. Run the discovery script

Pass only the non-slash-command parts to the script:

```bash
bash .claude/scripts/scope-by-recent-changes.sh <time-range-and-flags>
```

### 3. Present the results

Display the grouped output from the script clearly. The output includes:
- A human-readable grouped section (files organised by feature/component area)
- A `--- file list ---` section with the raw file list

### 4. Check for a follow-on command

If a follow-on slash command was found in `$ARGUMENTS`, invoke that skill immediately — the `--- file list ---` block is already in context so the follow-on command can consume it via its "Mode B" / context-based file discovery.

If no follow-on command was specified, show follow-on options instead:

```
Follow-on options:
  /scope-by-recent-changes <range> /i18n          — i18n all files above
  /scope-by-recent-changes <range> /walkthrough   — create walkthroughs for routes above
  bun run test                                     — run static analysis checks
```

Do not automatically run any follow-on action unless one was specified in `$ARGUMENTS`.
