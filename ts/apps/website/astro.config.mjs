// @ts-check
import { defineConfig } from "astro/config";
import tailwindcss from "@tailwindcss/vite";
import node from "@astrojs/node";

export default defineConfig({
  site: "https://sunred.app",
  // The website is the front door for sunred.app: it serves its own pages
  // (marketing) and proxies every other path to the Next.js app (see
  // src/middleware.ts and src/pages/[...path].ts). Self-hosted deployments
  // that don't want the marketing site just run the Next.js app directly.
  output: "server",
  adapter: node({ mode: "standalone" }),
  vite: {
    plugins: [tailwindcss()],
  },
});
