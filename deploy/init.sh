#!/usr/bin/env bash
# ECS 首次部署（在 ECS 本机执行）
# 用法: SSH 到 ECS 后:
#   git clone <repo> /opt/tokenjoy/src
#   cd /opt/tokenjoy/src
#   ./deploy/init.sh
#
# 可选环境变量:
#   NEWAPI_ROOT_PASSWORD  NewAPI root 密码（默认随机生成）
#   DOMAIN_APP            apps 域名（默认 app.tokenjoy.me）
#   DOMAIN_SMS            sms 域名（默认 sms.tokenjoy.me）
#   DOMAIN_WEB            web 域名（默认 www.tokenjoy.me）
set -euo pipefail

DOMAIN_APP="${DOMAIN_APP:-app.tokenjoy.me}"
DOMAIN_SMS="${DOMAIN_SMS:-sms.tokenjoy.me}"
DOMAIN_WEB="${DOMAIN_WEB:-www.tokenjoy.me}"
NEWAPI_ROOT_PASSWORD="${NEWAPI_ROOT_PASSWORD:-$(openssl rand -hex 8)}"

echo "═══ TokenJoy 首次部署 ═══"
echo ""

# ─── 0. 前置检查 ──────────────────────────────────────────────
if [ ! -f deploy/docker-compose.yml ]; then
  echo "错误: 请在仓库根目录执行此脚本 (cd /opt/tokenjoy/src)"
  exit 1
fi

# ─── 1. 装 Docker ─────────────────────────────────────────────
echo ">>> [1/6] 检查 Docker..."
if ! command -v docker &>/dev/null; then
  echo "  安装 Docker（阿里云镜像）..."
  # ponytail: 国内 ECS 无法访问 get.docker.com，用阿里云镜像 + noble 包（Docker CE 尚未为 26.04 发包）
  apt-get update -qq
  apt-get install -y -qq ca-certificates curl gnupg
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | gpg --dearmor --yes -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg
  echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.gpg] https://mirrors.aliyun.com/docker-ce/linux/ubuntu noble stable" \
    > /etc/apt/sources.list.d/docker.list
  apt-get update -qq
  apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-compose-plugin
  systemctl enable --now docker
  echo "  ✓ Docker 已安装"
else
  echo "  · Docker 已存在 ($(docker --version | cut -d' ' -f3))"
fi

# ─── 2. 生成密码 ──────────────────────────────────────────────
echo ">>> [2/6] 生成配置文件..."
cd deploy/env

gen() { openssl rand -hex 16; }
gen32() { openssl rand -hex 32; }
genb64() { openssl rand -base64 32; }

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
CORS_ORIGINS=https://${DOMAIN_APP}
FRONTEND_URL=https://${DOMAIN_APP}
PLATFORM_BOOTSTRAP_EMAIL=admin@tokenjoy.me
PLATFORM_BOOTSTRAP_PASSWORD=$(gen)
EOF
  echo "  ✓ apps.env 已生成"
else
  echo "  · apps.env 已存在，跳过"
fi

if grep -q "CHANGE_ME" sms.env 2>/dev/null || [ ! -f sms.env ]; then
  cat > sms.env << EOF
PORT=8020
JWT_SECRET=$(gen)
CORS_ORIGINS=https://${DOMAIN_SMS}
SECURE_COOKIE=true
EOF
  echo "  ✓ sms.env 已生成"
else
  echo "  · sms.env 已存在，跳过"
fi

cd ../..

# ─── 3. 检查 SSL 证书 ─────────────────────────────────────────
echo ">>> [3/6] 检查 SSL 证书..."
MISSING_SSL=0
for domain in "${DOMAIN_APP}" "${DOMAIN_SMS}" "${DOMAIN_WEB}"; do
  if [ ! -f "deploy/ssl/${domain}.pem" ] || [ ! -f "deploy/ssl/${domain}.key" ]; then
    echo "  ⚠️  缺少: deploy/ssl/${domain}.pem / .key"
    MISSING_SSL=1
  fi
done
if [ "$MISSING_SSL" -eq 0 ]; then
  echo "  ✓ SSL 证书齐全"
else
  echo ""
  echo "  请上传证书后重新运行，或先跳过 nginx（服务仍可启动）。"
  echo "  继续部署? (y/N)"
  read -r confirm
  if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "中止。请上传证书到 deploy/ssl/ 后重试。"
    exit 1
  fi
fi

# ─── 4. 启动全栈 ──────────────────────────────────────────────
echo ">>> [4/6] 启动全栈（首次构建约 5-10 分钟）..."
DC="docker compose -f deploy/docker-compose.yml --env-file deploy/env/infra.env"
$DC up -d --build
echo "  等待服务就绪..."
sleep 15
$DC ps --format "table {{.Name}}\t{{.Status}}"

# ─── 5. 初始化 NewAPI ─────────────────────────────────────────
echo ">>> [5/6] 初始化 NewAPI..."

# 等 NewAPI 就绪
for i in $(seq 1 30); do
  if $DC exec -T newapi wget -qO- http://localhost:3000/api/setup 2>/dev/null | grep -q '"success"'; then
    break
  fi
  sleep 2
done

STATUS=$($DC exec -T newapi wget -qO- http://localhost:3000/api/setup 2>/dev/null || echo '{}')
if echo "$STATUS" | grep -q '"root_init":false'; then
  RESULT=$($DC exec -T newapi wget -qO- \
    --post-data="{\"username\":\"root\",\"password\":\"${NEWAPI_ROOT_PASSWORD}\",\"confirmPassword\":\"${NEWAPI_ROOT_PASSWORD}\"}" \
    --header='Content-Type: application/json' \
    http://localhost:3000/api/setup 2>/dev/null)
  if echo "$RESULT" | grep -q '"success":true'; then
    echo "  ✓ NewAPI admin 创建成功"
  else
    echo "  ⚠️  NewAPI setup 响应: $RESULT"
  fi
else
  echo "  · NewAPI 已初始化，跳过"
fi

# 设置 access_token
TOKEN="sk-admin-$(openssl rand -hex 16)"
$DC exec -T postgres psql -U tokenjoy -d newapi -c \
  "UPDATE users SET access_token = '${TOKEN}' WHERE id = 1 AND (access_token IS NULL OR access_token = '');" \
  > /dev/null 2>&1
echo "  ✓ admin access_token 已设置"

# 重启 apps-backend
$DC restart apps-backend > /dev/null 2>&1
echo "  ✓ apps-backend 已重启"

# ─── 6. 完成 ──────────────────────────────────────────────────
echo ""
echo ">>> [6/6] 最终状态:"
sleep 5
$DC ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "═══════════════════════════════════════════"
echo "  部署完成!"
echo ""
echo "  访问地址:"
echo "    https://${DOMAIN_APP}  (客户平台)"
echo "    https://${DOMAIN_SMS}  (供应商管理)"
echo "    https://${DOMAIN_WEB}  (官网)"
echo ""
echo "  管理员: 见 deploy/env/apps.env 中的"
echo "    PLATFORM_BOOTSTRAP_EMAIL / PASSWORD"
echo ""
echo "  NewAPI root: root / ${NEWAPI_ROOT_PASSWORD}"
echo ""
echo "  日常部署: cd /opt/tokenjoy/src && git pull && "
echo "    docker compose -f deploy/docker-compose.yml \\"
echo "    --env-file deploy/env/infra.env up -d --build"
echo "═══════════════════════════════════════════"
