# syntax=docker/dockerfile:1.4

FROM node:22-bookworm-slim AS base

# Prune stage: This breaks the app into "json" (dependencies) and "full" (source)
FROM base AS prune
WORKDIR /app
RUN npm install -g turbo
COPY . .
RUN turbo prune naurffxiv --docker

# Installer stage: Install dependencies using the pruned lockfile
FROM base AS installer
WORKDIR /app
# Only install build tools if you have native dependencies
RUN apt-get update && \
    apt-get install -y build-essential python3 && \
    rm -rf /var/lib/apt/lists/*
COPY .gitignore .gitignore
COPY --from=prune /app/out/json/ .
COPY --from=prune /app/out/package-lock.json ./package-lock.json
# Remove prepare script to prevent husky/git hooks from running
RUN npm pkg delete scripts.prepare
# Install dependencies (only the ones needed for this workspace)
RUN npm ci

# Dev stage: for local development with hot reloading
FROM base AS dev
WORKDIR /app
COPY --from=installer /app/ .
COPY package.json turbo.json ./
COPY services/naurffxiv services/naurffxiv
CMD ["npm", "run", "-w", "naurffxiv", "dev"]

# Builder stage: Build the application
FROM base AS builder
WORKDIR /app
COPY --from=installer /app/ .
COPY package.json turbo.json ./
COPY services/naurffxiv services/naurffxiv

# Build with cache mount for Next.js
RUN --mount=type=cache,target=/app/.next/cache \
    npm run -w naurffxiv build

# Runner stage
FROM base AS runner
WORKDIR /app

# Install curl for healthchecks and clean up in same layer
RUN apt-get update && \
    apt-get install -y curl && \
    rm -rf /var/lib/apt/lists/*

ENV NODE_ENV=production

RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Copy public assets
COPY --from=builder --chown=nextjs:nodejs /app/services/naurffxiv/public ./services/naurffxiv/public
COPY --from=builder --chown=nextjs:nodejs /app/services/naurffxiv/public ./public

# Set the correct permission for prerender cache
RUN mkdir .next && \
    chown nextjs:nodejs .next

# Copy standalone output (Next.js bundles everything needed)
COPY --from=builder --chown=nextjs:nodejs /app/services/naurffxiv/.next/standalone ./
# Copy static files to correct location within standalone structure
COPY --from=builder --chown=nextjs:nodejs /app/services/naurffxiv/.next/static ./services/naurffxiv/.next/static

USER nextjs

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s \
  CMD curl -f http://localhost:3000/api/health || exit 1

# Start the Next.js server
CMD ["node", "services/naurffxiv/server.js"]
