#!/usr/bin/env bash
# scope-by-recent-changes.sh — list recently changed source files, grouped by area.
#
# Usage:
#   bash .claude/scripts/scope-by-recent-changes.sh [--dir=<path>] [TIME_RANGE]
#
# Options:
#   --dir=<path>   Subdirectory to scope the search (default: entire repo)
#
# TIME_RANGE examples:
#   24h | 1d           → 24 hours ago  (default)
#   48h | 2d           → 48 hours ago
#   1w  | 7d           → 1 week ago
#   2w  | 14d          → 2 weeks ago
#   2026-03-10         → since that date
#   2026-03-10T14:00   → since that datetime
#   since 2026-03-10   → same as above (leading "since" stripped)

set -euo pipefail

# --- Parse flags and collect remaining positional args ---

DIR_FILTER=""            # default: entire repo
REMAINING=()

for arg in "$@"; do
  case "$arg" in
    --dir=*)
      DIR_FILTER="${arg#--dir=}"
      # Strip trailing slash for consistency
      DIR_FILTER="${DIR_FILTER%/}"
      ;;
    *)
      REMAINING+=("$arg")
      ;;
  esac
done

# --- Parse time range from remaining args ---

RAW="${REMAINING[*]:-24h}"
# Strip optional leading "since " or "since:" prefix
RAW="${RAW#since }"
RAW="${RAW#since:}"

case "$RAW" in
  *h)
    NUM="${RAW%h}"
    SINCE="${NUM} hours ago"
    LABEL="${NUM} hour(s)"
    ;;
  *d)
    NUM="${RAW%d}"
    HOURS=$(( NUM * 24 ))
    SINCE="${HOURS} hours ago"
    LABEL="${NUM} day(s)"
    ;;
  *w)
    NUM="${RAW%w}"
    SINCE="${NUM} weeks ago"
    LABEL="${NUM} week(s)"
    ;;
  [0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]*)
    SINCE="$RAW"
    LABEL="$RAW"
    ;;
  *)
    echo "Error: unrecognised time range '${RAW}'" >&2
    echo "Supported: 24h, 48h, 1d, 7d, 1w, 2w, YYYY-MM-DD, YYYY-MM-DDTHH:MM" >&2
    exit 1
    ;;
esac

# --- Collect changed files ---

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

if [ -n "$DIR_FILTER" ]; then
  CHANGED=$(git log \
    --since="$SINCE" \
    --diff-filter=ACMR \
    --name-only \
    --format="" \
    -- "${DIR_FILTER}/**" \
    | grep -E "^${DIR_FILTER}/" \
    | sort -u)
  SCOPE_LABEL="${DIR_FILTER}/"
else
  CHANGED=$(git log \
    --since="$SINCE" \
    --diff-filter=ACMR \
    --name-only \
    --format="" \
    | sort -u)
  SCOPE_LABEL="entire repo"
fi

if [ -z "$CHANGED" ]; then
  echo "No source files changed in ${SCOPE_LABEL} in the last ${LABEL}."
  exit 0
fi

# --- Group and print using awk ---

echo "Recently changed files in ${SCOPE_LABEL} (since ${LABEL}):"
echo ""

echo "$CHANGED" | awk -v dir="$DIR_FILTER" '
function group_for(path,    m) {
  # SvelteKit routes: routes/(app)/<area>/<sub>/
  if (match(path, /\/routes\/\(app\)\/([^\/]+\/[^\/]+)\//, m)) return m[1]
  if (match(path, /\/routes\/\(app\)\/([^\/]+)\//, m))         return m[1]
  if (match(path, /\/routes\/([^\/]+)\//, m))                  return "routes/" m[1]
  # components/<name>/
  if (match(path, /\/components\/([^\/]+)\//, m))              return "components/" m[1]
  # src/<area>/
  if (match(path, /\/src\/([^\/]+)\//, m))                     return m[1]
  # top-level dir after dir_filter root
  n = split(path, parts, "/")
  if (n >= 2) return parts[2]
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

# --- Raw file list (machine-readable marker consumed by follow-on commands) ---

echo "--- file list ---"
echo "$CHANGED"
