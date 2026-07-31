# 部署

全 Docker Compose。ECS 上只装 Docker + Git。

## 架构

```
ECS
└─ docker compose
    ├─ nginx         :80/:443  SSL + 域名路由
    ├─ apps-frontend           SPA + /api,/v1 反代 → apps-backend
    ├─ apps-backend  :8010     Go API + LLM Gateway
    ├─ sms-frontend            SPA + /api 反代 → sms-backend
    ├─ sms-backend   :8020     Go API
    ├─ web                     静态官网
    ├─ newapi        :3000     LLM 转发
    ├─ postgres      :5432
    └─ redis         :6379
```

## 文件

```
deploy/
├── init.sh                首次部署（在 ECS 上执行）
├── deploy.sh              日常部署（从本地 SSH 到 ECS）
├── docker-compose.yml     全服务编排
├── nginx.conf             入口网关 (SSL 终止)
├── init-db.sh             PG 首次初始化脚本
├── dockerfiles/
│   ├── apps-backend.Dockerfile
│   ├── apps-frontend.Dockerfile
│   ├── sms-backend.Dockerfile
│   ├── sms-frontend.Dockerfile
│   ├── web.Dockerfile
│   ├── spa-proxy.conf     前端 nginx: SPA + API 反代
│   └── spa-static.conf    前端 nginx: 纯静态
├── env/
│   ├── infra.env.example  基础设施密码 (PG/Redis/NewAPI)
│   ├── apps.env.example   apps-backend 配置
│   └── sms.env.example    sms-backend 配置
└── ssl/                   SSL 证书 (不入 git)
```

## 首次部署

```bash
# SSH 到 ECS
ssh root@ECS_IP

# 装 git（如果没有）
apt-get update && apt-get install -y git

# clone 仓库
git clone <REPO_URL> /opt/tokenjoy/src
cd /opt/tokenjoy/src

# 上传 SSL 证书到 deploy/ssl/（从本地 scp，或手动放）
# scp deploy/ssl/*.pem deploy/ssl/*.key root@ECS:/opt/tokenjoy/src/deploy/ssl/

# 一键初始化（装 Docker + 生成密码 + 构建启动 + 初始化 NewAPI）
./deploy/init.sh
```

脚本自动完成:
1. 检查/安装 Docker
2. 生成随机密码写入 env 文件
3. 检查 SSL 证书
4. `docker compose up -d --build` 启动全栈
5. 初始化 NewAPI admin 并设置 access_token
6. 打印登录信息

## 日常部署

```bash
./deploy/deploy.sh           # 用 DEPLOY_HOST 环境变量
./deploy/deploy.sh 47.x.x.x # 或直接传 IP
```

## 本地测试

```bash
cd deploy
cp env/infra.env.example env/infra.env   # 填密码或保持 CHANGE_ME
cp env/apps.env.example env/apps.env
cp env/sms.env.example env/sms.env
# 去掉 nginx 容器（无 SSL 证书），直接访问前端容器端口:
docker compose --env-file env/infra.env up --build \
  apps-frontend sms-frontend web apps-backend sms-backend postgres redis newapi
```

apps-frontend 暴露端口需加 `ports` 或用 `docker compose port` 查看。

## 运维

```bash
DC="docker compose -f deploy/docker-compose.yml --env-file deploy/env/infra.env"

# 状态
ssh ECS "cd /opt/tokenjoy/src && $DC ps"

# 日志
ssh ECS "cd /opt/tokenjoy/src && $DC logs -f apps-backend --tail 100"

# 回滚
ssh ECS "cd /opt/tokenjoy/src && git checkout <sha> && $DC up -d --build"

# 只重启某个服务
ssh ECS "cd /opt/tokenjoy/src && $DC restart apps-backend"

# 数据库备份
ssh ECS "cd /opt/tokenjoy/src && $DC exec -T postgres pg_dump -U tokenjoy tokenjoy" > backup.sql

# 重建 newapi (更新 patch 后)
ssh ECS "cd /opt/tokenjoy/src && $DC up -d --build newapi"
```

## 域名

| 域名 | 服务 |
|------|------|
| app.tokenjoy.me | apps |
| sms.tokenjoy.me | sms |
| www.tokenjoy.me | web |

## 生成密码参考

```bash
openssl rand -hex 16          # PG_PASSWORD, REDIS_PASSWORD, SESSION_SECRET 等
openssl rand -hex 32          # INVITE_SECRET
openssl rand -base64 32       # DATA_SOURCE_CREDENTIAL_KEY
```
