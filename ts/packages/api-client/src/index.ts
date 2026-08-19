/**
 * Earthed API Client
 * Generated using @hey-api/openapi-ts
 */

import { createClient } from "./generated/client/client.gen";
import type { Config } from "./generated/client/types.gen";

const DEFAULT_BASE_URL = "http://localhost:8080";

export interface EarthedClientOptions extends Config {
  /**
   * Base URL of the Earthed API (no trailing slash). Paths in the spec already
   * include the `/v1` (or `/auth`) prefix, so this should be the host only
   * (e.g. `http://localhost:8080` or `https://api.earthed.app`).
   */
  baseUrl?: string;
}

/**
 * Create a typed client for the Earthed API.
 *
 * The returned client exposes typed methods for all API operations and a
 * `config`/`interceptors` surface compatible with the hey-api client-fetch
 * plugins (used by `lib/logger.ts` for request logging).
 */
export function createEarthedClient(options: EarthedClientOptions = {}) {
  const client = createClient({
    baseUrl: options.baseUrl ?? DEFAULT_BASE_URL,
    ...options,
  });
  return client;
}

export type EarthedClient = ReturnType<typeof createEarthedClient>;

// Re-export all generated types and functions
export * from "./generated/sdk.gen";
export * from "./generated/types.gen";
export * from "./generated/schemas.gen";
