# 部署指南

TokenJoy 三平台阿里云 ECS 部署，全镜像化方案。

## 架构概览

```
GitHub (push main)
  → GitHub Actions: 构建 6 个 Docker 镜像
  → 推送到阿里云 ACR
  → SSH 到 ECS: pull + up -d

ECS 内部:
  nginx:443 ──┬── www.tokenjoy.me → web (静态)
              ├── app.tokenjoy.me → apps-frontend → apps-backend → PG/Redis/NewAPI
              └── sms.tokenjoy.me → sms-frontend → sms-backend → PG/Redis/NewAPI-SMS
```

## 目录结构

```
deploy/
├── dockerfiles/              # 各服务 Dockerfile（build context = repo root）
│   ├── apps-backend.Dockerfile
│   ├── apps-frontend.Dockerfile
│   ├── sms-backend.Dockerfile
│   ├── sms-frontend.Dockerfile
│   └── web.Dockerfile
├── nginx/                    # Nginx 配置
│   ├── gateway.conf          # 入口网关（SSL 终止 + 域名路由）
│   ├── apps-frontend.conf    # apps 前端容器内（SPA + API 反代）
│   └── sms-frontend.conf    # sms 前端容器内（SPA + API 反代）
├── docker-compose.prod.yml   # 生产 Compose（ECS 上直接用）
├── .dockerignore             # 构建时排除文件
├── .env.example              # 环境变量模板
├── deploy.sh                 # 手动部署脚本（备用）
└── README.md                 # 本文档
```

## 首次部署（Step by Step）

### 1. 阿里云资源准备

| 资源 | 操作 |
|------|------|
| ECS | 购买 2C4G+，选 Ubuntu 22.04 或 Alibaba Cloud Linux |
| ACR | 创建个人版实例，命名空间 `tokenjoy`，创建 6 个镜像仓库 |
| 域名 | 三条 A 记录指向 ECS 公网 IP |
| SSL  | 申请免费证书（每个域名一张）或使用通配符证书 |

ACR 镜像仓库清单：
- `tokenjoy-apps-frontend`
- `tokenjoy-apps-backend`
- `tokenjoy-newapi`
- `tokenjoy-sms-frontend`
- `tokenjoy-sms-backend`
- `tokenjoy-web`

### 2. ECS 初始化

```bash
# SSH 到 ECS
ssh root@你的IP

# 安装 Docker
curl -fsSL https://get.docker.com | sh
systemctl enable --now docker

# 创建目录
mkdir -p /opt/tokenjoy/{ssl,nginx,postgres-init}

# 安全组：只开放 80、443、SSH(限 IP)
```

### 3. 上传配置文件

在本地项目根目录执行：

```bash
SERVER="root@你的ECS_IP"

# 上传 Compose 文件
scp deploy/docker-compose.prod.yml $SERVER:/opt/tokenjoy/

# 上传 nginx 入口配置
scp deploy/nginx/gateway.conf $SERVER:/opt/tokenjoy/nginx/

# 上传数据库初始化脚本
scp scripts/postgres-init/01-create-all-dbs.sh $SERVER:/opt/tokenjoy/postgres-init/

# 上传 SSL 证书（你自己准备的）
scp ssl/*.pem ssl/*.key $SERVER:/opt/tokenjoy/ssl/

# 创建 .env（从模板编辑，填入真实密码）
scp deploy/.env.example $SERVER:/opt/tokenjoy/.env
ssh $SERVER "vim /opt/tokenjoy/.env"  # 编辑填入真实值
```

### 4. 配置 GitHub Secrets

在仓库 Settings → Secrets and variables → Actions 中添加：

| Secret | 值 |
|--------|---|
| `ACR_USERNAME` | ACR 登录用户名 |
| `ACR_PASSWORD` | ACR 登录密码 |
| `DEPLOY_HOST` | ECS 公网 IP |
| `DEPLOY_USER` | `root`（或创建专用部署用户） |
| `SSH_PRIVATE_KEY` | SSH 私钥内容 |

创建 Environment `production`（可配审批人，可选）。

### 5. 首次启动

推送到 main 分支，CI 会自动构建并部署。或手动触发：

```bash
# 本地构建推镜像（如果 CI 还没配好）
docker login registry.cn-hangzhou.aliyuncs.com
docker build -f deploy/dockerfiles/apps-backend.Dockerfile -t registry.cn-hangzhou.aliyuncs.com/tokenjoy/tokenjoy-apps-backend:latest .
docker push registry.cn-hangzhou.aliyuncs.com/tokenjoy/tokenjoy-apps-backend:latest
# ... 重复其他 5 个镜像

# 在 ECS 上拉起
ssh root@ECS_IP "cd /opt/tokenjoy && docker compose -f docker-compose.prod.yml pull && docker compose -f docker-compose.prod.yml up -d"
```

---

## 日常运维

### 查看服务状态

```bash
ssh ECS "cd /opt/tokenjoy && docker compose -f docker-compose.prod.yml ps"
```

### 查看日志

```bash
# 跟踪 apps 后端日志
ssh ECS "cd /opt/tokenjoy && docker compose -f docker-compose.prod.yml logs -f apps-backend --tail 100"
```

### 手动部署指定版本

```bash
export DEPLOY_HOST=你的IP
./deploy/deploy.sh abc12345   # 指定 git sha 前 8 位
```

### 回滚

```bash
export DEPLOY_HOST=你的IP
./deploy/deploy.sh 上一个tag   # 用之前的 git sha
```

### 数据库备份

```bash
ssh ECS "cd /opt/tokenjoy && docker compose -f docker-compose.prod.yml exec -T postgres pg_dump -U tokenjoy tokenjoy" > backup_$(date +%Y%m%d).sql
```

### 证书续期

如使用 Let's Encrypt，建议在 ECS 上装 certbot：

```bash
apt install certbot
certbot certonly --standalone -d www.tokenjoy.me -d app.tokenjoy.me -d sms.tokenjoy.me
# 配置 cron 每月自动续期
echo "0 3 1 * * certbot renew --deploy-hook 'docker compose -f /opt/tokenjoy/docker-compose.prod.yml restart nginx'" | crontab -
```

---

## 安全检查清单

- [ ] ECS 安全组：仅开放 80/443/SSH（SSH 限 IP）
- [ ] .env 中所有密码已替换为强随机值（`openssl rand -hex 16`）
- [ ] SSL 证书已部署且有效
- [ ] Redis 设置了密码 + maxmemory 限制
- [ ] PostgreSQL 不暴露公网端口
- [ ] Docker 容器以非 root 用户运行（Go 后端已配置）
- [ ] Nginx 开启 HSTS + 安全头
- [ ] 日志有大小限制（10MB × 3 文件轮转）

---

## 资源用量估算（2C4G ECS）

| 服务 | 内存限制 | 实际预估 |
|------|----------|----------|
| PostgreSQL | 1G | 200-500MB |
| Redis | 512M | 50-100MB |
| NewAPI-apps | 512M | 100-200MB |
| NewAPI-sms | 512M | 100-200MB |
| apps-backend | 256M | 30-80MB |
| sms-backend | 256M | 30-80MB |
| apps-frontend (nginx) | 64M | 5-10MB |
| sms-frontend (nginx) | 64M | 5-10MB |
| web (nginx) | 64M | 5-10MB |
| gateway nginx | 128M | 10-20MB |
| **合计** | **3.3G** | **~800MB-1.2GB** |

4G 内存足够。预留给操作系统约 800MB。

---

## 升级路径

| 触发条件 | 操作 |
|----------|------|
| 内存不够 | PG 外移到阿里云 RDS（改 .env 连接串即可） |
| 需要 HA | 加一台 ECS + ALB 做蓝绿部署 |
| 需要自动扩缩 | 迁移到 SAE（镜像不变，改部署目标） |
| 想全自动 | 加 Watchtower 容器，ACR 有新镜像自动重启 |
| 构建太慢 | CI 已配 GitHub Actions Cache，后续可加并行 job |

---

## 故障排查

```bash
# 服务启动失败
docker compose -f docker-compose.prod.yml logs <service-name>

# 数据库连接失败
docker compose -f docker-compose.prod.yml exec postgres psql -U tokenjoy -c "SELECT 1"

# Nginx 配置错误
docker compose -f docker-compose.prod.yml exec nginx nginx -t

# 磁盘满了
docker system df              # 查看 Docker 磁盘用量
docker image prune -a -f     # 清理所有未使用镜像
docker volume prune -f       # 清理未使用卷（⚠️ 注意不要删数据卷）

# 内存不足
docker stats --no-stream     # 查看各容器实时资源用量
```
