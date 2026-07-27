# 阿里云部署方案

三个平台部署到一台 ECS，Nginx 反代，Docker Compose 编排所有服务。

| 域名 | 服务 | 内容 |
|------|------|------|
| www.tokenjoy.me | web | 静态官网（Vite 构建产物） |
| app.tokenjoy.me | apps frontend + backend + newapi | 客户侧平台 |
| sms.tokenjoy.me | sms frontend + backend + newapi-sms | 内部供应商管理 |

## 前置条件

- 阿里云 ECS（建议 2C4G+），装 Docker + Docker Compose
- 域名解析：三条 A 记录指向 ECS 公网 IP
- SSL 证书（可用阿里云免费证书或 Let's Encrypt）
- 阿里云容器镜像服务 ACR（可选，也可用 GitHub Container Registry）

## 架构

```
Internet
  │
  ▼
Nginx (80/443)
  ├─ www.tokenjoy.me  → 静态文件 /srv/web/dist
  ├─ app.tokenjoy.me  → apps-frontend:80 (静态) + /api → apps-backend:8010
  └─ sms.tokenjoy.me  → sms-frontend:80 (静态) + /api → sms-backend:8020

Docker Compose (内部网络):
  - postgres:17
  - redis:7
  - newapi-apps (端口 3010)
  - newapi-sms (端口 3020)
  - apps-backend (端口 8010)
  - sms-backend (端口 8020)
```

## 目录结构（服务器侧）

```
/opt/tokenjoy/
├── docker-compose.prod.yml
├── nginx/
│   ├── nginx.conf
│   └── ssl/              # 证书文件
├── web-dist/             # www.tokenjoy.me 静态资源
├── apps-frontend-dist/   # app.tokenjoy.me 静态资源
├── sms-frontend-dist/    # sms.tokenjoy.me 静态资源
├── .env                  # 生产环境变量
└── deploy.sh            # 一键部署脚本
```

## 1. 生产 Docker Compose

```yaml
# docker-compose.prod.yml
services:
  postgres:
    image: postgres:17-alpine
    restart: always
    environment:
      POSTGRES_USER: ${PG_USER}
      POSTGRES_PASSWORD: ${PG_PASSWORD}
      POSTGRES_DB: tokenjoy
    volumes:
      - pg_data:/var/lib/postgresql/data
      - ./scripts/postgres-init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ['CMD-SHELL', 'pg_isready -U ${PG_USER}']
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    restart: always
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ['CMD', 'redis-cli', '-a', '${REDIS_PASSWORD}', 'ping']
      interval: 10s
      timeout: 5s
      retries: 5

  newapi-apps:
    build:
      context: ./apps/newapi
      dockerfile: Dockerfile
    restart: always
    depends_on:
      postgres: { condition: service_healthy }
      redis: { condition: service_healthy }
    environment:
      SQL_DSN: postgresql://${PG_USER}:${PG_PASSWORD}@postgres:5432/newapi?sslmode=disable
      LOG_SQL_DSN: postgresql://${PG_USER}:${PG_PASSWORD}@postgres:5432/logs?sslmode=disable&search_path=newapi
      REDIS_CONN_STRING: redis://:${REDIS_PASSWORD}@redis:6379/0
      SESSION_SECRET: ${NEWAPI_SESSION_SECRET}
      SYNC_FREQUENCY: 60
      MANAGEMENT_WEBHOOK_URL: http://apps-backend:8010/api/internal/webhooks/newapi-log
      MANAGEMENT_WEBHOOK_SECRET: ${WEBHOOK_SECRET}

  newapi-sms:
    image: calciumion/new-api:latest
    restart: always
    depends_on:
      postgres: { condition: service_healthy }
      redis: { condition: service_healthy }
    environment:
      SQL_DSN: postgresql://${PG_USER}:${PG_PASSWORD}@postgres:5432/sms_newapi?sslmode=disable
      LOG_SQL_DSN: postgresql://${PG_USER}:${PG_PASSWORD}@postgres:5432/sms_logs?sslmode=disable&search_path=newapi
      REDIS_CONN_STRING: redis://:${REDIS_PASSWORD}@redis:6379/1
      SESSION_SECRET: ${SMS_NEWAPI_SESSION_SECRET}
      SYNC_FREQUENCY: 60

  apps-backend:
    build:
      context: ./apps/backend
      dockerfile: Dockerfile
    restart: always
    depends_on:
      postgres: { condition: service_healthy }
      redis: { condition: service_healthy }
      newapi-apps: { condition: service_started }
    ports:
      - '127.0.0.1:8010:8010'
    environment:
      PORT: '8010'
      DATABASE_URL: postgresql://${PG_USER}:${PG_PASSWORD}@postgres:5432/tokenjoy?sslmode=disable
      REDIS_URL: redis://:${REDIS_PASSWORD}@redis:6379/2
      SESSION_SECRET: ${SESSION_SECRET}
      DEPLOY_ENV: production
      SUPPORT_SAAS: ${SUPPORT_SAAS:-false}
      NEWAPI_BASE_URL: http://newapi-apps:3000
      NEWAPI_ADMIN_TOKEN: ${NEWAPI_ADMIN_TOKEN}
      WEBHOOK_SECRET: ${WEBHOOK_SECRET}

  sms-backend:
    build:
      context: ./sms/backend
      dockerfile: Dockerfile
    restart: always
    depends_on:
      postgres: { condition: service_healthy }
      redis: { condition: service_healthy }
      newapi-sms: { condition: service_started }
    ports:
      - '127.0.0.1:8020:8020'
    environment:
      PORT: '8020'
      DATABASE_URL: postgresql://${PG_USER}:${PG_PASSWORD}@postgres:5432/sms?sslmode=disable
      SESSION_SECRET: ${SMS_SESSION_SECRET}

  nginx:
    image: nginx:alpine
    restart: always
    ports:
      - '80:80'
      - '443:443'
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
      - ./web-dist:/srv/web:ro
      - ./apps-frontend-dist:/srv/apps-frontend:ro
      - ./sms-frontend-dist:/srv/sms-frontend:ro
    depends_on:
      - apps-backend
      - sms-backend

volumes:
  pg_data:
  redis_data:
```

## 2. Dockerfile — apps/backend

```dockerfile
# apps/backend/Dockerfile
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /build/server /usr/local/bin/server
EXPOSE 8010
ENTRYPOINT ["server"]
```

## 3. Dockerfile — sms/backend

```dockerfile
# sms/backend/Dockerfile
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /build/server /usr/local/bin/server
EXPOSE 8020
ENTRYPOINT ["server"]
```

## 4. Nginx 配置

```nginx
# nginx/nginx.conf
worker_processes auto;
events { worker_connections 1024; }

http {
    include       mime.types;
    default_type  application/octet-stream;
    sendfile      on;
    gzip          on;
    gzip_types    text/plain text/css application/json application/javascript text/xml;

    # --- www.tokenjoy.me (静态官网) ---
    server {
        listen 443 ssl;
        server_name www.tokenjoy.me;

        ssl_certificate     /etc/nginx/ssl/www.tokenjoy.me.pem;
        ssl_certificate_key /etc/nginx/ssl/www.tokenjoy.me.key;

        root /srv/web;
        index index.html;

        location / {
            try_files $uri $uri/ /index.html;
        }

        location ~* \.(js|css|png|jpg|svg|woff2?)$ {
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }

    # --- app.tokenjoy.me (客户侧平台) ---
    server {
        listen 443 ssl;
        server_name app.tokenjoy.me;

        ssl_certificate     /etc/nginx/ssl/app.tokenjoy.me.pem;
        ssl_certificate_key /etc/nginx/ssl/app.tokenjoy.me.key;

        # 前端静态资源
        root /srv/apps-frontend;
        index index.html;

        # API 反代到 apps-backend
        location /api/ {
            proxy_pass http://apps-backend:8010;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # SPA 路由回退
        location / {
            try_files $uri $uri/ /index.html;
        }

        location ~* \.(js|css|png|jpg|svg|woff2?)$ {
            expires 7d;
            add_header Cache-Control "public, immutable";
        }
    }

    # --- sms.tokenjoy.me (内部管理) ---
    server {
        listen 443 ssl;
        server_name sms.tokenjoy.me;

        ssl_certificate     /etc/nginx/ssl/sms.tokenjoy.me.pem;
        ssl_certificate_key /etc/nginx/ssl/sms.tokenjoy.me.key;

        root /srv/sms-frontend;
        index index.html;

        location /api/ {
            proxy_pass http://sms-backend:8020;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        location / {
            try_files $uri $uri/ /index.html;
        }

        location ~* \.(js|css|png|jpg|svg|woff2?)$ {
            expires 7d;
            add_header Cache-Control "public, immutable";
        }
    }

    # HTTP → HTTPS 重定向
    server {
        listen 80;
        server_name www.tokenjoy.me app.tokenjoy.me sms.tokenjoy.me;
        return 301 https://$host$request_uri;
    }
}
```

## 5. 环境变量模板

```bash
# .env.example
PG_USER=tokenjoy
PG_PASSWORD=<强密码>
REDIS_PASSWORD=<强密码>
SESSION_SECRET=<随机 32 字符>
SMS_SESSION_SECRET=<随机 32 字符>
NEWAPI_SESSION_SECRET=<随机 32 字符>
SMS_NEWAPI_SESSION_SECRET=<随机 32 字符>
NEWAPI_ADMIN_TOKEN=<newapi 管理 token>
WEBHOOK_SECRET=<webhook 签名密钥>
SUPPORT_SAAS=false
```

## 6. 一键部署脚本

```bash
#!/usr/bin/env bash
# deploy.sh — 本地执行，构建并部署到阿里云 ECS
set -euo pipefail

# ========== 配置 ==========
SERVER_USER="root"
SERVER_HOST="你的ECS公网IP"
DEPLOY_DIR="/opt/tokenjoy"
SSH_KEY="${SSH_KEY:-~/.ssh/id_rsa}"
SSH="ssh -i ${SSH_KEY} ${SERVER_USER}@${SERVER_HOST}"
SCP="scp -i ${SSH_KEY}"

echo "=== 1. 安装依赖 ==="
pnpm install --frozen-lockfile

echo "=== 2. 构建前端 (web) ==="
pnpm -F @tokenjoy/web build

echo "=== 3. 构建前端 (apps) ==="
pnpm -F @tokenjoy/frontend build

echo "=== 4. 构建前端 (sms) ==="
pnpm -F @sms/frontend build

echo "=== 5. 打包前端产物 ==="
tar -czf /tmp/web-dist.tar.gz -C web/dist .
tar -czf /tmp/apps-frontend-dist.tar.gz -C apps/frontend/dist .
tar -czf /tmp/sms-frontend-dist.tar.gz -C sms/frontend/dist .

echo "=== 6. 同步后端源码到服务器 ==="
# 同步需要构建 Docker 镜像的文件
rsync -avz --delete \
  -e "ssh -i ${SSH_KEY}" \
  --include='apps/backend/***' \
  --include='apps/newapi/***' \
  --include='sms/backend/***' \
  --include='scripts/postgres-init/***' \
  --include='docker-compose.prod.yml' \
  --include='nginx/***' \
  --include='.env' \
  --exclude='*' \
  ./ "${SERVER_USER}@${SERVER_HOST}:${DEPLOY_DIR}/"

echo "=== 7. 上传前端产物 ==="
${SCP} /tmp/web-dist.tar.gz "${SERVER_USER}@${SERVER_HOST}:/tmp/"
${SCP} /tmp/apps-frontend-dist.tar.gz "${SERVER_USER}@${SERVER_HOST}:/tmp/"
${SCP} /tmp/sms-frontend-dist.tar.gz "${SERVER_USER}@${SERVER_HOST}:/tmp/"

${SSH} <<'REMOTE'
set -euo pipefail
DEPLOY_DIR="/opt/tokenjoy"

# 解压前端产物
rm -rf ${DEPLOY_DIR}/web-dist && mkdir -p ${DEPLOY_DIR}/web-dist
tar -xzf /tmp/web-dist.tar.gz -C ${DEPLOY_DIR}/web-dist

rm -rf ${DEPLOY_DIR}/apps-frontend-dist && mkdir -p ${DEPLOY_DIR}/apps-frontend-dist
tar -xzf /tmp/apps-frontend-dist.tar.gz -C ${DEPLOY_DIR}/apps-frontend-dist

rm -rf ${DEPLOY_DIR}/sms-frontend-dist && mkdir -p ${DEPLOY_DIR}/sms-frontend-dist
tar -xzf /tmp/sms-frontend-dist.tar.gz -C ${DEPLOY_DIR}/sms-frontend-dist

# 重建容器
cd ${DEPLOY_DIR}
docker compose -f docker-compose.prod.yml build --parallel
docker compose -f docker-compose.prod.yml up -d --remove-orphans

# 等待健康检查
echo "等待服务启动..."
sleep 10
docker compose -f docker-compose.prod.yml ps

# 清理
rm -f /tmp/web-dist.tar.gz /tmp/apps-frontend-dist.tar.gz /tmp/sms-frontend-dist.tar.gz
REMOTE

echo "=== 部署完成 ==="
echo "  www.tokenjoy.me → web 官网"
echo "  app.tokenjoy.me → apps 客户平台"
echo "  sms.tokenjoy.me → sms 内部管理"
```

## 7. 首次服务器初始化

首次部署前在 ECS 上执行：

```bash
# 安装 Docker
curl -fsSL https://get.docker.com | sh
systemctl enable --now docker

# 创建部署目录
mkdir -p /opt/tokenjoy/nginx/ssl

# 上传 .env（从 .env.example 改好密码后上传）
# 上传 SSL 证书到 /opt/tokenjoy/nginx/ssl/
```

## 8. 后续自动化（可选）

把 `deploy.sh` 接入 GitHub Actions：

```yaml
# .github/workflows/deploy.yml
name: Deploy
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v5
        with:
          node-version: 24
          cache: pnpm
      - run: pnpm install --frozen-lockfile
      - run: pnpm -F @tokenjoy/web build
      - run: pnpm -F @tokenjoy/frontend build
      - run: pnpm -F @sms/frontend build
      - name: Deploy to server
        env:
          SSH_PRIVATE_KEY: ${{ secrets.SSH_PRIVATE_KEY }}
          SERVER_HOST: ${{ secrets.SERVER_HOST }}
        run: |
          mkdir -p ~/.ssh
          echo "${SSH_PRIVATE_KEY}" > ~/.ssh/deploy_key
          chmod 600 ~/.ssh/deploy_key
          export SSH_KEY=~/.ssh/deploy_key
          export SERVER_HOST
          bash deploy.sh
```

## 注意事项

1. **数据库备份**：生产环境请配置 pg_dump 定期备份到 OSS
2. **SSL 续期**：如用 Let's Encrypt，建议装 certbot + cron 自动续期
3. **防火墙**：ECS 安全组只开放 80/443，SSH 限制 IP
4. **日志**：docker compose logs 查看，建议对接阿里云 SLS
5. **监控**：接入阿里云云监控，设置 CPU/内存/磁盘告警
