---
description: "Create a new walkthrough for a route. Scans the page for key UI elements, adds data-wt attributes, registers steps in walkthroughs.ts, and adds i18n messages."
---

Create a walkthrough for the route specified in `$ARGUMENTS`.

## Steps

### 1. Resolve the route

Parse `$ARGUMENTS` to determine the route path. Examples:
- `features/monitoring` → route `/features/monitoring`, files in `ui/src/routes/(app)/features/monitoring/`
- `use-cases/shift` → route `/use-cases/shift`, files in `ui/src/routes/(app)/use-cases/shift/`
- `/analytics/data-engine` → route `/analytics/data-engine` (leading slash is fine)

Strip any leading `/` for the file path lookup but keep it for the walkthrough registry key. If the route is ambiguous, search for matching route directories and ask the user to confirm.

### 2. Check for existing walkthrough

Read `ui/src/routes/(app)/walkthroughs.ts` and verify the route does not already have a walkthrough entry. If it does, inform the user and ask what they'd like to do (update existing or abort).

Also check `ui/tests/static-analysis/walkthrough-coverage.test.ts` — if the route is listed in `PAGES_WITHOUT_WALKTHROUGH`, you will need to remove it from that list at the end.

### 3. Identify target elements

Read the `+page.svelte` file for the route and any components it imports from `ui/src/components/`. Identify the key interactive/visual elements that a user should learn about in a walkthrough:

- **Buttons** — create, delete, save, toggle actions
- **Navigation elements** — tabs, breadcrumbs, sidebar sections
- **Content areas** — cards, tables, charts, panels, empty states
- **Form elements** — key inputs, selectors, configurators
- **Toolbar actions** — search, filter, view mode toggles

Aim for 4–8 steps per walkthrough. Prioritize elements that:
1. Are visible on initial page load (no dialog/modal content)
2. Help users understand the page's core workflow
3. Follow a natural left-to-right, top-to-bottom visual order

### 4. Choose a naming prefix

Derive a short prefix from the route for `data-wt` attributes and i18n keys:

| Route | Prefix |
|-------|--------|
| `/features/monitoring` | `monitoring` |
| `/use-cases/shift` | `shift` |
| `/analytics/data-engine` | `dataengine` |
| `/use-cases/ml-systems/[ml_system_uuid]` | `mlsystem` |

The prefix should be concise (one word if possible). Check existing walkthroughs in `walkthroughs.ts` for naming precedent.

### 5. Add `data-wt` attributes

For each identified element, add a `data-wt` attribute to the HTML tag:

```svelte
<Button data-wt="wt-{prefix}-{element}">Create</Button>
```

Naming convention: `wt-{prefix}-{element}` where `{element}` is a short lowercase descriptor (e.g. `create`, `table`, `tabs`, `card`, `empty`, `search`).

If the element is inside a reusable component that passes `...restProps`, add `data-wt` at the call site. If the component does NOT spread rest props, add it to the outermost DOM element inside the component instead, or wrap with a `<div>`.

### 6. Add i18n messages

For each step, add two keys to both `ui/messages/en.json` and `ui/messages/de.json`:

- `wt_{prefix}_{element}_title` — short title (2-4 words) describing the element
- `wt_{prefix}_{element}_description` — 1-2 sentence description explaining what the element does and how to use it

Place keys alphabetically near other `wt_` keys in the JSON files. Write meaningful, user-friendly descriptions — not technical implementation details.

For German translations: use formal "Sie" form, capitalize nouns, and ensure grammatical correctness.

### 7. Register the walkthrough

Add the new walkthrough entry to the `getWalkthroughs()` return object in `ui/src/routes/(app)/walkthroughs.ts`:

```typescript
'{route}': {
    steps: [
        {
            target: 'wt-{prefix}-{element}',
            title: m.wt_{prefix}_{element}_title(),
            description: m.wt_{prefix}_{element}_description(),
            position: '{position}'
        },
        // ... more steps
    ]
},
```

Position guidelines:
- `'bottom'` — for toolbar buttons and top-of-page elements (most common)
- `'right'` — for left-side panels, sidebars
- `'left'` — for right-side panels
- `'top'` — for bottom-of-page elements
- `'center'` — for large content areas

Use fallbacks when an element might not exist (e.g., a card vs empty state):
```typescript
{
    target: 'wt-{prefix}-card',
    title: m.wt_{prefix}_card_title(),
    description: m.wt_{prefix}_card_description(),
    position: 'bottom',
    fallback: {
        target: 'wt-{prefix}-empty',
        title: m.wt_{prefix}_empty_title(),
        description: m.wt_{prefix}_empty_description(),
        position: 'bottom'
    }
}
```

### 8. Update coverage test

If the route was listed in `PAGES_WITHOUT_WALKTHROUGH` in `ui/tests/static-analysis/walkthrough-coverage.test.ts`, remove it from the allowlist.

### 9. Verify

- Confirm all `m.wt_*()` calls in the new walkthrough entry have matching keys in `en.json` and `de.json`
- Confirm all `data-wt` targets referenced in the steps exist in the source files
- Run `cd ui && bun run fix:i18n` to clean up any unused keys

## Important rules

- **Step order** should follow the natural visual flow of the page (top-left to bottom-right)
- **Do not** add walkthrough steps for elements inside dialogs/modals (use `dialog:` prefix walkthroughs for those)
- **Reuse** existing `data-wt` attributes if they already exist on elements (check first)
- **Keep descriptions** user-friendly and actionable — explain *what* the user can do, not implementation details
- Match the exact formatting style of existing entries in `walkthroughs.ts` (tabs, single quotes, trailing commas)
