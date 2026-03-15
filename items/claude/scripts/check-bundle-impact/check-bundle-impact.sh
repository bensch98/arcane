#!/usr/bin/env bash
# check-bundle-impact.sh — find frontend files importing known heavy packages.
#
# Usage:
#   bash .claude/scripts/check-bundle-impact.sh [--dir=<path>]

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

SEARCH_PATH="${DIR_FILTER:-ui/src}"
SCOPE_LABEL="${DIR_FILTER:-ui/src}"

# Heavy packages pattern — add more as needed
# Format: package-name (reason / approximate size)
HEAVY_PATTERN='(lodash[^/]|moment[^-]|date-fns|chart\.js|d3[^a-z]|three[^a-z]|@firebase|rxjs|pdf-lib|xlsx|exceljs|pdfmake|mapbox-gl|leaflet|echarts(?!-)|echarts-gl|apache-arrow|@duckdb|mathjs|katex|highlight\.js|prismjs|monaco-editor|codemirror|@codemirror)'

# Find files with heavy imports
RESULTS=$(grep -rnE "^\s*(import|from)\s.*${HEAVY_PATTERN}" "$SEARCH_PATH" \
  --include='*.svelte' --include='*.ts' --include='*.svelte.ts' --include='*.js' \
  | grep -v node_modules | grep -v '.svelte-kit' | grep -v dist) || true

if [ -z "$RESULTS" ]; then
  echo "No heavy imports detected in ${SCOPE_LABEL}."
  exit 0
fi

echo "Heavy imports detected in ${SCOPE_LABEL}:"
echo ""

# Extract unique files
FILES=$(echo "$RESULTS" | cut -d: -f1 | sort -u)

# Group files by area, with import details
echo "$RESULTS" | awk -F: '
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
  filepath = $1
  lineno = $2
  # Rejoin remaining fields (the import line itself)
  line = ""
  for (i = 3; i <= NF; i++) line = line (i > 3 ? ":" : "") $i
  gsub(/^[[:space:]]+/, "", line)

  g = group_for(filepath)
  if (!(g in seen_group)) { order[++n] = g; seen_group[g] = 1 }

  key = filepath
  if (!(key in seen_file)) {
    seen_file[key] = 1
    file_order[g] = file_order[g] key "\n"
    file_count[g]++
  }
  details[key] = details[key] sprintf("    L%s: %s\n", lineno, line)
}
END {
  for (i = 1; i <= n; i++) {
    g = order[i]
    printf "[%s] — %d file(s)\n", g, file_count[g]
    split(file_order[g], flist, "\n")
    for (j = 1; j <= length(flist); j++) {
      f = flist[j]
      if (f == "") continue
      printf "  %s\n", f
      printf "%s", details[f]
    }
    print ""
  }
}
'

echo "--- file list ---"
echo "$FILES"
