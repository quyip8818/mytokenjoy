#!/usr/bin/env bash
# 全量重置：推镜像 + 清数据 + 重新初始化（本地执行）
# 用法: ./deploy/reset.sh <ECS_IP>
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

SERVER_HOST="${1:-${DEPLOY_HOST:?用法: ./deploy/reset.sh <ECS_IP> 或设置 DEPLOY_HOST}}"
SSH_KEY="${SSH_KEY:-~/.ssh/id_ed25519_aliyun}"
SERVER_USER="${DEPLOY_USER:-root}"
SSH="ssh -o StrictHostKeyChecking=accept-new -i ${SSH_KEY} ${SERVER_USER}@${SERVER_HOST}"

echo "⚠️  将清除 ECS 上所有数据（数据库、env）并重新初始化"
read -rp "确认? (y/N) " c
[[ "$c" =~ ^[yY]$ ]] || { echo "取消。"; exit 0; }

echo "═══ 全量重置 ${SERVER_HOST} ═══"

# 1. 构建 + 推送镜像
./deploy/push-images.sh "${SERVER_HOST}"

# 2. 远程：停服 + 清数据 + 重新初始化
echo ""
echo ">>> 重置并重新初始化..."
${SSH} << 'EOF'
set -euo pipefail
cd /opt/mytokenjoy
DC="docker compose -f deploy/docker-compose.yml --env-file deploy/env/secret.env"
$DC down -v 2>/dev/null || true
rm -f deploy/env/secret.env
./deploy/init.sh
EOF

echo ""
echo "═══ 重置完成 ═══"
