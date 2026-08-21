# syntax=docker/dockerfile:1

# ---- deps stage ----
FROM node:22-alpine AS deps
RUN corepack enable
WORKDIR /app
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY apps/website/package.json ./apps/website/package.json
COPY packages/typescript-config/package.json ./packages/typescript-config/package.json
COPY packages/ui/package.json ./packages/ui/package.json
RUN pnpm install --frozen-lockfile

# ---- builder stage ----
FROM node:22-alpine AS builder
RUN corepack enable
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/apps/website/node_modules ./apps/website/node_modules
COPY --from=deps /app/packages/ui/node_modules ./packages/ui/node_modules
COPY ./ ./
WORKDIR /app/apps/website

RUN pnpm run build

# ---- runtime stage ----
# The website is the front door for sunred.app: an Astro SSR server that
# serves its own pages and proxies every other path to the Next.js app
# (APP_URL, e.g. http://web:3000). See src/middleware.ts and
# src/pages/[...path].ts.
FROM node:22-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
# Next.js app origin to proxy unknown paths to. Override in compose.
ENV APP_URL=http://web:3000
ENV HOST=0.0.0.0
ENV PORT=80

# Astro's @astrojs/node standalone build outputs the server at
# dist/server/entry.mjs and serves static client assets from dist/client/.
COPY --from=builder /app/apps/website/dist ./

EXPOSE 80

CMD ["node", "server/entry.mjs"]
