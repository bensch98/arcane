---
description: "Remove unused imports, functions, variables, and unreachable code. Shows a removal summary. Can be prefixed by a scoping command (e.g. /scope-by-staged /rm-dead-code)."
---

Find and remove dead code (unused imports, functions, variables, unreachable code, dead branches) from source files, then show a summary of what was removed.

## Steps

### 1. Determine file scope

**Mode A — `$ARGUMENTS` contains `--dir=<path>`:**
Use that directory as the search root. Default extensions: `.ts`, `.svelte`, `.svelte.ts`, `.py`.

**Mode B — A `--- file list ---` block exists in the current conversation context** (provided by a scoping command like `/scope-by-staged`, `/scope-by-recent-changes`, etc.):
Use exactly those files. Ignore `--dir` if both are present.

**Mode C — No scope provided:**
Default to `--dir=ui/src` plus `backend/` (`.py` files).

Only process source files (`.ts`, `.svelte`, `.svelte.ts`, `.js`, `.py`). Skip generated files, `node_modules`, `.svelte-kit`, `__pycache__`, `dist`, `build`, `static`.

### 2. Analyse each file for dead code

Read every in-scope file and look for:

1. **Unused imports** — imported names that are never referenced in the rest of the file. Be careful with:
   - Svelte components used in the template (markup) section
   - Type-only imports used in type annotations
   - Side-effect imports (`import './styles.css'`, `import 'module'`) — **keep these**
   - Re-exports (`export { X } from '...'`) — **keep these**
   - Imports used only in `$derived`, `$effect`, or reactive declarations
2. **Unused variables / constants** — declared but never read. Ignore:
   - Variables prefixed with `_` (intentionally unused)
   - Exported variables (they may be consumed elsewhere)
   - Variables used in template/markup sections of `.svelte` files
3. **Unused functions** — declared but never called or referenced. Ignore:
   - Exported functions
   - Lifecycle callbacks, event handlers referenced in markup
   - Functions passed as callbacks
4. **Unreachable code** — code after unconditional `return`, `throw`, `break`, `continue`
5. **Dead branches** — `if (false)`, `if (true) { ... } else { DEAD }`, etc. Only for literal boolean conditions.

**Important:** Be conservative. When in doubt, keep the code. Only remove things you are confident are unused *within the file* (for non-exported items) or provably unreachable.

For **exported** symbols: do NOT remove them based on single-file analysis. Only remove an exported symbol if you can confirm via a codebase-wide search (Grep) that it has zero external consumers.

### 3. Remove dead code

For each file with dead code found, use the Edit tool to remove it. Clean up:
- Trailing commas or empty lines left behind
- Empty import statements (e.g. `import {} from '...'` or `import '...'` that was only pulling in removed names)
- Blank lines where a block of dead code was removed (collapse to single blank line max)

### 4. Present a summary

After all edits, show a summary table:

```
## Dead code removal summary

| File | Removed | Details |
|------|---------|---------|
| src/components/foo/bar.svelte | 3 items | 2 unused imports, 1 unused variable |
| src/apis/baz/index.ts | 1 item  | 1 unused import |
| ...  | ...     | ...     |

**Total: N items removed across M files**
```

Categories to count: unused imports, unused variables, unused functions, unreachable code, dead branches.

If nothing was found, say so: "No dead code found in the scoped files."
