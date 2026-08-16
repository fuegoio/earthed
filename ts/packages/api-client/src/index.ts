/**
 * Planetary API Client
 * Generated using @hey-api/openapi-ts
 */

import { createClient } from "./generated/client/client.gen";
import type { Config } from "./generated/client/types.gen";

const DEFAULT_BASE_URL = "http://localhost:8080";

export interface PlanetaryClientOptions extends Config {
  /**
   * Base URL of the Planetary API (no trailing slash). Paths in the spec already
   * include the `/api/v1` prefix, so this should be the host only
   * (e.g. `http://localhost:8080`).
   */
  baseUrl?: string;
}

/**
 * Create a typed client for the Planetary API.
 *
 * The returned client exposes typed methods for all API operations and a
 * `config`/`interceptors` surface compatible with the hey-api client-fetch
 * plugins (used by `lib/logger.ts` for request logging).
 */
export function createPlanetaryClient(options: PlanetaryClientOptions = {}) {
  const client = createClient({
    baseUrl: options.baseUrl ?? DEFAULT_BASE_URL,
    ...options,
  });
  return client;
}

export type PlanetaryClient = ReturnType<typeof createPlanetaryClient>;

// Re-export all generated types and functions
export * from "./generated/sdk.gen";
export * from "./generated/types.gen";
export * from "./generated/schemas.gen";
