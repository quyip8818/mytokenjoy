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
├── deploy.sh              一键部署 (SSH → git pull → up --build)
├── setup-ecs.sh           ECS 首次初始化
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
# 一条命令搞定（初始化 ECS + 生成密码 + 上传证书 + 启动 + 初始化 NewAPI）
./deploy/first-deploy.sh <ECS_IP> <REPO_URL>

# 例:
./deploy/first-deploy.sh 47.96.1.2 git@github.com:yourorg/mytokenjoy.git
```

脚本自动完成:
1. 装 Docker + Git + clone 仓库
2. 生成随机密码写入 env 文件
3. 上传本地 `deploy/ssl/` 下的 SSL 证书
4. `docker compose up -d --build` 启动全栈
5. 初始化 NewAPI admin 并设置 access_token
6. 重启 apps-backend

前提：
- 本地 `deploy/ssl/` 下有 SSL 证书（`*.pem` + `*.key`）
- ECS 能 SSH（默认用 `~/.ssh/id_rsa`，可通过 `SSH_KEY` 环境变量覆盖）
- ECS 能 git clone 仓库（SSH key 或 HTTPS token）

如需手动分步操作:

```bash
# 1. 初始化 ECS
REPO_URL=https://github.com/你的org/mytokenjoy.git \
  ssh root@ECS "bash -s" < deploy/setup-ecs.sh

# 2. 编辑 env
ssh root@ECS "vi /opt/tokenjoy/src/deploy/env/infra.env"
ssh root@ECS "vi /opt/tokenjoy/src/deploy/env/apps.env"
ssh root@ECS "vi /opt/tokenjoy/src/deploy/env/sms.env"

# 3. 上传 SSL 证书
scp certs/*.pem certs/*.key root@ECS:/opt/tokenjoy/src/deploy/ssl/

# 4. 部署
./deploy/deploy.sh <ECS_IP>
```

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
