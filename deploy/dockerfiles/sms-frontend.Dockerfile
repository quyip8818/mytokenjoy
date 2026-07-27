# sms-frontend: Vite build → nginx serve + API reverse proxy
# Build context: repo root
FROM node:24-alpine AS builder
RUN corepack enable pnpm
WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY sms/frontend/package.json sms/frontend/
COPY packages/ packages/

RUN pnpm install --frozen-lockfile --filter @sms/frontend...

COPY sms/frontend/ sms/frontend/
RUN pnpm -F @sms/frontend build

FROM nginx:alpine
COPY --from=builder /app/sms/frontend/dist /usr/share/nginx/html
COPY deploy/nginx/sms-frontend.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
