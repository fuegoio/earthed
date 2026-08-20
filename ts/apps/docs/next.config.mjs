import { createMDX } from "fumadocs-mdx/next";
import path from "path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/** @type {import("next").NextConfig} */
const config = {
  output: "standalone",
  outputFileTracingRoot: path.join(__dirname, "../../"),
  reactStrictMode: true,
  // Mounted under /docs on earthed.app via the website proxy. All routes,
  // _next assets, and next/link redirects are prefixed accordingly.
  basePath: "/docs",
  turbopack: {
    root: path.join(__dirname, "../../"),
  },
};

const withMDX = createMDX({
  configPath: "source.config.ts",
});

export default withMDX(config);
