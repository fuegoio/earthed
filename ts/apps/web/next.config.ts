import type { NextConfig } from "next";
import { withSerwist } from "@serwist/turbopack";

const nextConfig: NextConfig = {
  transpilePackages: ["@workspace/ui", "@earthed/api-client"],
  output: "standalone",
};

export default withSerwist(nextConfig);
