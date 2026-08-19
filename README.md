# Earthed

A self-hosted RSS reader.

## Repository structure

```
earthed/
├── go/                      # Go (API server, SDK, CLI/TUI)
│   ├── go.work              # Go workspace — links all modules
│   ├── api/                 # API server (huma, PostgreSQL)
│   ├── sdk/                 # Go client generated from OpenAPI (oapi-codegen)
│   └── cli/                 # CLI + TUI (cobra, bubbletea)
├── ts/                      # TypeScript (web, docs, shared packages)
│   ├── apps/
│   │   ├── web/             # Next.js frontend
│   │   └── docs/            # Fumadocs documentation site
│   └── packages/
│       ├── api-client/      # TS client generated from OpenAPI (openapi-ts)
│       ├── ui/              # Shared UI components
│       ├── eslint-config/
│       └── typescript-config/
├── Makefile                 # Root orchestration (make gen, make dev, ...)
└── .github/workflows/       # CI
```

## Quick start

### Prerequisites

- Go 1.25+
- Node.js 20+ with pnpm 10+
- PostgreSQL 16+ (or Docker for local dev)

### API server

```bash
cd go/api
make db-up          # start PostgreSQL via docker compose
make migrate        # run database migrations
make run            # start the API server on :8080
```

### CLI

```bash
cd go/cli
make build
./earthed config set base_url http://localhost:8080
./earthed config set token <your-api-token>
./earthed feeds list
./earthed-tui      # interactive TUI
```

### Web frontend

```bash
cd ts
pnpm install
pnpm dev
```

### Code generation

Both the Go SDK and TS client are generated from the OpenAPI spec:

```bash
make gen             # from repo root — regenerates spec + both clients
```
