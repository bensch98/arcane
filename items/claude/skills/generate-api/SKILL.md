---
name: generate-api
description: Generate frontend API module (types.ts + index.ts) from an OpenAPI JSON excerpt
disable-model-invocation: true
argument-hint: <paste OpenAPI JSON excerpt>
---

# Generate Frontend API Module

Generate a new API module in `ui/src/apis/` from the provided OpenAPI JSON excerpt.

## Input

The user will paste an OpenAPI JSON excerpt containing endpoint paths and optionally the schema definition. Parse:
- **Resource name** from the URL path (e.g., `/api/v1/tensoryze/sqlquery/` → `sqlquery`)
- **Base path prefix** to determine placement:
  - `/api/v1/tensoryze/<resource>/` → `ui/src/apis/tensoryze/<resource>/`
  - `/api/v1/integrations/<category>/<resource>/` → `ui/src/apis/integrations/<category>/` (add to existing module if it exists)
- **HTTP methods** (GET, POST, PUT, DELETE) and their parameters, request bodies, response schemas
- **Schema references** (`$ref: "#/components/schemas/ModelName"`) to derive TypeScript types

## OpenAPI excerpt

$ARGUMENTS

## Steps

1. **Parse the OpenAPI JSON** to extract all endpoints, their methods, parameters, request/response types.

2. **Derive TypeScript types** from schema references. If the schema definition is included, map fields directly. If only `$ref` names are given, infer field types from context (parameters, response shapes). Use `string | null` for optional/nullable fields, matching the backend Pydantic model.

3. **Check if the target directory already exists** (e.g., `ui/src/apis/tensoryze/<resource>/`). If it does, add new functions to the existing `index.ts` and new types to `types.ts`. If not, create the directory and both files.

4. **Generate `types.ts`** with exported type definitions.

5. **Generate `index.ts`** following this exact pattern for every endpoint:

```typescript
import { getAccessToken } from '$apis/auth-token';
import { BASE_URL } from '$apis/env';
import { type RequestLogContext, logRequest, logRequestError } from '$apis/utils';
import type { ModelName } from './types';

export const getFunctionName = async (
	fetchFn: typeof fetch = fetch,
	token: string,
	// additional params as needed (id, body, etc.)
): Promise<ReturnType> => {
	let error = null;

	if (!token) {
		token = getAccessToken();
	}

	const url = new URL(`${BASE_URL}/path/to/endpoint`);
	// For query params: url.searchParams.set('key', value);
	const method = 'GET'; // or POST, PUT, DELETE
	const ctx: RequestLogContext = { method, url }; // add `body` for POST/PUT

	const res = await fetchFn(url.toString(), {
		method,
		headers: {
			Accept: 'application/json',
			'Content-Type': 'application/json',
			Authorization: `Bearer ${token}`
		},
		// body: JSON.stringify(body)  // for POST/PUT only
	})
		.then(async (res) => {
			logRequest(ctx);
			if (!res.ok) throw await res.json();
			return res.json();
		})
		.catch((err) => {
			error = err.detail;
			logRequestError(ctx, err);
			return null;
		});

	if (error) {
		throw error;
	}

	return res as ReturnType;
};
```

## Naming conventions

- **Function names**: use camelCase descriptive names:
  - GET all → `get<Resources>` (plural)
  - GET by id → `get<Resource>` (singular)
  - GET by custom param → `get<Resources>By<Param>`
  - POST → `create<Resource>`
  - PUT → `update<Resource>`
  - DELETE → `delete<Resource>`
- **URL path**: Strip `/api/v1` prefix — `BASE_URL` already includes the API version. Use the remaining path.
- **Parameter types**: Use `string | number` for path params that accept `anyOf[integer, string]` per OpenAPI spec.
- **Return types**: Match the OpenAPI response schema — arrays return `ModelName[]`, single items may also return `ModelName[]` if the API does.
- **DELETE return type**: Use `string | number` if the response is `anyOf[integer, string]`.

## Rules

- Follow the EXACT pattern shown above — do not simplify, abstract, or deviate from the error handling structure.
- Every function must have the `fetchFn` and `token` parameters first, matching the existing codebase convention.
- POST/PUT functions include `const body = paramName;` and `const ctx: RequestLogContext = { method, url, body };`.
- GET/DELETE functions use `const ctx: RequestLogContext = { method, url };`.
- Do not add comments or docstrings to the generated code.
- After generating, confirm the files created and list all exported functions.
