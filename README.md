# arcane

A shadcn-style registry for Claude Code commands, scripts, skills, and hooks. Browse, install, and manage reusable Claude Code extensions across your projects.

## Install

### Linux / macOS

One-liner (auto-detects OS and architecture):

```bash
curl -fsSL https://raw.githubusercontent.com/bensch98/arcane/main/install.sh | bash
```

Or download a specific binary from [GitHub Releases](https://github.com/bensch98/arcane/releases):

```bash
# Linux (amd64)
curl -fsSL https://github.com/bensch98/arcane/releases/latest/download/arcane-linux-amd64 -o arcane
chmod +x arcane && sudo mv arcane /usr/local/bin/

# macOS (Apple Silicon)
curl -fsSL https://github.com/bensch98/arcane/releases/latest/download/arcane-darwin-arm64 -o arcane
chmod +x arcane && sudo mv arcane /usr/local/bin/
```

### Windows (PowerShell)

One-liner (auto-detects architecture, installs to `%LOCALAPPDATA%\Programs\arcane` and adds it to user PATH):

```powershell
irm https://raw.githubusercontent.com/bensch98/arcane/main/install.ps1 | iex
```

Or download a specific binary manually:

```powershell
iwr https://github.com/bensch98/arcane/releases/latest/download/arcane-windows-amd64.exe -OutFile arcane.exe
```

Prerequisites: [Git for Windows](https://git-scm.com/download/win) must be installed and `git` on PATH — `arcane` shells out to `git` to fetch the registry.

### Go / source

Or via `go install`:

```bash
go install github.com/bensch98/arcane@latest
```

Or build from source:

```bash
git clone git@github.com:bensch98/arcane.git ~/repos/arcane
cd ~/repos/arcane && make && sudo cp arcane /usr/local/bin/
```

## Dev

```bash
# Check latest tag
git tag --sort=-v:refname | head -1

# Tag a new release and push to trigger the release workflow
git tag v0.0.2
git push origin v0.0.2
```

Pushing a `v*` tag builds cross-platform binaries and creates a GitHub Release automatically.

## Usage

```bash
# Browse available items
arcane list
arcane list --type=command
arcane list --type=script
arcane list i18n

# View item details
arcane info commit-message

# Install items (+ dependencies) into the current project
arcane init                                    # creates .arcane.json tracking file
arcane add command commit-message              # install a single command
arcane add command commit-message i18n ideate  # install multiple at once
arcane add hook stop-notify-toast              # install a hook
arcane add all                                 # install everything
arcane add sync                                # reinstall items from .arcane.json
arcane add command i18n --global               # install to ~/.claude instead
arcane add command walkthrough --dry-run       # preview without writing files

# Update registry and check for outdated items
arcane update

# Remove an installed item
arcane remove commit-message

# Upgrade the CLI itself
arcane version                                 # print current version
arcane upgrade                                 # download latest release from GitHub
```

## Available Items

### Commands

| Name | Description |
|------|-------------|
| `commit-message` | Generate a commit message for staged git changes |
| `i18n` | Integrate a feature/page into paraglide i18n |
| `ideate` | Brainstorm feature ideas based on current codebase scope |
| `load-arch` | Load UI architectural patterns and component conventions |
| `rm-dead-code` | Remove unused imports, functions, variables, and unreachable code |
| `check-bundle-impact` | Flag heavy/large package imports in frontend files |
| `walkthrough` | Create a new walkthrough for a route with data-wt attributes and i18n |
| `scope-by-staged` | Show staged files grouped by area + follow-on command |
| `scope-by-flagged-todos` | Show files with TODO/FIXME/HACK comments grouped by area |
| `scope-by-large-files` | Show files above a line threshold -- refactoring candidates |
| `scope-by-unstaged` | Show unstaged files grouped by area + follow-on command |
| `scope-by-recent-changes` | Show recently changed files grouped by area + follow-on command |

### Scripts

| Name | Description |
|------|-------------|
| `notify-toast-script` | Windows toast notification when Claude finishes |
| `notify-done-script` | Voice notification when Claude finishes (WSL/Linux) |
| `scope-by-staged-script` | List staged files grouped by area |
| `scope-by-flagged-todos-script` | List files with TODO/FIXME/HACK grouped by area |
| `scope-by-large-files-script` | List source files above a line threshold grouped by area |
| `scope-by-unstaged-script` | List unstaged files grouped by area |
| `scope-by-recent-changes-script` | List recently changed files grouped by area |
| `check-bundle-impact-script` | Find frontend files importing known heavy packages |

### Skills

| Name | Description |
|------|-------------|
| `generate-api` | Generate frontend API module from an OpenAPI JSON excerpt |

### Hooks

| Name | Description |
|------|-------------|
| `stop-notify-toast` | Show Windows toast when Claude stops responding |
| `post-edit-prettier` | Run prettier on files after Edit/Write tool use |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ARCANE_REGISTRY` | `~/repos/arcane` | Path to the registry repo |

If `ARCANE_REGISTRY` is unset and no local clone is found, the registry is auto-fetched into:

- Linux: `$XDG_DATA_HOME/arcane/registry` or `~/.local/share/arcane/registry`
- macOS: `~/Library/Application Support/arcane/registry`
- Windows: `%LOCALAPPDATA%\arcane\registry`

## How It Works

Items are defined in `registry.json` with their source files, target paths, dependencies, and install hooks. When you run `arcane add <type> <name...>`, the CLI:

1. Resolves the dependency tree (topological sort)
2. Copies source files to their target locations (`.claude/commands/`, `.claude/scripts/`, etc.)
3. Merges hook entries into `.claude/settings.json` (for hook-type items)
4. Tracks installed items in `.arcane.json` for updates and removal
