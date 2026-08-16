/// <reference lib="esnext" />
/// <reference lib="webworker" />
import { defaultCache } from "@serwist/turbopack/worker";
import type { PrecacheEntry, SerwistGlobalConfig } from "serwist";
import { Serwist, NetworkFirst, ExpirationPlugin } from "serwist";

declare global {
  interface WorkerGlobalScope extends SerwistGlobalConfig {
    __SW_MANIFEST: (PrecacheEntry | string)[] | undefined;
  }
}

declare const self: ServiceWorkerGlobalScope;

const serwist = new Serwist({
  precacheEntries: self.__SW_MANIFEST,
  skipWaiting: true,
  clientsClaim: true,
  navigationPreload: true,
  runtimeCaching: [
    // API JSON responses — NetworkFirst so the latest data is served when
    // online, and the last successful response is cached for offline use.
    // Only GETs are matched; mutations (PUT/POST/DELETE) pass through.
    {
      matcher: ({ url }) => url.pathname.startsWith("/api/v1/"),
      method: "GET",
      handler: new NetworkFirst({
        cacheName: "api-cache",
        networkTimeoutSeconds: 10,
        plugins: [
          new ExpirationPlugin({
            maxEntries: 100,
            maxAgeSeconds: 60 * 60 * 24 * 7, // 7 days
          }),
        ],
      }),
    },
    // SSR documents (HTML pages) — NetworkFirst so the fresh HTML is
    // served when online, and the last visited page is cached for offline.
    {
      matcher: ({ request }) => request.destination === "document",
      method: "GET",
      handler: new NetworkFirst({
        cacheName: "page-cache",
        networkTimeoutSeconds: 10,
        plugins: [
          new ExpirationPlugin({
            maxEntries: 50,
            maxAgeSeconds: 60 * 60 * 24 * 7, // 7 days
          }),
        ],
      }),
    },
    // Default caching for static assets (fonts, images, etc.)
    ...defaultCache,
  ],
  fallbacks: {
    entries: [
      {
        url: "/~offline",
        matcher({ request }) {
          return request.destination === "document";
        },
      },
    ],
  },
});

serwist.addEventListeners();
