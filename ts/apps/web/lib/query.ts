import { QueryClient, dehydrate, HydrationBoundary } from "@tanstack/react-query";

/**
 * Shared `QueryClient` factory. Used both by the client-side `QueryProvider`
 * (one per browser session) and by server components / layouts that prefetch
 * queries before rendering. Keeping the factory in one place guarantees the
 * default options stay in sync.
 */
export function makeQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: 1,
      },
    },
  });
}

export { dehydrate, HydrationBoundary };
