# syntax=docker/dockerfile:1

# ---- deps stage ----
FROM node:22-alpine AS deps
RUN corepack enable
WORKDIR /app
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY apps/web/package.json ./apps/web/package.json
COPY packages/api-client/package.json ./packages/api-client/package.json
COPY packages/typescript-config/package.json ./packages/typescript-config/package.json
COPY packages/eslint-config/package.json ./packages/eslint-config/package.json
COPY packages/ui/package.json ./packages/ui/package.json
RUN pnpm install --frozen-lockfile

# ---- builder stage ----
FROM node:22-alpine AS builder
RUN corepack enable
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/apps/web/node_modules ./apps/web/node_modules
COPY --from=deps /app/packages/api-client/node_modules ./packages/api-client/node_modules
COPY --from=deps /app/packages/ui/node_modules ./packages/ui/node_modules
COPY ./ ./
WORKDIR /app/apps/web

RUN pnpm run build

# ---- runtime stage ----
FROM node:22-alpine AS runner
WORKDIR /app/apps/web

ENV NODE_ENV=production

RUN addgroup -g 10001 -S nodejs \
 && adduser -S nextjs -u 10001 -G nodejs

COPY --from=builder --chown=nextjs:nodejs /app/apps/web/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/apps/web/.next/static ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/apps/web/public ./public

USER nextjs
EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "apps/web/server.js"]
