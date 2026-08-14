import type { NextConfig } from "next"

const nextConfig: NextConfig = {
  transpilePackages: ["@workspace/ui", "@planetary/api-client"],
  output: "standalone",
}

export default nextConfig
