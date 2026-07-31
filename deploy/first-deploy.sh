#!/usr/bin/env bash
# 首次部署（从本地执行）：初始化 ECS → 上传证书 → 生成密码 → 启动全栈 → 初始化 NewAPI
# 用法: ./deploy/first-deploy.sh <ECS_IP> <REPO_URL>
#   例: ./deploy/first-deploy.sh 47.96.1.2 git@github.com:yourorg/mytokenjoy.git
set -euo pipefail

SERVER_HOST="${1:?用法: $0 <ECS_IP> <REPO_URL>}"
REPO_URL="${2:?用法: $0 <ECS_IP> <REPO_URL>}"
SERVER_USER="${DEPLOY_USER:-root}"
SSH_KEY="${SSH_KEY:-~/.ssh/id_rsa}"
SSH="ssh -o StrictHostKeyChecking=accept-new -i ${SSH_KEY} ${SERVER_USER}@${SERVER_HOST}"
SCP="scp -o StrictHostKeyChecking=accept-new -i ${SSH_KEY}"
NEWAPI_ROOT_PASSWORD="${NEWAPI_ROOT_PASSWORD:-$(openssl rand -hex 8)}"

DEPLOY_DIR="/opt/tokenjoy/src"

echo "═══ 首次部署到 ${SERVER_HOST} ═══"
echo ""

# ─── 1. 初始化 ECS ────────────────────────────────────────────
echo ">>> [1/6] 初始化 ECS（Docker + Git + Clone）..."
${SSH} "REPO_URL=${REPO_URL} bash -s" < "$(dirname "$0")/setup-ecs.sh"

# ─── 2. 生成密码并写入 env 文件 ───────────────────────────────
echo ">>> [2/6] 生成密码 & 写入 env 文件..."
${SSH} << 'ENVEOF'
set -euo pipefail
cd /opt/tokenjoy/src/deploy/env

gen() { openssl rand -hex 16; }
gen32() { openssl rand -hex 32; }
genb64() { openssl rand -base64 32; }

# infra.env
if grep -q "CHANGE_ME" infra.env 2>/dev/null || [ ! -f infra.env ]; then
  cat > infra.env << EOF
PG_PASSWORD=$(gen)
REDIS_PASSWORD=$(gen)
NEWAPI_SESSION_SECRET=$(gen)
WEBHOOK_SECRET=$(gen)
EOF
  echo "  ✓ infra.env 已生成"
else
  echo "  · infra.env 已存在，跳过"
fi

# apps.env
if grep -q "CHANGE_ME" apps.env 2>/dev/null || [ ! -f apps.env ]; then
  cat > apps.env << EOF
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
CORS_ORIGINS=https://app.tokenjoy.me
FRONTEND_URL=https://app.tokenjoy.me
PLATFORM_BOOTSTRAP_EMAIL=admin@tokenjoy.me
PLATFORM_BOOTSTRAP_PASSWORD=$(gen)
EOF
  echo "  ✓ apps.env 已生成"
else
  echo "  · apps.env 已存在，跳过"
fi

# sms.env
if grep -q "CHANGE_ME" sms.env 2>/dev/null || [ ! -f sms.env ]; then
  cat > sms.env << EOF
PORT=8020
JWT_SECRET=$(gen)
CORS_ORIGINS=https://sms.tokenjoy.me
SECURE_COOKIE=true
EOF
  echo "  ✓ sms.env 已生成"
else
  echo "  · sms.env 已存在，跳过"
fi
ENVEOF

# ─── 3. 上传 SSL 证书 ─────────────────────────────────────────
echo ">>> [3/6] 上传 SSL 证书..."
SSL_DIR="$(dirname "$0")/ssl"
if ls "${SSL_DIR}"/*.pem "${SSL_DIR}"/*.key &>/dev/null; then
  ${SCP} "${SSL_DIR}"/*.pem "${SSL_DIR}"/*.key "${SERVER_USER}@${SERVER_HOST}:${DEPLOY_DIR}/deploy/ssl/"
  echo "  ✓ SSL 证书已上传"
else
  echo "  ⚠️  本地 deploy/ssl/ 下没有证书文件，跳过。请后续手动上传。"
fi

# ─── 4. 启动全栈 ──────────────────────────────────────────────
echo ">>> [4/6] 启动全栈（首次构建可能需要 5-10 分钟）..."
${SSH} << 'UPEOF'
set -euo pipefail
cd /opt/tokenjoy/src
DC="docker compose -f deploy/docker-compose.yml --env-file deploy/env/infra.env"
$DC up -d --build
echo "  等待服务就绪..."
sleep 10
$DC ps --format "table {{.Name}}\t{{.Status}}"
UPEOF

# ─── 5. 初始化 NewAPI ─────────────────────────────────────────
echo ">>> [5/6] 初始化 NewAPI admin..."
${SSH} "export NEWAPI_ROOT_PASSWORD='${NEWAPI_ROOT_PASSWORD}'; bash -s" << 'INITEOF'
set -euo pipefail
cd /opt/tokenjoy/src
DC="docker compose -f deploy/docker-compose.yml --env-file deploy/env/infra.env"

# 等 NewAPI 启动完成
for i in $(seq 1 30); do
  if $DC exec -T newapi wget -qO- http://localhost:3000/api/setup 2>/dev/null | grep -q '"success"'; then
    break
  fi
  sleep 2
done

# 检查是否已初始化
STATUS=$($DC exec -T newapi wget -qO- http://localhost:3000/api/setup 2>/dev/null || echo '{}')
if echo "$STATUS" | grep -q '"root_init":false'; then
  # 执行 setup
  RESULT=$($DC exec -T newapi wget -qO- \
    --post-data="{\"username\":\"root\",\"password\":\"${NEWAPI_ROOT_PASSWORD}\",\"confirmPassword\":\"${NEWAPI_ROOT_PASSWORD}\"}" \
    --header='Content-Type: application/json' \
    http://localhost:3000/api/setup 2>/dev/null)
  if echo "$RESULT" | grep -q '"success":true'; then
    echo "  ✓ NewAPI admin 创建成功 (root / ${NEWAPI_ROOT_PASSWORD})"
  else
    echo "  ⚠️  NewAPI setup 响应: $RESULT"
  fi
else
  echo "  · NewAPI 已初始化，跳过"
fi

# 设置 admin access_token
TOKEN="sk-admin-$(openssl rand -hex 16)"
$DC exec -T postgres psql -U tokenjoy -d newapi -c \
  "UPDATE users SET access_token = '${TOKEN}' WHERE id = 1 AND (access_token IS NULL OR access_token = '');" \
  > /dev/null 2>&1
echo "  ✓ admin access_token 已设置"

# 重启 apps-backend 让它读到新 token
$DC restart apps-backend > /dev/null 2>&1
echo "  ✓ apps-backend 已重启"
INITEOF

# ─── 6. 最终状态 ──────────────────────────────────────────────
echo ">>> [6/6] 最终状态..."
${SSH} << 'STATUSEOF'
cd /opt/tokenjoy/src
DC="docker compose -f deploy/docker-compose.yml --env-file deploy/env/infra.env"
echo ""
$DC ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "═══ 部署完成 ═══"
echo ""
echo "访问地址:"
echo "  https://app.tokenjoy.me  (apps)"
echo "  https://sms.tokenjoy.me  (sms)"
echo "  https://www.tokenjoy.me  (官网)"
echo ""
echo "管理员登录: 查看 apps.env 中的 PLATFORM_BOOTSTRAP_EMAIL / PASSWORD"
echo "NewAPI 后台: 通过 SSH 访问（不对外暴露）"
echo ""
echo "日常部署: ./deploy/deploy.sh $(hostname -I | awk '{print $1}')"
STATUSEOF

echo ""
echo "═══ 首次部署完成 ═══"
echo "NewAPI root 密码: ${NEWAPI_ROOT_PASSWORD}（请妥善保存）"
