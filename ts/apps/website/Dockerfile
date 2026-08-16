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
FROM nginx:alpine AS runner
WORKDIR /usr/share/nginx/html

COPY --from=builder /app/apps/website/dist ./

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
