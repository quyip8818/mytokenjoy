# apps-frontend: pnpm build → nginx serve + API 反代
# Context: repo root
FROM node:24-alpine AS builder
RUN corepack enable pnpm
ENV npm_config_registry=https://registry.npmmirror.com
WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/frontend/package.json apps/frontend/
COPY packages/ packages/
RUN pnpm install --frozen-lockfile --filter @tokenjoy/frontend...

COPY apps/frontend/ apps/frontend/
ENV VITE_SUPPORT_SAAS=true
RUN pnpm -F @tokenjoy/frontend build

FROM nginx:alpine
COPY --from=builder /app/apps/frontend/dist /usr/share/nginx/html
COPY deploy/dockerfiles/spa-proxy.conf /etc/nginx/templates/default.conf.template
ENV BACKEND_HOST=apps-backend:8010
EXPOSE 80
