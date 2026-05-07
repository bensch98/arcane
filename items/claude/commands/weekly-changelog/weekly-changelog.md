# Weekly Customer Changelog

Generate a customer-facing, feature-oriented changelog from the last week of git commits.

**Arguments:** `$ARGUMENTS`

Parse the arguments:

- Optional `--since=<date-or-range>` to override the default 1-week lookback (e.g. `--since="2 weeks ago"`, `--since=2026-04-28`).
- Optional `--lang=de` to output the changelog in German instead of English.

---

## Your task

### Step 1 — Collect commits

Run:

```
git log --since="1 week ago" --oneline --no-merges
```

If `--since` was provided, substitute it. This gives you the raw commit list.

### Step 2 — Group by product area

Map each commit to a customer-visible product area based on its message and changed paths. Typical areas for this project:

- OCR / Document Extraction
- Shift Monitoring
- Computer Vision / Annotation
- Process Control & Quality
- Data Views / Tables
- Notifications
- Data Engine / Analytics
- Namespace Management
- General UI & Performance
- Infrastructure / Reliability (only include if customer-visible, e.g. stability improvements)

Ignore purely internal commits: refactors, chores, style tweaks, docs, test-only changes, linting — unless they produced a visible user-facing effect (e.g. a redesign, a new font, a rebranded theme).

### Step 3 — Write the changelog

Rules:

- Audience: **non-technical customer stakeholders**. No code terms, file names, or implementation details.
- Each bullet describes _what the user can now do or see_, not _what changed in the code_.
- Use plain, confident language. Present tense. Active voice.
- Group bullets under bold area headings (e.g. **OCR / Document Extraction**).
- Omit areas with no user-visible changes.
- Add a short date range header at the top, e.g. `## Product Update — Week of April 30 – May 7, 2026`.
- End with a one-line note offering to adjust tone, add detail, or translate (unless `--lang=de` was given, in which case offer English instead).

If `--lang=de` was passed, write the entire changelog in German.

### Step 4 — Output

Print the changelog as formatted markdown. Do not save it to a file unless the user asks.
