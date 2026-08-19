# Repository conventions

## Commits

This repository uses **Conventional Commits**.

Commit messages must follow the format:

```
<type>(<optional scope>): <description>

<optional body>
<optional footer>
```

- Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`.
- Scope is optional and lowercased (e.g. `api`, `web`, `keys`, `store`).
- Description is lowercase, imperative mood, no trailing period.
- Breaking changes use a `!` after the type/scope (e.g. `feat(api)!: ...`) and a `BREAKING CHANGE:` footer.

Examples:

```
fix(keys): use authorized_keys format and surface store errors as 500
feat(web): add repository settings page
docs: clarify API token creation flow
```

## Documentation

The docs site lives in `ts/apps/docs/` and is a Fumadocs (Next.js) workspace. It
renders the API reference from `go/api/openapi.json` and ships hand-written
guides under `ts/apps/docs/content/docs/`. The documentation is **public-facing**:
it should only contain information useful to users of Planetary, not
internal implementation details (package layout, file paths, database
schema, test coverage, etc. belong in code comments or internal docs,
not in the docs site).

When you change the API, keep the docs in sync if the change is covered
there:

- **OpenAPI spec** — the reference at `ts/apps/docs/openapi/*` is generated from
  `go/api/openapi.json`. After changing Go handlers or types, run
  `make gen-openapi` from `go/api/` — it regenerates `go/api/openapi.json`,
  copies it to `ts/apps/docs/openapi/openapi.json`, and regenerates the
  TypeScript client at `ts/packages/api-client/src/generated/`. Commit all
  changed files.
- **Guides** (`ts/apps/docs/content/docs/*.mdx`) — if your change affects
  behaviour described in a guide (e.g. roles in `roles.mdx`, setup steps
  in `getting-started.mdx`), update the guide in the same PR.
- **New endpoints or roles** — if you add an endpoint that introduces a
  concept worth documenting, or change how access roles work, update the
  relevant guide rather than relying on the generated reference alone.

When in doubt, search `ts/apps/docs/content/docs/` for the topic you touched; if
it's mentioned, update it.
