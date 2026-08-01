#!/usr/bin/env bash
# ECS 首次部署（在 ECS 本机执行，幂等可重跑）
# 用法: cd /opt/mytokenjoy && ./deploy/init.sh
#
# 可选环境变量:
#   NEWAPI_ROOT_PASSWORD  NewAPI root 密码（默认随机生成）
#   DOMAIN_APP            apps 域名（默认 app.tokenjoy.me）
#   DOMAIN_SMS            sms 域名（默认 sms.tokenjoy.me）
#   DOMAIN_WEB            web 域名（默认 www.tokenjoy.me）
set -euo pipefail

# ─── 配置 ─────────────────────────────────────────────────────
DOMAIN_APP="${DOMAIN_APP:-app.tokenjoy.me}"
DOMAIN_SMS="${DOMAIN_SMS:-sms.tokenjoy.me}"
DOMAIN_WEB="${DOMAIN_WEB:-www.tokenjoy.me}"
NEWAPI_ROOT_PASSWORD="${NEWAPI_ROOT_PASSWORD:-$(openssl rand -hex 8)}"
DC="docker compose -f deploy/docker-compose.yml --env-file deploy/env/infra.env"

# ─── 工具函数 ─────────────────────────────────────────────────
log()  { echo "  $*"; }
step() { echo ""; echo ">>> $*"; }
gen()  { openssl rand -hex 16; }
gen32(){ openssl rand -hex 32; }
genb64(){ openssl rand -base64 32; }
need_generate() { [ ! -f "$1" ] || grep -q "CHANGE_ME" "$1" 2>/dev/null; }

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

if need_generate deploy/env/infra.env; then
  cat > deploy/env/infra.env <<EOF
PG_PASSWORD=$(gen)
REDIS_PASSWORD=$(gen)
NEWAPI_SESSION_SECRET=$(gen)
WEBHOOK_SECRET=$(gen)
EOF
  log "✓ infra.env"
else
  log "· infra.env 已存在"
fi

if need_generate deploy/env/apps.env; then
  cat > deploy/env/apps.env <<EOF
PORT=8010
DEPLOY_ENV=production
SUPPORT_SAAS=true
SECURE_COOKIE=true
COMPANY_NAME=TokenJoy
SESSION_SECRET=$(gen)
INVITE_SECRET=$(gen32)
DATA_SOURCE_CREDENTIAL_KEY=$(genb64)
NEW_API_ENABLED=true
NEW_API_GATEWAY_ENABLED=true
CORS_ORIGINS=https://${DOMAIN_APP}
FRONTEND_URL=https://${DOMAIN_APP}
PLATFORM_BOOTSTRAP_EMAIL=admin@tokenjoy.me
PLATFORM_BOOTSTRAP_PASSWORD=$(gen)
EOF
  log "✓ apps.env"
else
  log "· apps.env 已存在"
fi

if need_generate deploy/env/sms.env; then
  cat > deploy/env/sms.env <<EOF
PORT=8020
JWT_SECRET=$(gen)
CORS_ORIGINS=https://${DOMAIN_SMS}
SECURE_COOKIE=true
EOF
  log "✓ sms.env"
else
  log "· sms.env 已存在"
fi

# ─── 3. SSL 证书 ──────────────────────────────────────────────
step "[3/5] SSL 证书"
MISSING=0
for d in "$DOMAIN_APP" "$DOMAIN_SMS" "$DOMAIN_WEB"; do
  [ -f "deploy/ssl/${d}.pem" ] && [ -f "deploy/ssl/${d}.key" ] || { log "⚠️  缺少 ${d}"; MISSING=1; }
done
if [ "$MISSING" -eq 1 ]; then
  read -rp "  证书不全，继续? (y/N) " c
  [[ "$c" =~ ^[yY]$ ]] || { echo "中止。上传证书到 deploy/ssl/ 后重试。"; exit 1; }
else
  log "✓ 齐全"
fi

# ─── 4. 启动全栈 ──────────────────────────────────────────────
step "[4/5] 构建并启动（首次约 5-10 分钟）"
$DC up -d --build --remove-orphans
log "等待服务就绪..."
sleep 15
$DC ps --format "table {{.Name}}\t{{.Status}}"

# ─── 5. 初始化 NewAPI ─────────────────────────────────────────
step "[5/5] NewAPI 初始化"

# 等就绪
for _ in $(seq 1 30); do
  $DC exec -T newapi wget -qO- http://localhost:3000/api/setup 2>/dev/null | grep -q '"success"' && break
  sleep 2
done

STATUS=$($DC exec -T newapi wget -qO- http://localhost:3000/api/setup 2>/dev/null || echo '{}')
if echo "$STATUS" | grep -q '"root_init":false'; then
  RESULT=$($DC exec -T newapi wget -qO- \
    --post-data="{\"username\":\"root\",\"password\":\"${NEWAPI_ROOT_PASSWORD}\",\"confirmPassword\":\"${NEWAPI_ROOT_PASSWORD}\"}" \
    --header='Content-Type: application/json' \
    http://localhost:3000/api/setup 2>/dev/null)
  echo "$RESULT" | grep -q '"success":true' && log "✓ admin 创建成功" || log "⚠️  响应: $RESULT"
else
  log "· 已初始化"
fi

# 设置 access_token
TOKEN="sk-admin-$(openssl rand -hex 16)"
$DC exec -T postgres psql -U tokenjoy -d newapi -c \
  "UPDATE users SET access_token = '${TOKEN}' WHERE id = 1 AND (access_token IS NULL OR access_token = '');" \
  >/dev/null 2>&1
log "✓ access_token 已设置"

$DC restart apps-backend >/dev/null 2>&1
log "✓ apps-backend 重启完成"

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
