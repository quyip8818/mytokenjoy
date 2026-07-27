# web: Static website (www.tokenjoy.me)
# Build context: repo root
FROM node:24-alpine AS builder
RUN corepack enable pnpm
WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY web/package.json web/
COPY packages/ packages/

RUN pnpm install --frozen-lockfile --filter @tokenjoy/web...

COPY web/ web/
RUN pnpm -F @tokenjoy/web build

FROM nginx:alpine
COPY --from=builder /app/web/dist /usr/share/nginx/html
# ponytail: 内联 nginx 配置，静态站点不需要单独文件
RUN printf 'server {\n\
    listen 80;\n\
    root /usr/share/nginx/html;\n\
    index index.html;\n\
    location / { try_files $uri $uri/ /index.html; }\n\
    location ~* \\.(js|css|png|jpg|svg|woff2?)$ {\n\
        expires 30d;\n\
        add_header Cache-Control "public, immutable";\n\
    }\n\
}\n' > /etc/nginx/conf.d/default.conf
EXPOSE 80
