# sms-frontend: pnpm build → nginx serve + API 反代
# Context: repo root
FROM node:24-alpine AS builder
RUN corepack enable pnpm
ENV npm_config_registry=https://registry.npmmirror.com
WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY sms/frontend/package.json sms/frontend/
COPY packages/ packages/
RUN pnpm install --frozen-lockfile --filter @sms/frontend...

COPY sms/frontend/ sms/frontend/
RUN pnpm -F @sms/frontend build

FROM nginx:alpine
COPY --from=builder /app/sms/frontend/dist /usr/share/nginx/html
COPY deploy/dockerfiles/spa-proxy.conf /etc/nginx/templates/default.conf.template
ENV BACKEND_HOST=sms-backend:8020
EXPOSE 80
