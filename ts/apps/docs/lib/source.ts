import { loader } from "fumadocs-core/source";
import { docs } from "collections/server";
import { openapi } from "./openapi";

// Three sections served from a single source so the sidebar can cross-link them:
//   /docs          — product docs   (content/docs/product/)
//   /self-hosting  — self-hosting   (content/docs/self-hosting/)
//   /api-reference — OpenAPI spec   (content/docs/openapi/)
export const source = loader(
  {
    docs: docs.toFumadocsSource(),
    openapi: await openapi.staticSource({
      baseDir: "openapi",
      per: "operation",
      groupBy: "tag",
    }),
  },
  {
    baseUrl: "/",
    url: (slugs) => {
      if (slugs[0] === "openapi") {
        const rest = slugs.slice(1);
        return rest.length ? `/api-reference/${rest.join("/")}` : "/api-reference";
      }
      if (slugs[0] === "self-hosting") {
        const rest = slugs.slice(1);
        return rest.length ? `/self-hosting/${rest.join("/")}` : "/self-hosting";
      }
      if (slugs[0] === "product") {
        const rest = slugs.slice(1);
        return rest.length ? `/docs/${rest.join("/")}` : "/docs";
      }
      return slugs.length ? `/docs/${slugs.join("/")}` : "/docs";
    },
    plugins: [openapi.loaderPlugin()],
  },
);
