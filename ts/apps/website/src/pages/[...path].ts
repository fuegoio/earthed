import type { APIContext } from "astro";
import { proxyToApp } from "../proxy";

// Catch-all: every path the website doesn't own (no matching .astro page and
// no static asset) is proxied to the Next.js app. This is the linear.app-style
// fallthrough — add a page under src/pages/ and it's served instead of proxied.
export const ALL = (context: APIContext): Promise<Response> => proxyToApp(context);
