#!/usr/bin/env bash
# scope-by-large-files.sh — list source files above a line threshold, grouped by area.
#
# Usage:
#   bash .claude/scripts/scope-by-large-files.sh [--dir=<path>] [--min=<lines>]
#
# Defaults: --min=800, searches entire repo

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# --- Parse flags ---
DIR_FILTER=""
MIN_LINES=800

for arg in "$@"; do
  case "$arg" in
    --dir=*)  DIR_FILTER="${arg#--dir=}"; DIR_FILTER="${DIR_FILTER%/}" ;;
    --min=*)  MIN_LINES="${arg#--min=}" ;;
  esac
done

SEARCH_PATH="${DIR_FILTER:-.}"
SCOPE_LABEL="${DIR_FILTER:-entire repo}"

# --- Find large files ---
FILES=$(find "$SEARCH_PATH" \
  \( -name '*.svelte' -o -name '*.ts' -o -name '*.svelte.ts' \
     -o -name '*.js' -o -name '*.py' -o -name '*.sh' \) \
  ! -path '*/node_modules/*' ! -path '*/.svelte-kit/*' ! -path '*/dist/*' \
  -exec awk -v min="$MIN_LINES" 'END { if (NR >= min) print NR "\t" FILENAME }' {} \; \
  | sort -rn) || true

if [ -z "$FILES" ]; then
  echo "No files with >=${MIN_LINES} lines found in ${SCOPE_LABEL}."
  exit 0
fi

echo "Files with >=${MIN_LINES} lines in ${SCOPE_LABEL} (sorted by size):"
echo ""

# --- Group and print ---
echo "$FILES" | awk '
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
  lines = $1
  filepath = $2
  g = group_for(filepath)
  if (!(g in seen)) { order[++n] = g; seen[g] = 1 }
  files[g] = files[g] sprintf("  %s (%d lines)\n", filepath, lines)
  plain[g] = plain[g] filepath "\n"
  count[g]++
}
END {
  for (i = 1; i <= n; i++) {
    g = order[i]
    printf "[%s] — %d file(s)\n", g, count[g]
    printf "%s", files[g]
    print ""
  }
  print "--- file list ---"
  for (i = 1; i <= n; i++) {
    printf "%s", plain[order[i]]
  }
}
'
