import { createEarthedClient, type EarthedClient } from "@earthed/api-client";
import { attachApiLogger } from "./logger";
import { env } from "./env";

const fetchNoStore: typeof fetch = (input, init) => fetch(input, { ...init, cache: "no-store" });

/**
 * Adapts the SDK's `{ data, error }` discriminated union to react-query's
 * throw-on-error contract: returns `data` on success, throws the parsed huma
 * `ErrorModel` body on failure so `getApiErrorMessage(query.error)` keeps
 * working in components.
 *
 *   const folders = await unwrap(
 *     listFolders({ client: await getClient() }),
 *   );
 */
export async function unwrap<T>(
  result: Promise<{ data: T | null | undefined; error: unknown }>,
): Promise<T> {
  const { data, error } = await result;
  if (error) throw error;
  return data as T;
}

export async function getClient(cookieHeader?: string): Promise<EarthedClient> {
  const isServer = typeof window === "undefined";
  const headers: Record<string, string> = {};

  if (isServer) {
    const { cookies } = await import("next/headers");
    const ch = cookieHeader ?? (await cookies()).toString();
    if (ch) headers.Cookie = ch;
  }

  const client = createEarthedClient({
    baseUrl: isServer ? env.EARTHED_API_URL : env.NEXT_PUBLIC_EARTHED_API_URL,
    fetch: fetchNoStore,
    headers: Object.keys(headers).length > 0 ? headers : undefined,
    credentials: isServer ? undefined : "include",
  });

  attachApiLogger(client, { isServer });

  return client;
}

export * from "@earthed/api-client";
