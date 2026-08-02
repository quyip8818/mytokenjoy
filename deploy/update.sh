#!/usr/bin/env bash
# 日常更新：推镜像 + 重启服务（不动数据库/env/证书）
# 用法: ./deploy/update.sh <ECS_IP>
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

SERVER_HOST="${1:-${DEPLOY_HOST:?用法: ./deploy/update.sh <ECS_IP> 或设置 DEPLOY_HOST}}"

echo "═══ 更新部署 ═══"

# 1. 构建 + 推送镜像
./deploy/push-images.sh "${SERVER_HOST}"

# 2. 重启服务
SSH_KEY="${SSH_KEY:-~/.ssh/id_ed25519_aliyun}"
SERVER_USER="${DEPLOY_USER:-root}"
SSH="ssh -o StrictHostKeyChecking=accept-new -i ${SSH_KEY} ${SERVER_USER}@${SERVER_HOST}"

echo ""
echo ">>> 重启服务..."
${SSH} "cd /opt/mytokenjoy && docker compose -f deploy/docker-compose.yml --env-file deploy/env/secret.env up -d"

echo ""
echo "═══ 更新完成 ═══"
