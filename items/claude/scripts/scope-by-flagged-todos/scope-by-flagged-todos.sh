#!/usr/bin/env bash
# scope-by-flagged-todos.sh — list files containing TODO/FIXME/HACK comments, grouped by area.
#
# Usage:
#   bash .claude/scripts/scope-by-flagged-todos.sh [--dir=<path>]

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# --- Parse flags ---
DIR_FILTER=""
for arg in "$@"; do
  case "$arg" in
    --dir=*) DIR_FILTER="${arg#--dir=}"; DIR_FILTER="${DIR_FILTER%/}" ;;
  esac
done

# --- Find files with TODO/FIXME/HACK ---
SEARCH_PATH="${DIR_FILTER:-.}"

FILES=$(grep -rlE '\b(TODO|FIXME|HACK)\b' "$SEARCH_PATH" \
  --include='*.svelte' --include='*.ts' --include='*.svelte.ts' \
  --include='*.js' --include='*.py' --include='*.sh' \
  | grep -v node_modules | grep -v '.svelte-kit' | grep -v dist \
  | sort -u) || true

if [ -z "$FILES" ]; then
  echo "No files with TODO/FIXME/HACK comments found in ${DIR_FILTER:-entire repo}."
  exit 0
fi

SCOPE_LABEL="${DIR_FILTER:-entire repo}"

echo "Files with TODO/FIXME/HACK in ${SCOPE_LABEL}:"
echo ""

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

echo "--- file list ---"
echo "$FILES"
