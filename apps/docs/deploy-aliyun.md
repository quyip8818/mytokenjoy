# 阿里云部署方案（方案 A：ECS + Docker Compose + ACR）

所有服务容器化，CI 构建镜像推到阿里云 ACR，ECS 上只需 `docker compose pull && up -d`。

| 域名 | 服务 | 镜像 |
|------|------|------|
| www.tokenjoy.me | web 静态官网 | `tokenjoy-web` |
| app.tokenjoy.me | apps 前端 + 后端 + newapi | `tokenjoy-apps-frontend` / `tokenjoy-apps-backend` / `tokenjoy-newapi` |
| sms.tokenjoy.me | sms 前端 + 后端 + newapi-sms | `tokenjoy-sms-frontend` / `tokenjoy-sms-backend` / `calciumion/new-api` |

## 架构图

```
GitHub Push (main)
  │
  ▼
GitHub Actions CI
  ├─ 构建 6 个镜像（多阶段 Dockerfile）
  └─ 推送到阿里云 ACR
  │
  ▼
ECS (docker compose pull && up -d)
  │
  ▼
Nginx (容器, 80/443)
  ├─ www.tokenjoy.me  → tokenjoy-web:80
  ├─ app.tokenjoy.me  → tokenjoy-apps-frontend:80 (/api → apps-backend:8010)
  └─ sms.tokenjoy.me  → tokenjoy-sms-frontend:80 (/api → sms-backend:8020)

内部网络:
  - postgres:17 (或阿里云 RDS)
  - redis:7
  - newapi-apps:3000
  - newapi-sms:3000
  - apps-backend:8010
  - sms-backend:8020
```

## 前置条件

- 阿里云 ECS（2C4G+），安装 Docker + Docker Compose
- 阿里云 ACR 个人版实例（免费，3 个命名空间）
- 域名解析：三条 A 记录 → ECS 公网 IP
- SSL 证书（阿里云免费证书或 Let's Encrypt）

## 服务器目录结构

```
/opt/tokenjoy/
├── docker-compose.prod.yml
├── nginx.conf
├── ssl/                  # 证书
├── .env                  # 生产环境变量
└── postgres-init/        # 数据库初始化脚本
```

极简——服务器上没有源码、没有 Node/Go 环境，只有配置文件和 Docker。

---

## 1. Dockerfile 清单

### apps/frontend/Dockerfile

```dockerfile
FROM node:24-alpine AS builder
RUN corepack enable pnpm
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/frontend/package.json apps/frontend/
COPY packages/ packages/
RUN pnpm install --frozen-lockfile --filter @tokenjoy/frontend...
COPY apps/frontend/ apps/frontend/
RUN pnpm -F @tokenjoy/frontend build

FROM nginx:alpine
COPY --from=builder /app/apps/frontend/dist /usr/share/nginx/html
COPY apps/frontend/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

### apps/frontend/nginx.conf (容器内)

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location /api/ {
        proxy_pass http://apps-backend:8010;
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
```

### apps/backend/Dockerfile

```dockerfile
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

### sms/frontend/Dockerfile

```dockerfile
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
COPY sms/frontend/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

### sms/frontend/nginx.conf (容器内)

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
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
```

### sms/backend/Dockerfile

```dockerfile
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

### web/Dockerfile

```dockerfile
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
RUN echo 'server { listen 80; root /usr/share/nginx/html; index index.html; location / { try_files $uri $uri/ /index.html; } location ~* \.(js|css|png|jpg|svg|woff2?)$ { expires 30d; add_header Cache-Control "public, immutable"; } }' > /etc/nginx/conf.d/default.conf
EXPOSE 80
```

### apps/newapi/Dockerfile

已有，不需要修改。

---

## 2. 生产 Docker Compose

```yaml
# docker-compose.prod.yml
# 放在 ECS /opt/tokenjoy/ 下

x-restart: &restart
  restart: unless-stopped

services:
  postgres:
    image: postgres:17-alpine
    <<: *restart
    environment:
      POSTGRES_USER: ${PG_USER}
      POSTGRES_PASSWORD: ${PG_PASSWORD}
      POSTGRES_DB: tokenjoy
    volumes:
      - pg_data:/var/lib/postgresql/data
      - ./postgres-init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ['CMD-SHELL', 'pg_isready -U ${PG_USER}']
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    <<: *restart
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ['CMD', 'redis-cli', '-a', '${REDIS_PASSWORD}', 'ping']
      interval: 10s
      timeout: 5s
      retries: 5

  newapi-apps:
    image: ${ACR_REGISTRY}/tokenjoy-newapi:${IMAGE_TAG:-latest}
    <<: *restart
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
    <<: *restart
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
    image: ${ACR_REGISTRY}/tokenjoy-apps-backend:${IMAGE_TAG:-latest}
    <<: *restart
    depends_on:
      postgres: { condition: service_healthy }
      redis: { condition: service_healthy }
      newapi-apps: { condition: service_started }
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
    image: ${ACR_REGISTRY}/tokenjoy-sms-backend:${IMAGE_TAG:-latest}
    <<: *restart
    depends_on:
      postgres: { condition: service_healthy }
    environment:
      PORT: '8020'
      DATABASE_URL: postgresql://${PG_USER}:${PG_PASSWORD}@postgres:5432/sms?sslmode=disable
      SESSION_SECRET: ${SMS_SESSION_SECRET}

  apps-frontend:
    image: ${ACR_REGISTRY}/tokenjoy-apps-frontend:${IMAGE_TAG:-latest}
    <<: *restart
    depends_on: [apps-backend]

  sms-frontend:
    image: ${ACR_REGISTRY}/tokenjoy-sms-frontend:${IMAGE_TAG:-latest}
    <<: *restart
    depends_on: [sms-backend]

  web:
    image: ${ACR_REGISTRY}/tokenjoy-web:${IMAGE_TAG:-latest}
    <<: *restart

  nginx:
    image: nginx:alpine
    <<: *restart
    ports:
      - '80:80'
      - '443:443'
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - apps-frontend
      - sms-frontend
      - web

volumes:
  pg_data:
  redis_data:
```

---

## 3. Nginx 入口配置

```nginx
# nginx.conf — 放在 ECS /opt/tokenjoy/nginx.conf
worker_processes auto;
events { worker_connections 1024; }

http {
    include       mime.types;
    default_type  application/octet-stream;
    sendfile      on;
    gzip          on;
    gzip_types    text/plain text/css application/json application/javascript text/xml;

    # www.tokenjoy.me
    server {
        listen 443 ssl;
        server_name www.tokenjoy.me;
        ssl_certificate     /etc/nginx/ssl/www.tokenjoy.me.pem;
        ssl_certificate_key /etc/nginx/ssl/www.tokenjoy.me.key;

        location / {
            proxy_pass http://web:80;
            proxy_set_header Host $host;
        }
    }

    # app.tokenjoy.me
    server {
        listen 443 ssl;
        server_name app.tokenjoy.me;
        ssl_certificate     /etc/nginx/ssl/app.tokenjoy.me.pem;
        ssl_certificate_key /etc/nginx/ssl/app.tokenjoy.me.key;

        location / {
            proxy_pass http://apps-frontend:80;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # sms.tokenjoy.me
    server {
        listen 443 ssl;
        server_name sms.tokenjoy.me;
        ssl_certificate     /etc/nginx/ssl/sms.tokenjoy.me.pem;
        ssl_certificate_key /etc/nginx/ssl/sms.tokenjoy.me.key;

        location / {
            proxy_pass http://sms-frontend:80;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # HTTP → HTTPS
    server {
        listen 80;
        server_name www.tokenjoy.me app.tokenjoy.me sms.tokenjoy.me;
        return 301 https://$host$request_uri;
    }
}
```

注意：入口 Nginx 只做 SSL 终止 + 域名路由，API 反代由各前端容器内的 nginx 完成。

---

## 4. 环境变量

```bash
# .env（放在 ECS /opt/tokenjoy/.env）
ACR_REGISTRY=registry.cn-hangzhou.aliyuncs.com/tokenjoy
IMAGE_TAG=latest

PG_USER=tokenjoy
PG_PASSWORD=<强密码>
REDIS_PASSWORD=<强密码>

SESSION_SECRET=<随机32字符>
SMS_SESSION_SECRET=<随机32字符>
NEWAPI_SESSION_SECRET=<随机32字符>
SMS_NEWAPI_SESSION_SECRET=<随机32字符>
NEWAPI_ADMIN_TOKEN=<newapi管理token>
WEBHOOK_SECRET=<webhook签名密钥>
SUPPORT_SAAS=false
```

---

## 5. GitHub Actions CI/CD

```yaml
# .github/workflows/deploy.yml
name: Build & Deploy

on:
  push:
    branches: [main]

env:
  ACR_REGISTRY: registry.cn-hangzhou.aliyuncs.com/tokenjoy
  IMAGE_TAG: ${{ github.sha }}

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5

      - name: Login to ACR
        run: echo "${{ secrets.ACR_PASSWORD }}" | docker login $ACR_REGISTRY -u ${{ secrets.ACR_USERNAME }} --password-stdin

      - name: Build & push images
        run: |
          # apps-frontend
          docker build -f apps/frontend/Dockerfile -t $ACR_REGISTRY/tokenjoy-apps-frontend:$IMAGE_TAG .
          docker push $ACR_REGISTRY/tokenjoy-apps-frontend:$IMAGE_TAG
          docker tag $ACR_REGISTRY/tokenjoy-apps-frontend:$IMAGE_TAG $ACR_REGISTRY/tokenjoy-apps-frontend:latest
          docker push $ACR_REGISTRY/tokenjoy-apps-frontend:latest

          # apps-backend
          docker build -f apps/backend/Dockerfile -t $ACR_REGISTRY/tokenjoy-apps-backend:$IMAGE_TAG apps/backend
          docker push $ACR_REGISTRY/tokenjoy-apps-backend:$IMAGE_TAG
          docker tag $ACR_REGISTRY/tokenjoy-apps-backend:$IMAGE_TAG $ACR_REGISTRY/tokenjoy-apps-backend:latest
          docker push $ACR_REGISTRY/tokenjoy-apps-backend:latest

          # newapi (apps fork)
          docker build -f apps/newapi/Dockerfile -t $ACR_REGISTRY/tokenjoy-newapi:$IMAGE_TAG apps/newapi
          docker push $ACR_REGISTRY/tokenjoy-newapi:$IMAGE_TAG
          docker tag $ACR_REGISTRY/tokenjoy-newapi:$IMAGE_TAG $ACR_REGISTRY/tokenjoy-newapi:latest
          docker push $ACR_REGISTRY/tokenjoy-newapi:latest

          # sms-frontend
          docker build -f sms/frontend/Dockerfile -t $ACR_REGISTRY/tokenjoy-sms-frontend:$IMAGE_TAG .
          docker push $ACR_REGISTRY/tokenjoy-sms-frontend:$IMAGE_TAG
          docker tag $ACR_REGISTRY/tokenjoy-sms-frontend:$IMAGE_TAG $ACR_REGISTRY/tokenjoy-sms-frontend:latest
          docker push $ACR_REGISTRY/tokenjoy-sms-frontend:latest

          # sms-backend
          docker build -f sms/backend/Dockerfile -t $ACR_REGISTRY/tokenjoy-sms-backend:$IMAGE_TAG sms/backend
          docker push $ACR_REGISTRY/tokenjoy-sms-backend:$IMAGE_TAG
          docker tag $ACR_REGISTRY/tokenjoy-sms-backend:$IMAGE_TAG $ACR_REGISTRY/tokenjoy-sms-backend:latest
          docker push $ACR_REGISTRY/tokenjoy-sms-backend:latest

          # web
          docker build -f web/Dockerfile -t $ACR_REGISTRY/tokenjoy-web:$IMAGE_TAG .
          docker push $ACR_REGISTRY/tokenjoy-web:$IMAGE_TAG
          docker tag $ACR_REGISTRY/tokenjoy-web:$IMAGE_TAG $ACR_REGISTRY/tokenjoy-web:latest
          docker push $ACR_REGISTRY/tokenjoy-web:latest

  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to ECS
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: root
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          script: |
            cd /opt/tokenjoy
            echo "${{ secrets.ACR_PASSWORD }}" | docker login ${{ secrets.ACR_REGISTRY }} -u ${{ secrets.ACR_USERNAME }} --password-stdin
            docker compose -f docker-compose.prod.yml pull
            docker compose -f docker-compose.prod.yml up -d --remove-orphans
            docker image prune -f
```

---

## 6. 手动部署脚本（备用）

```bash
#!/usr/bin/env bash
# deploy.sh — 在本地手动触发部署（CI 挂了时用）
set -euo pipefail

SERVER="root@你的ECS公网IP"
SSH_KEY="${SSH_KEY:-~/.ssh/id_rsa}"

ssh -i "${SSH_KEY}" "${SERVER}" << 'EOF'
cd /opt/tokenjoy
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d --remove-orphans
docker image prune -f
echo "✓ 部署完成"
EOF
```

就这么短。所有构建逻辑都在 CI 里完成了。

---

## 7. 首次初始化

```bash
# 1. ECS 安装 Docker
curl -fsSL https://get.docker.com | sh
systemctl enable --now docker

# 2. 创建部署目录
mkdir -p /opt/tokenjoy/ssl

# 3. 上传配置文件（本地执行）
scp docker-compose.prod.yml nginx.conf .env root@ECS_IP:/opt/tokenjoy/
scp -r scripts/postgres-init root@ECS_IP:/opt/tokenjoy/postgres-init/
scp ssl/*.pem ssl/*.key root@ECS_IP:/opt/tokenjoy/ssl/

# 4. 登录 ACR
docker login registry.cn-hangzhou.aliyuncs.com

# 5. 拉起服务
cd /opt/tokenjoy
docker compose -f docker-compose.prod.yml up -d
```

---

## 8. 日常操作

```bash
# 查看状态
docker compose -f docker-compose.prod.yml ps

# 查看日志
docker compose -f docker-compose.prod.yml logs -f apps-backend

# 回滚到上一个版本
# 修改 .env 中 IMAGE_TAG=<旧的 git sha>
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d

# 数据库备份
docker compose -f docker-compose.prod.yml exec postgres pg_dump -U tokenjoy tokenjoy > backup.sql
```

---

## 9. 升级路径

| 触发条件 | 操作 |
|----------|------|
| 内存紧张 | PG/Redis 外移到阿里云 RDS/Redis，删 compose 里的 postgres/redis 服务 |
| 需要自动扩缩 | 迁移到 SAE，镜像不变 |
| 需要零停机部署 | 前面加 ALB + 两台 ECS 蓝绿切换 |
| 想更懒 | 加 Watchtower 容器，自动拉最新镜像 |
