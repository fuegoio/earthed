<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->

# Frontend conventions

## Data fetching: TanStack Query + the Planetary SDK

All data fetching and mutations go through the **`@planetary/api-client`** typed SDK, re-exported from `lib/planetary.ts`. **Client-side reads use TanStack Query (`useQuery` / `useInfiniteQuery`)** — never hand-rolled `useEffect` + `useState` fetch loops.

- **Client components** call `useQuery` with a `queryKey` (a stable array identifying the query) and a `queryFn` that goes through the SDK. Use the `unwrap()` helper from `lib/planetary.ts` to adapt the SDK's `{ data, error }` discriminated union to react-query's throw-on-error contract:

  ```ts
  import { useQuery } from "@tanstack/react-query";
  import { getClient, listCategories, unwrap } from "@/lib/planetary";

  const {
    data: categories,
    error,
    isLoading,
    refetch,
  } = useQuery<Category[]>({
    queryKey: ["categories"],
    queryFn: async () => unwrap(listCategories({ client: await getClient() })),
  });
  ```

  On error, `unwrap()` throws the parsed huma `ErrorModel` body so `getApiErrorMessage(error)` keeps working in the component. `isLoading` is true only on the first fetch with no cached data — background refetches keep the stale data visible, which is what prevents flicker on navigation.

- **Server components** still call `getClient()` directly and `await` the SDK call (`const { data, error } = await listCategories({ client, ... })`), since `useQuery` is client-only. Render `<ApiError />` on failure as below.
- **Client-side mutations** (forms, delete actions, etc.) stay as event handlers calling the SDK directly (`const { error } = await createCategory({ client, ... })`) — not `useMutation` unless you need its cache/optimistic machinery. After a successful mutation, call `queryClient.invalidateQueries({ queryKey: [...] })` (or `router.refresh()`) to refetch affected reads.
- `getClient()` is isomorphic: on the server it forwards cookies via `next/headers` (dynamically imported to keep it out of the client bundle) and uses `env.PLANETARY_API_URL`; on the client it relies on the browser sending cookies automatically (`credentials: "include"`) and uses `env.NEXT_PUBLIC_PLANETARY_API_URL`.
- The `QueryClient` is provided once at the root via `components/query-provider.tsx` (mounted in `app/layout.tsx`). Do not instantiate `QueryClient` per-component; use `useQueryClient()` if you need imperative access from a child.
- Never use `fetch`, `axios`, or raw API calls in components, hooks, or pages. Go through `lib/planetary.ts`. The only layer allowed to import `@planetary/api-client` is `lib/planetary.ts`.
- Environment variables are validated via `@t3-oss/env-nextjs` in `lib/env.ts`. Access `env.PLANETARY_API_URL` (server) or `env.NEXT_PUBLIC_PLANETARY_API_URL` (client) instead of `process.env.*`.

## Avoid `useEffect`

`useEffect` is an escape hatch for synchronizing with **external systems** (non-React widgets, browser APIs, subscriptions). Most uses in app code are an anti-pattern. Before reaching for an effect, check the alternatives:

- **Fetching data** → `useQuery` (see above). Do not write `useEffect` + `setState` fetch loops; they re-implement caching, dedup, stale-while-revalidate, and request cancellation poorly.
- **Deriving values from props/state** → compute during render (`const filtered = entries.filter(...)`). Wrap expensive derivations in `useMemo`, not state + effect.
- **Resetting state when a prop changes** → give the component a `key` prop that changes, or reset in the event handler that caused the change — not an effect that calls `setState`.
- **Responding to user events** → put the logic in the event handler (`onClick`, `onSubmit`), not an effect that watches state.
- **Notifying a parent** → call the callback from the event handler, not from an effect.

Valid effect use cases: subscribing to a browser API (`matchMedia`, `IntersectionObserver`), third-party widget imperatives, and one-time setup that genuinely has no React equivalent. If you find yourself writing an effect that calls `setState` or fetches data, stop and reconsider.

## Forms: react-hook-form + Zod resolver

All forms use **react-hook-form** with **@hookform/resolvers/zod** for schema validation.

- Define one Zod schema per form, colocated with the form component (or in `lib/schemas.ts` for shared schemas).
- Infer TS types from the schema with `z.infer<typeof schema>` — do not hand-write form value types.
- Pass the schema to `useForm({ resolver: zodResolver(schema) })`.
- Do not introduce alternative form libraries (Formik, final-form, uncontrolled-native, etc.).

## Deletion: shadcn AlertDialog

Destructive delete actions use the **shadcn `AlertDialog`** (from `components/ui/alert-dialog.tsx`) composed directly inside the business component that owns the resource — never a shared half-DS `DeleteButton` wrapper.

- Each component that can delete a resource renders its own `AlertDialog` with the trigger `Button` (`variant="destructive"`) and the confirm/cancel actions inline. The mutation call (`deleteCategory`, `deleteFeed`, `deleteToken`, … from `lib/planetary.ts`) lives in the same component.
- Do **not** add a generic `components/delete-button.tsx` that takes a `kind` discriminator and switches on resource type. This does not scale beyond a couple of resources, hides the mutation behind a prop API, and forces every new resource to extend a union. Composing the `AlertDialog` inline keeps each delete flow local to its resource.
- Error handling follows the client-mutation rule above: `const { error } = await sdk(...)` then `toast.error(getApiErrorMessage(error))`. Use `router.refresh()` (or an optimistic update) after success.

## Empty states: shadcn Empty

Empty states use the **shadcn `Empty`** component family from `components/ui/empty.tsx` (`Empty`, `EmptyHeader`, `EmptyMedia`, `EmptyTitle`, `EmptyDescription`, `EmptyContent`) — never a bare `<p className="text-muted-foreground">` or ad-hoc centered div.

- Compose following the shadcn outline: `Empty > EmptyHeader > { EmptyMedia (variant="icon" with a lucide-react icon), EmptyTitle, EmptyDescription }` then `EmptyContent` for the primary action (usually a `Button` or `buttonVariants()` link).
- Prefer the `border` utility on `Empty` (outlined style) to match the surrounding surface density.
- Icons come from `lucide-react` (the only icon library in the app); pick a semantically relevant icon (e.g. `Rss` for feeds, `Folder` for categories, `Key` for tokens).
- The empty state owns the primary CTA for that view — if a "New X" button already lives in the page header, the `EmptyContent` CTA should link to the same destination so users can act from either place.

## Errors and logging

All API error handling lives in **`lib/errors.ts`**. On error the SDK resolves `error` to the parsed huma `ErrorModel` body (`{ title, status, detail, errors: [{ message }] }`) — not an `Error` — so always parse it with `getApiErrorMessage(e)` / `apiErrorStatus(e)`, never `e instanceof Error`.

- **Server-side fetches**: destructure `const { data, error } = await sdk(...)` and render `<ApiError message={getApiErrorMessage(error)} status={apiErrorStatus(error)} />` from `components/api-error.tsx` on failure. Never use `.catch(() => null)` — it silently swallows errors. 404s may keep their existing not-found / redirect behavior.
- **Client-side mutations** (forms, actions): destructure `const { error } = await sdk(...)` and call `toast.error(getApiErrorMessage(error))` on failure. The `<Toaster />` is mounted once in `app/layout.tsx`; import `toast` from `sonner` directly. Do not store errors in component state to render them as inline `<div>`s (field-level `formState.errors` from react-hook-form is the only inline exception). Success confirmations may use `toast.success(...)`.
- **Logging** (server-only, via `lib/logger.ts`): `getClient()` wires `attachApiLogger()`, which logs **every** API call with `{ status, method, url, duration }` — `info` on 2xx, `warn` on 4xx, `error` on 5xx and network failures (the latter also include the parsed `body`). The interceptor is logging-only: it returns the error value unchanged so `getApiErrorMessage` keeps seeing the raw huma body.

## Stack reference

- Next.js 16 (App Router), React 19, TypeScript, Tailwind v4, shadcn/ui. UI components and primitives live in `packages/ui` (`@workspace/ui`); app-specific components live in `components/`.
- Data fetching (client): TanStack Query v5, provided via `components/query-provider.tsx`. `useQuery` for reads, optional `useMutation` for mutations with cache invalidation.
- API client: `@planetary/api-client` (in `packages/api-client/`), generated from the OpenAPI spec via `@hey-api/openapi-ts`. Types come from `api/openapi.json`. Regenerate with `pnpm --filter @planetary/api-client gen`.
- API access: `lib/planetary.ts` — exports `getClient()` (an isomorphic factory that creates a `PlanetaryClient` with cookie forwarding on server, `credentials: "include"` on client, `no-store` fetch), `unwrap()` (adapts `{ data, error }` to react-query's throw-on-error contract), and re-exports all SDK functions and types from `@planetary/api-client`.
- Model types: `lib/types.ts` — re-exports named types from `@planetary/api-client` with web-friendly aliases (`APIToken`, `CreatedToken`).
- API error handling: `lib/errors.ts` — `getApiErrorMessage(e)` (huma `ErrorModel`), `apiErrorStatus(e)`, `isClientError(e)`.
- Logging: `lib/logger.ts` — shared pino `logger` and `attachApiLogger(client)` (pino-backed server log of every API call). `components/api-error.tsx` renders the server-side fetch error banner.
- Shared Zod schemas for forms: `lib/schemas.ts`.
- Environment: `lib/env.ts` — t3-env validated. Server-only vars like `PLANETARY_API_URL`.
- Zod for all runtime validation, both client and server.
