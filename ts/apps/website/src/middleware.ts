import { defineMiddleware } from "astro:middleware";
import { proxyToApp } from "./proxy";

export const onRequest = defineMiddleware(async (context, next) => {
  // Logged-in visitors of "/" get the app instead of the marketing page. The
  // session cookie is only a presence signal here; the app layout validates
  // it and redirects to /login if it's expired — no loop, since /login is a
  // public route that the catch-all proxies straight back to the app.
  if (context.url.pathname === "/" && context.cookies.get("limen_session")?.value) {
    return proxyToApp(context);
  }
  return next();
});
