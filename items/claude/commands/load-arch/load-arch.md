---
description: "Load UI architectural patterns and component conventions into context. Use at the start of a task to avoid re-explaining how components, routes, state, and APIs are structured."
---

Read and internalize the architectural patterns below. These are the conventions used throughout the Tensoryze UI codebase. Apply them when building new components, routes, or features — do not deviate unless explicitly asked.

---

## 1. Compound Component Pattern (Root > Toolbar > Content)

Complex UI is built as compound components with barrel exports. Each part is a separate `.svelte` file, composed together.

**File structure:**
```
ui/src/components/[name]/
├── [name]-root.svelte      # outer container, sets up context
├── [name]-toolbar.svelte    # actions bar
├── [name]-content.svelte    # main body
├── [name]-footer.svelte     # optional
├── ctx.svelte.ts            # context state class (if needed)
├── types.ts                 # shared types
└── index.ts                 # barrel exports
```

**Barrel export pattern (`index.ts`):**
```ts
export { default as Root } from './[name]-root.svelte';
export { default as Toolbar } from './[name]-toolbar.svelte';
export { default as Content } from './[name]-content.svelte';
// Also export long-form aliases
export { default as PageRoot } from './[name]-root.svelte';
```

**Usage:**
```svelte
<Page.Root>
  <Page.Toolbar>...</Page.Toolbar>
  <Page.Content>...</Page.Content>
</Page.Root>
```

**Reference implementations:** `ui/src/components/page/`, `ui/src/components/data-view/`, `ui/src/components/dashboard/`

---

## 2. Component Props & Extensibility

Every component follows this prop pattern:

```svelte
<script lang="ts">
  import type { HTMLAttributes } from 'svelte/elements';
  import { cn, type WithElementRef } from '$lib/utils.js';
  import type { Snippet } from 'svelte';

  let {
    ref = $bindable(null),
    class: className,
    children,
    ...restProps
  }: WithElementRef<HTMLAttributes<HTMLDivElement>> = $props();
</script>

<div bind:this={ref} class={cn("base-classes", className)} {...restProps}>
  {@render children?.()}
</div>
```

**Rules:**
- `$bindable(null)` for element refs
- `class: className` (destructure rename) + `cn()` for composable styling
- `...restProps` spread for passthrough
- `children` as `Snippet` — **never use slots**, always use snippets
- `{@render children?.()}` for rendering

---

## 3. TypeScript Generics on Components

Use the `generics` attribute on `<script>` for type-safe reusable components:

```svelte
<script lang="ts" generics="T extends Record<string, unknown>">
  let { items, schema }: { items: T[]; schema: Schema } = $props();
</script>
```

**For forms with path types:**
```svelte
<script lang="ts" generics="T extends Record<string, unknown>, U extends FormPath<T>">
```

---

## 4. Context API (Symbol Keys)

State shared across a compound component tree uses `setContext`/`getContext` with Symbol keys in a `ctx.svelte.ts` file:

```ts
// ctx.svelte.ts
const KEY = Symbol('feature-name');

export function setFeatureContext(state: FeatureState) {
  setContext(KEY, state);
}
export function getFeatureContext(): FeatureState {
  return getContext<FeatureState>(KEY);
}
```

**State class pattern:**
```ts
export class FeatureState {
  items = $state<Item[]>([]);
  activeId = $state<string | null>(null);
  filtered = $derived(this.items.filter(...));
}
```

The Root component creates and sets context; children consume it via `getFeatureContext()`.

**Reference:** `ui/src/components/data-view/ctx.svelte.ts`

---

## 5. State Management ($state runes)

### Global singletons (`ui/src/state/*.svelte.ts`)

```ts
class SomeService {
  static #instance: SomeService;
  #data = $state<Item[]>([]);

  static getInstance(): SomeService {
    if (!SomeService.#instance) SomeService.#instance = new SomeService();
    return SomeService.#instance;
  }

  get data() { return this.#data; }
}

export const someService = SomeService.getInstance();
```

### Rules
- **CRITICAL:** Files using runes outside `.svelte` must be `.svelte.ts`, imported as `./file.svelte.js`
- `$state` for mutable reactive values
- `$derived` for computed (expression) / `$derived.by()` for computed (function body)
- `$effect` for side effects (sync bindable props, watchers)
- `$bindable()` for two-way binding props between parent and child

---

## 6. Route Structure & Protection

```
ui/src/routes/
├── +layout.svelte          # root: ModeWatcher, Toaster, global CSS
├── auth/                   # public auth pages
└── (app)/                  # protected app shell
    ├── +layout.svelte      # sidebar, breadcrumbs, notifications, walkthroughs
    ├── +layout.ts           # load: auth token, user data
    ├── analytics/
    ├── features/
    ├── infrastructure/
    └── settings/
```

### Page load pattern (`+layout.ts` / `+page.ts`)
```ts
import type { LayoutLoad } from './$types';
export const load: LayoutLoad = async ({ fetch, params, parent }) => {
  const { tzuToken } = await parent();
  const data = await getData(fetch, tzuToken);
  return { data };
};
```

### Page component (`+page.svelte`)
```svelte
<script lang="ts">
  import type { PageData } from './$types';
  let { data }: { data: PageData } = $props();
</script>
```

**Key points:**
- `await parent()` chains auth token from `(app)/+layout.ts`
- Pass SvelteKit's `fetch` to API calls for SSR
- Route groups `(app)` for layout grouping, not URL segments

---

## 7. API Layer (`ui/src/apis/`)

```
ui/src/apis/
├── api.ts          # generic fetch wrappers (apiGet, apiPost, apiPut, apiDelete)
├── env.ts          # BASE_URL, httpUrl
├── auth-token.ts   # getAccessToken()
├── utils.ts        # logRequest, query param helpers
└── [domain]/
    ├── index.ts    # exported API functions
    └── types.ts    # TypeScript interfaces
```

### Standard API function
```ts
export const getItems = async (
  fetchFn: typeof fetch = fetch,
  token: string
): Promise<Item[]> => {
  if (!token) token = getAccessToken();

  const url = new URL(`${BASE_URL}/items`);
  const method = 'GET';
  const ctx: RequestLogContext = { method, url };

  const res = await fetchFn(url.toString(), {
    method,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`
    }
  }).then(async (res) => {
    logRequest(ctx);
    if (!res.ok) throw await res.json();
    return res.json();
  }).catch((err) => {
    logRequestError(ctx, err);
    throw err.detail ?? err;
  });

  return res as Item[];
};
```

**Rules:**
- First param is always `fetchFn` (SSR-compatible)
- Second param is `token` with `getAccessToken()` fallback
- Request logging via `logRequest`/`logRequestError`
- Bearer token in Authorization header

---

## 8. Shadcn-svelte Primitives (`ui/src/lib/components/ui/`)

- ~52 component directories (button, card, dialog, form, table, tabs, etc.)
- Built on **bits-ui** headless primitives
- Each uses `cn()`, `$bindable(null)` ref, `...restProps`, `data-slot` attributes
- Import from `$lib/components/ui/[name]`
- **Do not modify** these directly — wrap or compose on top of them in `ui/src/components/`

---

## 9. Forms (superforms + zod)

```svelte
<script lang="ts">
  import { z } from 'zod';
  import { defaults, superForm } from 'sveltekit-superforms';
  import { zod } from 'sveltekit-superforms/adapters';

  const schema = z.object({ name: z.string().min(1), email: z.string().email() });

  const form = superForm(defaults(zod(schema)), {
    validators: zod(schema),
    SPA: true,
    resetForm: true,
    async onUpdate({ form: f }) {
      if (!f.valid) return;
      // submit logic, toast on success/error
    }
  });

  const { form: formData, enhance, reset, submitting } = form;
  let isFormValid = $derived(schema.safeParse($formData).success);
</script>

<form method="POST" use:enhance>
  <Form.Field {form} name="email">
    {#snippet children({ constraints, errors, tainted, value })}
      <Form.Label>Email</Form.Label>
      <Input bind:value={$formData.email} {constraints} />
      {#if errors}<Form.FieldErrors {errors} />{/if}
    {/snippet}
  </Form.Field>
</form>
```

---

## 10. i18n (Paraglide)

```svelte
<script>
  import * as m from '$lib/paraglide/messages';
</script>

<h1>{m.feature_page_title()}</h1>
```

- Message files: `ui/messages/en.json`, `ui/messages/de.json`
- Keys are flat: `"feature_section_label": "Some Label"`
- Messages are functions (called per render for reactivity)
- `localizeHref(path)` for localized URLs

---

## 11. Styling

- **Tailwind CSS v4** — utility-first, CSS variable theming (`bg-card`, `text-muted-foreground`)
- **`cn()`** from `$lib/utils` — `twMerge(clsx(...inputs))` for conflict-free class composition
- **tailwind-variants** for component variants where needed
- **mode-watcher** for light/dark theme switching
- **Icons:** `@lucide/svelte` — import individual icons (`import { Plus } from '@lucide/svelte'`)

---

## 12. Event Handling

Use **callback props**, not custom events:

```svelte
<MyComponent
  onItemClick={(item) => { ... }}
  onDelete={(id) => { ... }}
  onViewsChange={(views) => { ... }}
/>
```

---

## 13. Services (`ui/src/services/`)

Lazy-init singletons with `$state` status tracking:

```ts
export class SomeService {
  status = $state<'idle' | 'initializing' | 'ready' | 'error'>('idle');
  private initPromise: Promise<void> | null = null;

  async init(): Promise<void> {
    if (this.status === 'ready') return;
    if (this.initPromise) return this.initPromise;
    this.initPromise = this._doInit();
    return this.initPromise;
  }
}
```

**Reference:** `ui/src/services/duckdb/duckdb-service.svelte.ts`

---

## 14. Key Aliases (import paths)

| Alias | Path |
|-------|------|
| `$lib` | `src/lib` |
| `$components` | `src/components` |
| `$apis` | `src/apis` |
| `$state` | `src/state` |
| `$utils` | `src/utils` |
| `$services` | `src/services` |

---

## 15. Performance Patterns

- **Lazy imports** for heavy components (Monaco, ECharts)
- **Virtual scrolling** via `@tanstack/virtual-core` for large lists (fixed 32px row height)
- **DuckDB-WASM** for client-side SQL on Arrow IPC data (no JSON conversion)
- **Viewport queries** (LIMIT/OFFSET) — only render what's visible
- **`svelte-check`** instead of `vite build` (project OOMs on full build)

---

These patterns are now loaded into context. Apply them consistently when building new components, routes, or features. Ask if anything is unclear before starting implementation.
