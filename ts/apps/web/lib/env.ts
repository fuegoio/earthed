import { createEnv } from "@t3-oss/env-nextjs"
import { z } from "zod"

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
    NEXT_PUBLIC_PLANETARY_API_URL: process.env.NEXT_PUBLIC_PLANETARY_API_URL,
  },
})
