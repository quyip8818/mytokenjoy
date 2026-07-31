#!/usr/bin/env bash
# ECS 首次初始化（在 ECS 上运行）
set -euo pipefail

echo "═══ ECS 初始化 ═══"

# 1. Docker
if ! command -v docker &>/dev/null; then
  echo ">>> 安装 Docker..."
  curl -fsSL https://get.docker.com | sh
  systemctl enable --now docker
fi

# 2. Git
if ! command -v git &>/dev/null; then
  echo ">>> 安装 Git..."
  apt-get update && apt-get install -y git
fi

# 3. Clone 仓库
if [ ! -d /opt/tokenjoy/src ]; then
  echo ">>> Clone 仓库..."
  mkdir -p /opt/tokenjoy
  REPO_URL="${REPO_URL:-}"
  if [ -z "$REPO_URL" ]; then
    echo "请设置 REPO_URL 环境变量，例如:"
    echo "  REPO_URL=https://github.com/yourorg/mytokenjoy.git bash setup-ecs.sh"
    exit 1
  fi
  git clone "$REPO_URL" /opt/tokenjoy/src
fi

# 4. 创建 env 文件（从模板）
cd /opt/tokenjoy/src/deploy
if [ ! -f env/infra.env ]; then
  cp env/infra.env.example env/infra.env
  cp env/apps.env.example env/apps.env
  cp env/sms.env.example env/sms.env
  echo ""
  echo "⚠️  请编辑以下文件填入真实密码:"
  echo "  vi /opt/tokenjoy/src/deploy/env/infra.env"
  echo "  vi /opt/tokenjoy/src/deploy/env/apps.env"
  echo "  vi /opt/tokenjoy/src/deploy/env/sms.env"
  echo ""
  echo "  生成密码: openssl rand -hex 16"
  echo "  DATA_SOURCE_CREDENTIAL_KEY: openssl rand -base64 32"
  echo "  INVITE_SECRET: openssl rand -hex 32"
fi

# 5. SSL 证书目录
mkdir -p /opt/tokenjoy/src/deploy/ssl

cat << 'DONE'

═══ 初始化完成 ═══

后续步骤:
1. 编辑 env 文件（填密码）
2. 放 SSL 证书到 deploy/ssl/
   - app.tokenjoy.me.pem + app.tokenjoy.me.key
   - sms.tokenjoy.me.pem + sms.tokenjoy.me.key
   - www.tokenjoy.me.pem + www.tokenjoy.me.key
3. 部署:
   cd /opt/tokenjoy/src
   docker compose -f deploy/docker-compose.yml --env-file deploy/env/infra.env up -d --build

   或从本地:
   ./deploy/deploy.sh <ECS_IP>

DONE
