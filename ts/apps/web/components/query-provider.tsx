"use client"

import { useState } from "react"
import { QueryClientProvider } from "@tanstack/react-query"
import { makeQueryClient } from "@/lib/query"

/**
 * App-wide react-query provider. The `QueryClient` is created once per browser
 * session (memoized via `useState`) so cache is shared across navigations but
 * not across SSR requests — each server render gets a fresh client. Server
 * components that prefetch queries create their own throw-away client via
 * `makeQueryClient()` and pass the dehydrated state through `HydrationBoundary`.
 */
export function QueryProvider({ children }: { children: React.ReactNode }) {
  const [client] = useState(() => makeQueryClient())

  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}
