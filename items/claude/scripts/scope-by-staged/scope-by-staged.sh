#!/usr/bin/env bash
# scope-by-staged.sh — list staged (index) files, grouped by area.
#
# Usage:
#   bash .claude/scripts/scope-by-staged.sh

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# Collect staged files (Added, Copied, Modified, Renamed)
STAGED=$(git diff --cached --diff-filter=ACMR --name-only | sort -u)

if [ -z "$STAGED" ]; then
  echo "No staged files."
  exit 0
fi

# --- Group and print using awk ---

echo "Staged files:"
echo ""

echo "$STAGED" | awk '
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
echo "$STAGED"
