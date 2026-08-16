import { createEnv } from "@t3-oss/env-nextjs"
import { z } from "zod"
import { getPublicEnv } from "./public-env"

// `NEXT_PUBLIC_*` vars are inlined into the client bundle at build time, so a
// value set at runtime in the deployment never reaches client code. We use
// `next-public-env` (lib/public-env.ts) to inject the public API URL at request
// time on the server and read it via `getPublicEnv()` on both server and client.
// The server-only `PLANETARY_API_URL` is unaffected (read from `process.env`).
export const env = createEnv({
  server: {
    PLANETARY_API_URL: z.string().url().default("http://localhost:8080"),
  },
  client: {
    NEXT_PUBLIC_PLANETARY_API_URL: z
      .string()
      .url()
      .default("http://localhost:8080"),
  },
  runtimeEnv: {
    PLANETARY_API_URL: process.env.PLANETARY_API_URL,
    NEXT_PUBLIC_PLANETARY_API_URL: getPublicEnv().NEXT_PUBLIC_PLANETARY_API_URL,
  },
})
