#!/usr/bin/env bash
# ECS 首次部署（在 ECS 本机执行，幂等可重跑）
# 用法: cd /opt/mytokenjoy && ./deploy/init.sh
#
# 可选环境变量:
#   DOMAIN_APP            apps 域名（默认 app.tokenjoy.me）
#   DOMAIN_SMS            sms 域名（默认 sms.tokenjoy.me）
#   DOMAIN_WEB            web 域名（默认 www.tokenjoy.me）
set -euo pipefail

# ─── 配置 ─────────────────────────────────────────────────────
DOMAIN_APP="${DOMAIN_APP:-app.tokenjoy.me}"
DOMAIN_SMS="${DOMAIN_SMS:-sms.tokenjoy.me}"
DOMAIN_WEB="${DOMAIN_WEB:-www.tokenjoy.me}"
DC="docker compose -f deploy/docker-compose.yml --env-file deploy/env/secret.env"
CTR_NEWAPI="deploy-newapi-1"

# ─── 工具函数 ─────────────────────────────────────────────────
log()  { echo "  $*"; }
step() { echo ""; echo ">>> $*"; }
gen()  { openssl rand -hex 16; }
gen32(){ openssl rand -hex 32; }
genb64(){ openssl rand -base64 32; }

# ─── 前置检查 ─────────────────────────────────────────────────
[ -f deploy/docker-compose.yml ] || { echo "错误: 请在仓库根目录执行 (cd /opt/mytokenjoy)"; exit 1; }
echo "═══ TokenJoy 首次部署 ═══"

# ─── 1. Docker ────────────────────────────────────────────────
step "[1/5] Docker"
if command -v docker &>/dev/null; then
  log "已存在 ($(docker --version | cut -d' ' -f3))"
else
  # ponytail: 阿里云镜像 + noble 包（Docker CE 尚未为 26.04 发包，二进制兼容）
  apt-get update -qq
  apt-get install -y -qq ca-certificates curl gnupg
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg \
    | gpg --dearmor --yes -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg
  ARCH=$(dpkg --print-architecture)
  echo "deb [arch=${ARCH} signed-by=/etc/apt/keyrings/docker.gpg] https://mirrors.aliyun.com/docker-ce/linux/ubuntu noble stable" \
    > /etc/apt/sources.list.d/docker.list
  apt-get update -qq
  apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-compose-plugin
  systemctl enable --now docker
  log "✓ 已安装"
fi

# 配置国内镜像加速（幂等）
if [ ! -f /etc/docker/daemon.json ] || ! grep -q "registry-mirrors" /etc/docker/daemon.json 2>/dev/null; then
  mkdir -p /etc/docker
  cat > /etc/docker/daemon.json <<'MIRRORS'
{
  "registry-mirrors": [
    "https://mirror.ccs.tencentyun.com",
    "https://docker.m.daocloud.io"
  ]
}
MIRRORS
  systemctl restart docker
  log "✓ 镜像加速已配置"
else
  log "· 镜像加速已存在"
fi

# ─── 2. 生成 env 文件 ─────────────────────────────────────────
step "[2/5] 环境配置"
mkdir -p deploy/env

# secret.env: 只有 init.sh 自动生成的密钥，ECS 上一次生成永久使用
if [ ! -f deploy/env/secret.env ]; then
  cat > deploy/env/secret.env <<EOF
PG_PASSWORD=$(gen)
REDIS_PASSWORD=$(gen)
NEWAPI_SESSION_SECRET=$(gen)
NEWAPI_ROOT_PASSWORD=$(openssl rand -hex 8)
WEBHOOK_SECRET=$(gen)
SESSION_SECRET=$(gen)
INVITE_SECRET=$(gen32)
DATA_SOURCE_CREDENTIAL_KEY=$(genb64)
JWT_SECRET=$(gen)
EOF
  log "✓ secret.env 已生成"
else
  log "· secret.env 已存在"
fi

# 从 secret.env 读取变量
# shellcheck disable=SC1091
set -a && source deploy/env/secret.env && set +a

# apps.env / sms.env / .secret.env: 本地上传的配置，不由 init 生成
if [ ! -f deploy/env/apps.env ]; then
  log "⚠️  deploy/env/apps.env 不存在，请通过 push-images.sh 上传"
fi
if [ ! -f deploy/env/sms.env ]; then
  log "⚠️  deploy/env/sms.env 不存在，请通过 push-images.sh 上传"
fi
# .secret.env 可能还没上传，创建空文件避免 compose 报错
touch deploy/env/.secret.env

# ─── 3. SSL 证书（Let's Encrypt） ─────────────────────────────
step "[3/5] SSL 证书"
if ! command -v certbot &>/dev/null; then
  apt-get install -y -qq certbot
  log "✓ certbot 已安装"
fi

# ponytail: standalone 模式申请，此时 nginx 还没启动所以 80 端口空闲。
# 续期也用 standalone，通过 pre/post hook 短暂停启 nginx（约 5 秒中断，90 天一次）。
CERT_NAME="tokenjoy"
CERT_DIR="/etc/letsencrypt/live/${CERT_NAME}"

if [ ! -f "${CERT_DIR}/fullchain.pem" ]; then
  log "申请证书（需要 80 端口空闲 + DNS 已解析）..."
  certbot certonly --standalone --non-interactive --agree-tos \
    --register-unsafely-without-email \
    --cert-name "${CERT_NAME}" \
    -d "${DOMAIN_APP}" -d "${DOMAIN_SMS}" -d "${DOMAIN_WEB}" -d "tokenjoy.me"
  log "✓ 证书已申请"
else
  log "· 证书已存在 (${CERT_DIR})"
  # 尝试续期（幂等，未到期不会操作）
  certbot renew --quiet || true
fi

# 配置自动续期 hook：停 nginx → 续期 → 启 nginx（幂等）
NGINX_CMD="docker compose -f /opt/mytokenjoy/deploy/docker-compose.yml --env-file /opt/mytokenjoy/deploy/env/secret.env"
DEPLOY_HOOK="/etc/letsencrypt/renewal-hooks/deploy/restart-nginx.sh"
if [ ! -f "${DEPLOY_HOOK}" ]; then
  mkdir -p "$(dirname "${DEPLOY_HOOK}")"
  cat > "${DEPLOY_HOOK}" <<HOOK
#!/bin/bash
${NGINX_CMD} restart nginx
HOOK
  chmod +x "${DEPLOY_HOOK}"
  log "✓ 自动续期 hook 已配置"
fi

# 配置 certbot renewal 使用 standalone（停 nginx 让出 80）
RENEWAL_CONF="/etc/letsencrypt/renewal/${CERT_NAME}.conf"
if [ -f "${RENEWAL_CONF}" ] && ! grep -q "pre_hook" "${RENEWAL_CONF}"; then
  cat >> "${RENEWAL_CONF}" <<RENEW
pre_hook = ${NGINX_CMD} stop nginx
post_hook = ${NGINX_CMD} start nginx
RENEW
  log "✓ 续期 pre/post hook 已配置"
fi

# ─── 4. 启动全栈 ──────────────────────────────────────────────
step "[4/5] 启动服务"
BUILD_FLAG=""
if docker images --format '{{.Repository}}' | grep -q "deploy-apps-backend"; then
  log "镜像已存在，跳过构建"
else
  BUILD_FLAG="--build"
fi

# 先启动基础设施 + NewAPI（apps-backend 需要 NewAPI token 才能健康启动）
log "启动基础设施..."
$DC up -d $BUILD_FLAG --remove-orphans postgres redis newapi
log "等待 NewAPI 就绪..."
sleep 10

# ─── 5. 初始化 NewAPI ─────────────────────────────────────────
step "[5/6] NewAPI 初始化"

# 等就绪
for _ in $(seq 1 30); do
  docker exec "${CTR_NEWAPI}" wget -qO- http://localhost:3000/api/setup 2>/dev/null | grep -q '"success"' && break
  sleep 2
done

STATUS=$(docker exec "${CTR_NEWAPI}" wget -qO- http://localhost:3000/api/setup 2>/dev/null || echo '{}')
if echo "$STATUS" | grep -q '"root_init":false'; then
  RESULT=$(docker exec "${CTR_NEWAPI}" wget -qO- \
    --post-data="{\"username\":\"root\",\"password\":\"${NEWAPI_ROOT_PASSWORD}\",\"confirmPassword\":\"${NEWAPI_ROOT_PASSWORD}\"}" \
    --header='Content-Type: application/json' \
    http://localhost:3000/api/setup 2>/dev/null || echo '{}')
  if echo "$RESULT" | grep -q '"success":true'; then
    log "✓ admin 创建成功"
  else
    log "⚠️  响应: $RESULT"
  fi
else
  log "· 已初始化"
fi

# 设置 access_token（通过 NewAPI API 生成，不直接写数据库）
# 登录获取 session
LOGIN_RESP=$(docker exec "${CTR_NEWAPI}" curl -sf \
  -c /tmp/cookies -b /tmp/cookies \
  -X POST http://localhost:3000/api/user/login \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"root\",\"password\":\"${NEWAPI_ROOT_PASSWORD}\"}" 2>/dev/null || echo '{}')

if echo "$LOGIN_RESP" | grep -q '"success":true'; then
  # 获取 token（NewAPI 自动生成 access_token 写入数据库）
  docker exec "${CTR_NEWAPI}" curl -sf \
    -b /tmp/cookies \
    -H "New-Api-User: 1" \
    http://localhost:3000/api/user/token >/dev/null 2>&1
  log "✓ access_token 已通过 API 生成"
else
  log "⚠️  NewAPI 登录失败: $LOGIN_RESP"
fi

# 启动剩余服务（现在 token 已就绪，apps-backend 可以正常启动）
step "[6/6] 启动应用服务"
$DC up -d $BUILD_FLAG --remove-orphans
log "等待服务就绪..."
sleep 10
$DC ps --format "table {{.Name}}\t{{.Status}}"

# ─── 完成 ─────────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════"
echo "  部署完成!"
echo ""
echo "  https://${DOMAIN_APP}  (客户平台)"
echo "  https://${DOMAIN_SMS}  (供应商管理)"
echo "  https://${DOMAIN_WEB}  (官网)"
echo ""
echo "  管理员: deploy/env/apps.env → PLATFORM_BOOTSTRAP_EMAIL/PASSWORD"
echo "  NewAPI:  root / ${NEWAPI_ROOT_PASSWORD}"
echo "═══════════════════════════════════════════"
