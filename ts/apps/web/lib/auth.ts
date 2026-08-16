import { env } from "./env";
import { getApiErrorMessage } from "./errors";

const SESSION_COOKIE = "limen_session";

export type AuthResult = { data: { ok: true }; error: null } | { data: null; error: string };

async function authFetch(path: string, body: Record<string, unknown>): Promise<AuthResult> {
  try {
    const res = await fetch(`${env.NEXT_PUBLIC_PLANETARY_API_URL}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      credentials: "include",
    });

    if (!res.ok) {
      let errorBody: unknown = null;
      try {
        errorBody = await res.json();
      } catch {
        // non-JSON error response
      }
      return {
        data: null,
        error: getApiErrorMessage(errorBody, res.statusText || "Request failed"),
      };
    }

    return { data: { ok: true }, error: null };
  } catch (err) {
    return {
      data: null,
      error: err instanceof Error ? err.message : "Network error",
    };
  }
}

export async function signin(values: { email: string; password: string }): Promise<AuthResult> {
  return authFetch("/api/auth/signin/credential", {
    credential: values.email,
    password: values.password,
  });
}

export async function signup(values: { email: string; password: string }): Promise<AuthResult> {
  return authFetch("/api/auth/signup/credential", values);
}

export async function signout(): Promise<void> {
  try {
    await fetch(`${env.NEXT_PUBLIC_PLANETARY_API_URL}/api/auth/signout`, {
      method: "POST",
      credentials: "include",
    });
  } catch {
    // best-effort; cookie will be cleared by the server or we redirect anyway
  }
}

export { SESSION_COOKIE };
