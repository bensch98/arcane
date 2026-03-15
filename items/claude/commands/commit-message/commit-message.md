---
description: "Generate a commit message for staged git changes. Use when the user asks to write/create/draft a commit message."
---

Generate a precise and concise commit message for the currently staged changes in git.

## Steps

1. Run `git diff --cached` to see all staged changes.
2. Run `git log --oneline -5` to see recent commit style.
3. Run `git branch --show-current` to get the current branch name.
4. If there are no staged changes, inform the user and stop.
5. Analyze the staged diff and draft a commit message following the conventions below.
6. Present the commit message to the user and show which branch it would be committed to (e.g. "This would commit to branch: `main`").
7. Ask the user if they want to proceed with the commit.
8. If the user confirms, first run `git fetch --all && git pull` to sync with remote before committing.
9. Then create the commit on the current branch. Do NOT push. Do NOT add a Co-Authored-By line.

## Commit message conventions

- Use conventional commit prefixes: `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `test:`, `style:`, `perf:`
- For tiny/trivial changes: a single one-liner is fine.
- For normal/comprehensive changes: a one-liner summary, then an empty line, then bullet points with key details. Each bullet point must be a single line — no line breaks within a bullet point.
- Keep the summary line under 72 characters.
- Focus on the **why** and **what changed**, not low-level implementation details.
- Match the style of recent commits in the repo.
- Do NOT append a Co-Authored-By line to the commit message.
