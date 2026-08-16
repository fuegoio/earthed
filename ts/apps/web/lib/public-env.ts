import { createPublicEnv } from "next-public-env"

// Access via a computed key so Next.js does NOT inline the build-time value of
// process.env.NEXT_PUBLIC_* into the bundle. This lets next-public-env read the
// runtime value on the server and inject it into the client.
const PUBLIC_API_URL_KEY = "NEXT_PUBLIC_PLANETARY_API_URL"

export const { getPublicEnv, PublicEnv } = createPublicEnv(
  {
    NEXT_PUBLIC_PLANETARY_API_URL: process.env[PUBLIC_API_URL_KEY],
  },
  {
    schema: (z) => ({
      NEXT_PUBLIC_PLANETARY_API_URL: z
        .string()
        .url()
        .default("http://localhost:8080"),
    }),
  }
)
