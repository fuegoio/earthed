import pino, { type Logger } from "pino";

/**
 * Server-side structured logger. Imported only from server code paths; the
 * API logger wired in `lib/earthed.ts` guards on `typeof window` before
 * calling into it so the pino dependency is kept out of the browser bundle.
 */
export const logger: Logger = pino({
  level: process.env.LOG_LEVEL ?? "info",
});

// ---------------------------------------------------------------------------
// API call logging
// ---------------------------------------------------------------------------

/** Subset of pino's Logger used by `attachApiLogger`. */
export type ApiLogger = Pick<Logger, "info" | "warn" | "error">;

/** Structural shape of a hey-api client's interceptor surface. */
interface InterceptableClient {
  interceptors: {
    // Method syntax (bivariant) so the real typed `Client` is assignable
    // without `lib/logger.ts` having to import from `@earthed/api-client`.
    request: {
      use(fn: (request: unknown, options: unknown) => unknown): void;
    };
    response: {
      use(fn: (response: unknown, request: unknown, options: unknown) => unknown): void;
    };
    error: {
      use(
        fn: (error: unknown, response: unknown, request: unknown, options: unknown) => unknown,
      ): void;
    };
  };
}

interface InterceptedResponse {
  status?: number;
  ok?: boolean;
}
interface InterceptedRequest {
  method?: string;
  url?: string;
}

/** Start time per in-flight request, keyed by the Request object. */
const requestStart = new WeakMap<object, number>();

/**
 * Wire pino logging into a hey-api client. On the server, every API call is
 * logged with its status, method, url and duration:
 *   - 2xx → `info`
 *   - 4xx → `warn` (with the parsed error body)
 *   - 5xx and network failures → `error` (with the parsed error body)
 *
 * The interceptor is **logging-only**: it returns the thrown value unchanged so
 * callers (e.g. `getApiErrorMessage` in `lib/errors.ts`) keep seeing the raw
 * huma `ErrorModel` body. On the client it's a no-op (pino is server-only).
 *
 * `client` is typed as `unknown` and cast internally so `lib/logger.ts` doesn't
 * have to import from `@earthed/api-client` (only `lib/earthed.ts` may).
 * `log` is injectable so tests can capture it; production passes the shared
 * pino logger above.
 */
export function attachApiLogger(
  client: unknown,
  options: { log?: ApiLogger; isServer?: boolean } = {},
): void {
  const log = options.log ?? logger;
  const isServer = options.isServer ?? typeof window === "undefined";
  if (!isServer) return;
  const c = client as InterceptableClient;

  c.interceptors.request.use((request) => {
    if (request && typeof request === "object") requestStart.set(request, performance.now());
    return request;
  });

  c.interceptors.response.use((response, request) => {
    const res = response as InterceptedResponse | undefined;
    if (!res?.ok) return response; // non-2xx handled by the error interceptor
    const { method, url, duration } = trace(request);
    log.info({ status: res.status, method, url, duration }, `api ${res.status} ${method} ${url}`);
    return response;
  });

  c.interceptors.error.use((error, response, request) => {
    const status = (response as InterceptedResponse | undefined)?.status ?? 0;
    const { method, url, duration } = trace(request);
    const level: "warn" | "error" = status === 0 || status >= 500 ? "error" : "warn";
    log[level](
      { status, method, url, duration, body: error },
      `api ${status || "error"} ${method} ${url}`,
    );
    return error;
  });
}

function trace(request: unknown): {
  method: string;
  url: string;
  duration: number | undefined;
} {
  const req = request as InterceptedRequest | undefined;
  const start = request && typeof request === "object" ? requestStart.get(request) : undefined;
  return {
    method: req?.method ?? "GET",
    url: req?.url ?? "",
    duration: start ? Math.round(performance.now() - start) : undefined,
  };
}
