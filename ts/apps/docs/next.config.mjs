import { createMDX } from "fumadocs-mdx/next";
import path from "path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/** @type {import("next").NextConfig} */
const config = {
  output: "standalone",
  outputFileTracingRoot: path.join(__dirname, "../../"),
  reactStrictMode: true,
  turbopack: {
    root: path.join(__dirname, "../../"),
  },
};

const withMDX = createMDX({
  configPath: "source.config.ts",
});

export default withMDX(config);
