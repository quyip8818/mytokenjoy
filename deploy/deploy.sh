#!/usr/bin/env bash
# 部署: SSH 到 ECS → git pull → docker compose up --build
# 用法: ./deploy/deploy.sh [ECS_IP]
set -euo pipefail

SERVER_HOST="${1:-${DEPLOY_HOST:?用法: ./deploy/deploy.sh <ECS_IP> 或设置 DEPLOY_HOST}}"
SERVER_USER="${DEPLOY_USER:-root}"
SSH_KEY="${SSH_KEY:-~/.ssh/id_rsa}"
SSH="ssh -o StrictHostKeyChecking=accept-new -i ${SSH_KEY} ${SERVER_USER}@${SERVER_HOST}"

echo "═══ 部署到 ${SERVER_HOST} ═══"

${SSH} << 'EOF'
set -euo pipefail
cd /opt/tokenjoy/src
DC="docker compose -f deploy/docker-compose.yml --env-file deploy/env/infra.env"

echo ">>> git pull..."
git pull --ff-only

echo ">>> build + up..."
$DC up -d --build --remove-orphans

echo ">>> 清理..."
docker image prune -f

echo ""
echo "✓ 部署完成"
$DC ps --format "table {{.Name}}\t{{.Status}}"
EOF

echo "═══ 完成 ═══"
