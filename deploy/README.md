# 部署

本地 Mac 构建镜像 → 推送到 ECS → ECS 只负责运行。

## 域名

| 域名 | 服务 | 说明 |
|------|------|------|
| app.tokenjoy.me | apps-frontend + apps-backend | 客户平台 |
| sms.tokenjoy.me | sms-frontend + sms-backend | 供应商管理 |
| www.tokenjoy.me | web | 官网 |
| tokenjoy.me | → 301 www | 裸域名重定向 |

## 文件

```
deploy/
├── push-images.sh         本地执行：构建镜像 + scp 到 ECS + 加载
├── update.sh              本地执行：push-images + 重启服务（日常更新一键完成）
├── init.sh                ECS 执行：首次部署（装 Docker + 生成密码 + certbot + 启动）
├── docker-compose.yml     服务编排
├── nginx.conf             入口网关（SSL 终止 + 域名路由）
├── init-db.sh             PG 首次初始化（创建数据库）
├── dockerfiles/           Dockerfile（本地构建用，不上传 ECS）
├── env/                   环境变量（init.sh 自动生成，不入 git）
└── ssl/                   证书（certbot 管理，不入 git）
```

## 首次部署

```bash
# 1. DNS 解析（阿里云控制台加 A 记录指向 ECS IP）
#    app → 47.99.63.146
#    sms → 47.99.63.146
#    www → 47.99.63.146
#    @   → 47.99.63.146

# 2. 本地构建 + 推送镜像
./deploy/push-images.sh 47.99.63.146

# 3. ECS 上执行初始化
ssh root@47.99.63.146
cd /opt/mytokenjoy && ./deploy/init.sh
```

init.sh 自动完成：
1. 安装 Docker（阿里云镜像源）+ 配置镜像加速
2. 生成随机密码写入 env 文件（含 NEWAPI_ROOT_PASSWORD）
3. Let's Encrypt 证书申请（certbot standalone）
4. 启动基础设施（postgres + redis + newapi）
5. 初始化 NewAPI admin + 生成 access_token
6. 启动所有应用服务

## 日常更新

```bash
./deploy/update.sh 47.99.63.146
```

等同于：
1. `push-images.sh`（构建 + 推镜像 + 上传 deploy 目录）
2. SSH 到 ECS 执行 `docker compose up -d`

## 全量重置

```bash
# local
./deploy/reset.sh 47.99.63.146

# ECS 上
cd /opt/mytokenjoy
docker compose -f deploy/docker-compose.yml --env-file deploy/env/infra.env down -v
rm -f deploy/env/infra.env deploy/env/apps.env deploy/env/sms.env
./deploy/init.sh
```

## 运维

```bash
# SSH 到 ECS 后
cd /opt/mytokenjoy
DC="docker compose -f deploy/docker-compose.yml --env-file deploy/env/infra.env"

$DC ps                              # 查看状态
$DC logs -f apps-backend --tail 50  # 看日志
$DC restart apps-backend            # 重启单个服务
$DC restart nginx                   # 重载 nginx 配置
```
