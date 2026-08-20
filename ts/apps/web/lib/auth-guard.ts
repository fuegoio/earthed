import { redirect } from "next/navigation";
import { getClient, getMe } from "@/lib/earthed";
import { isClientError } from "@/lib/errors";
import { safeRedirect } from "@/lib/auth";

/**
 * redirectIfAuthenticated validates the current session against the API and,
 * if the user is signed in, redirects to the `redirect` target (sanitized) or
 * `/`. Call this from server components that render auth pages (/login,
 * /signup) so logged-in visitors skip the form entirely.
 *
 * On an invalid/missing session (4xx) or an API error (5xx/network), it does
 * nothing and lets the caller render the auth form — never redirect to / on
 * an API outage, since that would loop with the app layout's redirect to
 * /login.
 */
export async function redirectIfAuthenticated(redirectTo: string | null | undefined): Promise<void> {
  const client = await getClient();
  try {
    const result = await getMe({ client });
    if (result.error) {
      // 4xx → not authenticated; render the form.
      if (isClientError(result.error)) return;
      // 5xx → API unhealthy; don't redirect (avoid loops), render the form.
      return;
    }
    if (result.data) {
      redirect(safeRedirect(redirectTo));
    }
  } catch {
    // Network failure → render the form rather than risk a redirect loop.
  }
}
