import { env } from "./env";

/**
 * Sunred authentication is AT Proto OAuth. Users enter their handle; the web
 * app redirects to the API's OAuth login endpoint, which resolves the handle,
 * sends a PAR request to the user's PDS, and redirects the browser to the PDS
 * to approve. After approval the PDS redirects back to the API callback, which
 * issues a session cookie and redirects back here.
 *
 * All of that happens as full-page browser navigations (not fetch), because the
 * user leaves the Sunred origin to approve on their PDS.
 */

/**
 * loginWithHandle redirects the browser to the API to start the OAuth flow for
 * the given AT Proto handle (e.g. "alice.bsky.social"). After a successful
 * login the API redirects back to `redirectTo` (sanitized) or the app root.
 */
export function loginWithHandle(handle: string, redirectTo?: string | null): void {
  const params = new URLSearchParams({ handle });
  if (redirectTo) params.set("redirect", safeRedirect(redirectTo));
  window.location.assign(`${env.NEXT_PUBLIC_SUNRED_API_URL}/auth/oauth/login?${params}`);
}

/**
 * signupWithDefaultPDS redirects the browser to the API to start the OAuth
 * signup flow against the instance's default PDS (SUNRED_DEFAULT_PDS). The
 * API returns 503 if no default PDS is configured.
 */
export function signupWithDefaultPDS(redirectTo?: string | null): void {
  const params = new URLSearchParams();
  if (redirectTo) params.set("redirect", safeRedirect(redirectTo));
  const qs = params.toString();
  window.location.assign(
    `${env.NEXT_PUBLIC_SUNRED_API_URL}/auth/oauth/signup${qs ? `?${qs}` : ""}`,
  );
}

/**
 * getOAuthConfig fetches the public OAuth configuration from the API (whether
 * signup via a default PDS is available). Server-safe.
 */
export async function getOAuthConfig(): Promise<{ default_pds: string | null }> {
  try {
    const res = await fetch(`${env.NEXT_PUBLIC_SUNRED_API_URL}/auth/oauth/config`, {
      headers: { Accept: "application/json" },
      // Forward cookies on the server so the request is consistent.
      next: { revalidate: 60 },
    });
    if (!res.ok) return { default_pds: null };
    return await res.json();
  } catch {
    return { default_pds: null };
  }
}

/**
 * signout redirects to the API, which clears the session cookie and sends the
 * user back to the web app root.
 */
export function signout(): void {
  window.location.assign(`${env.NEXT_PUBLIC_SUNRED_API_URL}/auth/signout`);
}

/**
 * safeRedirect returns a same-origin path from a redirect query param, or "/"
 * if the value is missing/external/protocol-relative. Prevents open redirects
 * via ?redirect=https://evil.com or //evil.com. Pure/isomorphic — safe to
 * call from both server components and client components.
 */
export function safeRedirect(value: string | null | undefined): string {
  if (!value) return "/";
  // Must be a root-relative path: starts with "/" and not "//" (protocol-relative).
  if (value.startsWith("/") && !value.startsWith("//")) return value;
  return "/";
}
