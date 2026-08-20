import type { APIContext } from "astro";

// Origin of the Next.js app the website proxies unknown paths to. Defaults
// to the local dev server; set APP_URL (e.g. http://web:3000) in production.
const APP_URL = process.env.APP_URL?.replace(/\/+$/, "") ?? "http://localhost:3000";

// Hop-by-hop headers that must not be forwarded across a proxy boundary.
const HOP_BY_HOP = new Set([
  "connection",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailer",
  "transfer-encoding",
  "upgrade",
]);

/**
 * proxyToApp forwards the incoming request to the Next.js app and returns its
 * response. It is used by the catch-all route (every path the website doesn't
 * own) and by the middleware (logged-in "/" so the app renders there).
 *
 * The app's absolute redirect Location headers are rewritten to the public
 * origin so redirects (e.g. unauth /feeds -> /login) stay on the public host
 * instead of leaking the internal app URL.
 */
export async function proxyToApp(context: APIContext): Promise<Response> {
  const { url, request } = context;
  const target = new URL(url.pathname + url.search, APP_URL);

  const reqHeaders = new Headers();
  for (const [key, value] of request.headers) {
    if (HOP_BY_HOP.has(key.toLowerCase())) continue;
    reqHeaders.set(key, value);
  }
  // Let the app build public URLs (redirects, canonical links) from the
  // original request instead of the internal proxy address.
  reqHeaders.set("x-forwarded-host", url.host);
  reqHeaders.set("x-forwarded-proto", url.protocol.replace(":", ""));
  reqHeaders.delete("x-forwarded-for");
  reqHeaders.set("x-forwarded-for", request.headers.get("x-forwarded-for") ?? "");

  const init: RequestInit = {
    method: request.method,
    headers: reqHeaders,
    redirect: "manual",
    // Forward the body for non-GET/HEAD so server actions and form posts pass
    // through. `duplex: "half"` is required by undici when streaming a body.
    body: request.method !== "GET" && request.method !== "HEAD" ? request.body : undefined,
    // @ts-expect-error duplex is a non-standard but required option for streamed bodies.
    duplex: "half",
  };

  const upstream = await fetch(target, init);

  const resHeaders = new Headers();
  for (const [key, value] of upstream.headers) {
    if (HOP_BY_HOP.has(key.toLowerCase())) continue;
    if (key.toLowerCase() === "location") {
      resHeaders.set(key, rewriteLocation(value, url));
    } else {
      resHeaders.set(key, value);
    }
  }

  return new Response(upstream.body, {
    status: upstream.status,
    statusText: upstream.statusText,
    headers: resHeaders,
  });
}

/** Rewrite an absolute redirect Location to the public origin. */
function rewriteLocation(location: string, publicUrl: URL): string {
  if (!/^https?:\/\//.test(location)) return location;
  try {
    const loc = new URL(location);
    loc.protocol = publicUrl.protocol;
    loc.host = publicUrl.host;
    return loc.toString();
  } catch {
    return location;
  }
}
