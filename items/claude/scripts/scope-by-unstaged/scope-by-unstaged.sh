#!/usr/bin/env bash
# scope-by-unstaged.sh — list unstaged (working tree) files, grouped by area.
#
# Usage:
#   bash .claude/scripts/scope-by-unstaged.sh

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# Collect unstaged modified files + untracked files
MODIFIED=$(git diff --diff-filter=ACMR --name-only 2>/dev/null | sort -u)
UNTRACKED=$(git ls-files --others --exclude-standard 2>/dev/null | sort -u)

# Merge and deduplicate
UNSTAGED=$(printf '%s\n' "$MODIFIED" "$UNTRACKED" | grep -v '^$' | sort -u)

if [ -z "$UNSTAGED" ]; then
  echo "No unstaged files."
  exit 0
fi

# --- Group and print using awk ---

echo "Unstaged files:"
echo ""

echo "$UNSTAGED" | awk '
function group_for(path,    m) {
  if (match(path, /\/routes\/\(app\)\/([^\/]+\/[^\/]+)\//, m)) return m[1]
  if (match(path, /\/routes\/\(app\)\/([^\/]+)\//, m))         return m[1]
  if (match(path, /\/routes\/([^\/]+)\//, m))                  return "routes/" m[1]
  if (match(path, /\/components\/([^\/]+)\//, m))              return "components/" m[1]
  if (match(path, /\/src\/([^\/]+)\//, m))                     return m[1]
  n = split(path, parts, "/")
  if (n >= 2) return parts[1]
  return "other"
}
{
  g = group_for($0)
  if (!(g in seen)) { order[++n] = g; seen[g] = 1 }
  files[g] = files[g] $0 "\n"
  count[g]++
}
END {
  for (i = 1; i <= n; i++) {
    g = order[i]
    printf "[%s] — %d file(s)\n", g, count[g]
    printf "%s", files[g]
    print ""
  }
}
'

# --- Raw file list ---

echo "--- file list ---"
echo "$UNSTAGED"
