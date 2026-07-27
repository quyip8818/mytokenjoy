# apps-frontend: Vite build → nginx serve + API reverse proxy
# Build context: repo root
FROM node:24-alpine AS builder
RUN corepack enable pnpm
WORKDIR /app

# 1) 只复制 lockfile + workspace 声明，利用缓存
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/frontend/package.json apps/frontend/
COPY packages/ packages/

RUN pnpm install --frozen-lockfile --filter @tokenjoy/frontend...

# 2) 复制源码并构建
COPY apps/frontend/ apps/frontend/
RUN pnpm -F @tokenjoy/frontend build

# 3) 生产镜像
FROM nginx:alpine
COPY --from=builder /app/apps/frontend/dist /usr/share/nginx/html
COPY deploy/nginx/apps-frontend.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
