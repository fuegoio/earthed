// ---------------------------------------------------------------------------
// huma error parsing
// ---------------------------------------------------------------------------

/**
 * The `@earthed/api-client` resolves failed responses as the parsed JSON body
 * (not an `Error` instance) — the SDK functions return a `{ data, error }`
 * discriminated union and `unwrap()` in `lib/earthed.ts` throws `error` on
 * failure. The Earthed API (huma) serializes errors using the `ErrorModel`
 * schema:
 *
 *   {
 *     "title": "Forbidden",
 *     "status": 403,
 *     "detail": "...",
 *     "errors": [{ "message": "insufficient permissions ...", "location": "...", "value": ... }]
 *   }
 *
 * The helpers below also tolerate plain `Error` instances (e.g. network
 * failures), plain strings, and auth-style error objects
 * (`{ message }` or `{ error: { message } }`).
 */

interface HumaErrorDetail {
  message?: string;
  location?: string;
  value?: unknown;
}

interface HumaErrorModel {
  title?: string;
  status?: number;
  detail?: string;
  errors?: HumaErrorDetail[] | null;
  message?: string;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function readHumaMessage(value: unknown): string | undefined {
  if (!isRecord(value)) return undefined;
  const model = value as HumaErrorModel;

  if (Array.isArray(model.errors)) {
    const first = model.errors.find((e) => typeof e?.message === "string" && e.message.length > 0);
    if (first?.message) return first.message;
  }
  if (typeof model.detail === "string" && model.detail.length > 0) return model.detail;
  if (typeof model.message === "string" && model.message.length > 0) return model.message;
  if (typeof model.title === "string" && model.title.length > 0) return model.title;
  return undefined;
}

/** Extract a human-readable message from an error thrown by the API client. */
export function getApiErrorMessage(error: unknown, fallback = "Request failed"): string {
  if (error == null) return fallback;

  if (typeof error === "string") return error.length > 0 ? error : fallback;

  if (error instanceof Error) {
    return error.message.length > 0 ? error.message : fallback;
  }

  if (isRecord(error)) {
    const huma = readHumaMessage(error);
    if (huma) return huma;

    // auth style: `{ error: { message } }` or `{ error: { code } }`.
    const nested = error.error;
    if (isRecord(nested)) {
      const msg =
        readHumaMessage(nested) ??
        (typeof nested.message === "string" ? nested.message : undefined);
      if (msg) return msg;
    }
  }

  return fallback;
}

/** Read the HTTP status off a thrown API error (huma `ErrorModel` carries `status`). */
export function apiErrorStatus(error: unknown): number | undefined {
  if (isRecord(error) && "status" in error) {
    const status = (error as { status?: unknown }).status;
    return typeof status === "number" ? status : undefined;
  }
  return undefined;
}

/** True when the thrown error is a 4xx API response. */
export function isClientError(error: unknown): boolean {
  const status = apiErrorStatus(error);
  return typeof status === "number" && status >= 400 && status < 500;
}
