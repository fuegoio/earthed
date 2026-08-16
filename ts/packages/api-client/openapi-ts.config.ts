import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../../../go/api/openapi.json",
  output: {
    path: "./src/generated",
    clean: true,
  },
  plugins: ["@hey-api/client-fetch", "@hey-api/schemas", "@hey-api/sdk", "@hey-api/typescript"],
});
