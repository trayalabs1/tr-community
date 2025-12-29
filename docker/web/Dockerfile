#
# FRONTEND WEB LAYER
#

FROM docker.io/node:22.21.1-alpine AS base

#
# DEPENDENCIES LAYER
#

FROM base AS deps

RUN apk add --no-cache libc6-compat
WORKDIR /storyden

COPY ./web/package.json ./web/yarn.lock* ./web/.yarnrc.yml ./
RUN --mount=type=cache,target=/root/.yarn \
    corepack enable && \
    YARN_CACHE_FOLDER=/root/.yarn yarn install --immutable --network-timeout 1000000

#
# BUILDER LAYER
#

FROM base AS builder
WORKDIR /storyden
COPY --from=deps /storyden/node_modules ./node_modules
COPY ./web .

ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_BUILD_STANDALONE=true

# NOTE: A regular Next.js build will attempt to do static generation for every
# page. But we don't want to constantly be adding force-dynamic to every page so
# this ensures no SSG runs during the Docker build.
RUN --mount=type=cache,target=/storyden/.next/cache \
    corepack enable && yarn next build --experimental-build-mode=compile

#
# RUNTIME IMAGE
#

FROM base AS runner
WORKDIR /storyden

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 storyden && \
    adduser --system --uid 1001 storyden && \
    chown storyden:storyden /storyden

USER storyden

COPY --from=builder /storyden/public ./public

RUN mkdir .next

COPY --from=builder --chown=storyden:storyden /storyden/.next/standalone ./
COPY --from=builder --chown=storyden:storyden /storyden/.next/static ./.next/static

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

# Frontend connects directly to the API server
# Change these via -e when deploying
ENV PUBLIC_API_ADDRESS="http://localhost:8000"
ENV PUBLIC_WEB_ADDRESS="http://localhost:3000"

# SSR requests go directly to the API server
ENV SSR_API_ADDRESS="http://localhost:8000"

ENTRYPOINT ["node", "server.js"]
