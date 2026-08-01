# web: 官网纯静态
# Context: repo root
FROM node:24-alpine AS builder
RUN corepack enable pnpm
ENV npm_config_registry=https://registry.npmmirror.com
WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY web/package.json web/
COPY packages/ packages/
RUN pnpm install --frozen-lockfile --filter @tokenjoy/web...

COPY web/ web/
RUN pnpm -F @tokenjoy/web build

FROM nginx:alpine
COPY --from=builder /app/web/dist /usr/share/nginx/html
COPY deploy/dockerfiles/spa-static.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
