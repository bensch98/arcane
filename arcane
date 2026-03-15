#!/usr/bin/env bash
# arcane — Agentic Registry CLI
# A shadcn-style registry for Claude Code commands, scripts, skills, and hooks.
#
# Usage:
#   arcane list [--tool=X] [--type=X] [SEARCH]
#   arcane info <name>
#   arcane add <name> [--global] [--force] [--dry-run]
#   arcane remove <name>
#   arcane init
#   arcane update

set -euo pipefail

# --- Configuration ---
ARCANE_REGISTRY="${ARCANE_REGISTRY:-$HOME/repos/arcane}"
REGISTRY_FILE="$ARCANE_REGISTRY/registry.json"
TRACKING_FILE=".arcane.json"

# --- Colors ---
BOLD='\033[1m'
DIM='\033[2m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'

# --- Helpers ---

die() { echo -e "${RED}error:${RESET} $*" >&2; exit 1; }

check_deps() {
  command -v jq &>/dev/null || die "jq is required but not installed. Install it with: sudo apt install jq"
}

check_registry() {
  [[ -f "$REGISTRY_FILE" ]] || die "Registry not found at $REGISTRY_FILE\nSet ARCANE_REGISTRY env var to your registry path."
}

get_item() {
  local name="$1"
  jq -e --arg n "$name" '.items[] | select(.name == $n)' "$REGISTRY_FILE" 2>/dev/null
}

get_target_root() {
  local global="${1:-false}"
  if [[ "$global" == "true" ]]; then
    echo "$HOME"
  else
    echo "."
  fi
}

# --- Dependency resolution (topological sort) ---

resolve_deps() {
  local name="$1"
  local -A visited=()
  local result=()

  _visit() {
    local n="$1"
    [[ -n "${visited[$n]:-}" ]] && return
    visited[$n]=1

    local deps
    deps=$(jq -r --arg n "$n" '.items[] | select(.name == $n) | .dependencies // [] | .[]' "$REGISTRY_FILE" 2>/dev/null)

    for dep in $deps; do
      # Verify dependency exists
      if ! jq -e --arg n "$dep" '.items[] | select(.name == $n)' "$REGISTRY_FILE" &>/dev/null; then
        die "Dependency '$dep' (required by '$n') not found in registry."
      fi
      _visit "$dep"
    done

    result+=("$n")
  }

  _visit "$name"
  echo "${result[@]}"
}

# --- Commands ---

cmd_list() {
  local tool_filter="" type_filter="" search=""

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --tool=*) tool_filter="${1#--tool=}" ;;
      --type=*) type_filter="${1#--type=}" ;;
      *) search="$1" ;;
    esac
    shift
  done

  local filter='.'
  [[ -n "$tool_filter" ]] && filter="$filter | select(.tool == \"$tool_filter\")"
  [[ -n "$type_filter" ]] && filter="$filter | select(.type == \"$type_filter\")"
  [[ -n "$search" ]] && filter="$filter | select((.name | test(\"$search\"; \"i\")) or (.description | test(\"$search\"; \"i\")) or (.tags | join(\" \") | test(\"$search\"; \"i\")))"

  local items
  items=$(jq -r ".items[] | $filter | [.type, .name, .description] | @tsv" "$REGISTRY_FILE" 2>/dev/null)

  if [[ -z "$items" ]]; then
    echo "No items found."
    return
  fi

  # Group by type
  local current_type=""
  while IFS=$'\t' read -r type name desc; do
    if [[ "$type" != "$current_type" ]]; then
      [[ -n "$current_type" ]] && echo ""
      echo -e "${BOLD}${type}s${RESET}"
      current_type="$type"
    fi
    printf "  ${CYAN}%-30s${RESET} %s\n" "$name" "$desc"
  done <<< "$(echo "$items" | sort -t$'\t' -k1,1 -k2,2)"
}

cmd_info() {
  local name="${1:-}"
  [[ -z "$name" ]] && die "Usage: arcane info <name>"

  local item
  item=$(get_item "$name") || die "Item '$name' not found."

  local tool type desc tags deps files
  tool=$(echo "$item" | jq -r '.tool')
  type=$(echo "$item" | jq -r '.type')
  desc=$(echo "$item" | jq -r '.description')
  tags=$(echo "$item" | jq -r '.tags // [] | join(", ")')
  deps=$(echo "$item" | jq -r '.dependencies // [] | join(", ")')
  files=$(echo "$item" | jq -r '.files[] | "  \(.src) → \(.target)"')

  echo -e "${BOLD}$name${RESET}"
  echo -e "${DIM}$desc${RESET}"
  echo ""
  echo -e "  Tool:         $tool"
  echo -e "  Type:         $type"
  [[ -n "$tags" ]] && echo -e "  Tags:         $tags"
  [[ -n "$deps" ]] && echo -e "  Dependencies: $deps"
  echo -e "  Files:"
  echo "$files"

  # Show file preview for single-file items
  local file_count
  file_count=$(echo "$item" | jq '.files | length')
  if [[ "$file_count" -eq 1 ]]; then
    local src
    src=$(echo "$item" | jq -r '.files[0].src')
    local full_path="$ARCANE_REGISTRY/$src"
    if [[ -f "$full_path" ]]; then
      echo ""
      echo -e "${DIM}--- preview (first 20 lines) ---${RESET}"
      head -20 "$full_path"
      local total_lines
      total_lines=$(wc -l < "$full_path")
      if [[ "$total_lines" -gt 20 ]]; then
        echo -e "${DIM}... ($total_lines lines total)${RESET}"
      fi
    fi
  fi

  # Show hook merge info for hook type
  if [[ "$type" == "hook" ]]; then
    local hook_event hook_matcher
    hook_event=$(echo "$item" | jq -r '.hookMerge.event // empty')
    hook_matcher=$(echo "$item" | jq -r '.hookMerge.entry.matcher // empty')
    if [[ -n "$hook_event" ]]; then
      echo ""
      echo -e "  Hook event:   $hook_event"
      [[ -n "$hook_matcher" ]] && echo -e "  Matcher:      $hook_matcher"
    fi
  fi
}

cmd_add() {
  local name="" global=false force=false dry_run=false

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --global) global=true ;;
      --force) force=true ;;
      --dry-run) dry_run=true ;;
      -*) die "Unknown flag: $1" ;;
      *) name="$1" ;;
    esac
    shift
  done

  [[ -z "$name" ]] && die "Usage: arcane add <name> [--global] [--force] [--dry-run]"

  # Verify item exists
  get_item "$name" &>/dev/null || die "Item '$name' not found in registry."

  # Resolve dependencies
  local items_to_install
  items_to_install=$(resolve_deps "$name")

  local target_root
  target_root=$(get_target_root "$global")

  echo -e "${BOLD}Installing:${RESET} $items_to_install"
  [[ "$dry_run" == "true" ]] && echo -e "${YELLOW}(dry run — no files will be written)${RESET}"
  echo ""

  local installed_files=()

  for item_name in $items_to_install; do
    local item
    item=$(get_item "$item_name")
    local item_type
    item_type=$(echo "$item" | jq -r '.type')

    echo -e "  ${CYAN}$item_name${RESET} ($item_type)"

    if [[ "$item_type" == "hook" ]]; then
      # Handle hook merge
      _install_hook "$item" "$target_root" "$force" "$dry_run"
    else
      # Handle file-based items
      local files_json
      files_json=$(echo "$item" | jq -c '.files[]')

      while IFS= read -r file_entry; do
        local src target
        src=$(echo "$file_entry" | jq -r '.src')
        target=$(echo "$file_entry" | jq -r '.target')

        local src_path="$ARCANE_REGISTRY/$src"
        local target_path="$target_root/$target"

        if [[ ! -f "$src_path" ]]; then
          echo -e "    ${RED}source not found:${RESET} $src_path"
          continue
        fi

        if [[ -f "$target_path" && "$force" != "true" ]]; then
          echo -e "    ${YELLOW}exists:${RESET} $target (use --force to overwrite)"
          continue
        fi

        if [[ "$dry_run" == "true" ]]; then
          echo -e "    ${DIM}would copy:${RESET} $target"
        else
          mkdir -p "$(dirname "$target_path")"
          cp "$src_path" "$target_path"
          echo -e "    ${GREEN}copied:${RESET} $target"
          installed_files+=("$target")
        fi
      done <<< "$files_json"

      # Run postInstall
      local post_install
      post_install=$(echo "$item" | jq -r '.postInstall // empty')
      if [[ -n "$post_install" && "$dry_run" != "true" ]]; then
        # Handle postInstall commands
        case "$post_install" in
          "chmod +x")
            while IFS= read -r file_entry; do
              local target
              target=$(echo "$file_entry" | jq -r '.target')
              local target_path="$target_root/$target"
              if [[ -f "$target_path" ]]; then
                chmod +x "$target_path"
                echo -e "    ${DIM}chmod +x $target${RESET}"
              fi
            done <<< "$files_json"
            ;;
          *)
            echo -e "    ${YELLOW}postInstall:${RESET} $post_install (manual step)"
            ;;
        esac
      fi
    fi
  done

  # Update tracking file
  if [[ "$dry_run" != "true" && "$global" != "true" && -f "$TRACKING_FILE" ]]; then
    _track_install "$name" "${installed_files[@]}"
  fi

  echo ""
  echo -e "${GREEN}Done.${RESET}"
}

_install_hook() {
  local item="$1" target_root="$2" force="$3" dry_run="$4"

  local settings_file="$target_root/.claude/settings.json"
  local event entry

  event=$(echo "$item" | jq -r '.hookMerge.event')
  entry=$(echo "$item" | jq -c '.hookMerge.entry')

  if [[ "$dry_run" == "true" ]]; then
    echo -e "    ${DIM}would merge hook into:${RESET} $settings_file (event: $event)"
    return
  fi

  # Ensure settings file exists
  mkdir -p "$(dirname "$settings_file")"
  if [[ ! -f "$settings_file" ]]; then
    echo '{}' > "$settings_file"
  fi

  # Read current settings
  local current
  current=$(cat "$settings_file")

  # Merge: ensure hooks.<event> is an array, then append entry if not already present
  local updated
  updated=$(echo "$current" | jq --arg event "$event" --argjson entry "$entry" '
    .hooks //= {} |
    .hooks[$event] //= [] |
    if (.hooks[$event] | map(select(.matcher == $entry.matcher and .hooks == $entry.hooks)) | length) > 0 then
      if '"$force"' then
        .hooks[$event] = [.hooks[$event][] | select(.matcher != $entry.matcher or .hooks != $entry.hooks)] + [$entry]
      else
        .
      end
    else
      .hooks[$event] += [$entry]
    end
  ')

  echo "$updated" | jq '.' > "$settings_file"
  echo -e "    ${GREEN}merged hook:${RESET} $event → $settings_file"
}

_track_install() {
  local name="$1"
  shift
  local files=("$@")

  local sha=""
  if git -C "$ARCANE_REGISTRY" rev-parse HEAD &>/dev/null 2>&1; then
    sha=$(git -C "$ARCANE_REGISTRY" rev-parse --short HEAD 2>/dev/null || echo "unknown")
  fi

  local files_json
  files_json=$(printf '%s\n' "${files[@]}" | jq -R . | jq -s .)

  local current
  current=$(cat "$TRACKING_FILE")

  local updated
  updated=$(echo "$current" | jq --arg name "$name" --arg version "$sha" --argjson files "$files_json" '
    .installed = [(.installed // [])[] | select(.name != $name)] + [{name: $name, version: $version, files: $files}]
  ')

  echo "$updated" | jq '.' > "$TRACKING_FILE"
}

cmd_remove() {
  local name="${1:-}"
  [[ -z "$name" ]] && die "Usage: arcane remove <name>"

  # Check tracking file
  if [[ ! -f "$TRACKING_FILE" ]]; then
    die "No $TRACKING_FILE found. Run 'arcane init' first or remove files manually."
  fi

  local tracked
  tracked=$(jq -e --arg n "$name" '.installed[] | select(.name == $n)' "$TRACKING_FILE" 2>/dev/null) || die "Item '$name' is not tracked in $TRACKING_FILE."

  local files
  files=$(echo "$tracked" | jq -r '.files[]')

  echo -e "${BOLD}Removing:${RESET} $name"

  for f in $files; do
    if [[ -f "$f" ]]; then
      rm "$f"
      echo -e "  ${RED}deleted:${RESET} $f"
    else
      echo -e "  ${DIM}not found:${RESET} $f"
    fi
  done

  # Also handle hook removal
  local item
  item=$(get_item "$name" 2>/dev/null || echo "")
  if [[ -n "$item" ]]; then
    local item_type
    item_type=$(echo "$item" | jq -r '.type')
    if [[ "$item_type" == "hook" ]]; then
      local settings_file=".claude/settings.json"
      if [[ -f "$settings_file" ]]; then
        local event entry
        event=$(echo "$item" | jq -r '.hookMerge.event')
        entry=$(echo "$item" | jq -c '.hookMerge.entry')

        local updated
        updated=$(jq --arg event "$event" --argjson entry "$entry" '
          if .hooks[$event] then
            .hooks[$event] = [.hooks[$event][] | select(.matcher != $entry.matcher or .hooks != $entry.hooks)]
          else . end
        ' "$settings_file")

        echo "$updated" | jq '.' > "$settings_file"
        echo -e "  ${RED}removed hook:${RESET} $event from $settings_file"
      fi
    fi
  fi

  # Remove from tracking
  local updated_tracking
  updated_tracking=$(jq --arg n "$name" '.installed = [.installed[] | select(.name != $n)]' "$TRACKING_FILE")
  echo "$updated_tracking" | jq '.' > "$TRACKING_FILE"

  echo -e "${GREEN}Done.${RESET}"
}

cmd_init() {
  if [[ -f "$TRACKING_FILE" ]]; then
    echo -e "${YELLOW}$TRACKING_FILE already exists.${RESET}"
    return
  fi

  echo '{ "installed": [] }' | jq '.' > "$TRACKING_FILE"
  echo -e "${GREEN}Created $TRACKING_FILE${RESET}"
}

cmd_update() {
  # Pull latest registry
  if git -C "$ARCANE_REGISTRY" rev-parse --is-inside-work-tree &>/dev/null 2>&1; then
    echo -e "${BOLD}Pulling latest registry...${RESET}"
    git -C "$ARCANE_REGISTRY" pull --quiet 2>/dev/null || echo -e "${YELLOW}Could not pull (not a git remote or offline).${RESET}"
  fi

  # Check for outdated items
  if [[ ! -f "$TRACKING_FILE" ]]; then
    echo "No $TRACKING_FILE found. Run 'arcane init' first."
    return
  fi

  local current_sha=""
  if git -C "$ARCANE_REGISTRY" rev-parse HEAD &>/dev/null 2>&1; then
    current_sha=$(git -C "$ARCANE_REGISTRY" rev-parse --short HEAD 2>/dev/null || echo "")
  fi

  echo ""
  echo -e "${BOLD}Installed items:${RESET}"

  local items
  items=$(jq -r '.installed[] | [.name, .version] | @tsv' "$TRACKING_FILE" 2>/dev/null)

  if [[ -z "$items" ]]; then
    echo "  No items installed."
    return
  fi

  while IFS=$'\t' read -r name version; do
    if [[ -n "$current_sha" && "$version" != "$current_sha" ]]; then
      echo -e "  ${YELLOW}$name${RESET} (installed: $version, latest: $current_sha) ${YELLOW}← outdated${RESET}"
    else
      echo -e "  ${GREEN}$name${RESET} ($version)"
    fi
  done <<< "$items"

  if [[ -n "$current_sha" ]]; then
    echo ""
    echo -e "Registry at: ${DIM}$current_sha${RESET}"
    echo -e "Run ${CYAN}arcane add <name> --force${RESET} to update an item."
  fi
}

# --- Main ---

check_deps
check_registry

cmd="${1:-}"
shift 2>/dev/null || true

case "$cmd" in
  list)   cmd_list "$@" ;;
  info)   cmd_info "$@" ;;
  add)    cmd_add "$@" ;;
  remove) cmd_remove "$@" ;;
  init)   cmd_init "$@" ;;
  update) cmd_update "$@" ;;
  ""|help|-h|--help)
    echo -e "${BOLD}arcane${RESET} — Agentic Registry CLI"
    echo ""
    echo "Usage:"
    echo "  arcane list [--tool=X] [--type=X] [SEARCH]   Browse items"
    echo "  arcane info <name>                            Show item details"
    echo "  arcane add <name> [--global] [--force]        Install item + deps"
    echo "  arcane remove <name>                          Remove installed item"
    echo "  arcane init                                   Create .arcane.json tracking"
    echo "  arcane update                                 Pull registry + check outdated"
    echo ""
    echo "Environment:"
    echo "  ARCANE_REGISTRY   Path to registry repo (default: ~/repos/arcane)"
    ;;
  *)
    die "Unknown command: $cmd\nRun 'arcane help' for usage."
    ;;
esac
