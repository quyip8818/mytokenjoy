#!/usr/bin/env bash
# 手动部署脚本（CI 挂了时备用）
# 用法: ./deploy/deploy.sh [image_tag]
#   image_tag 可选，默认 latest
set -euo pipefail

SERVER_USER="${DEPLOY_USER:-root}"
SERVER_HOST="${DEPLOY_HOST:?请设置 DEPLOY_HOST 环境变量}"
SSH_KEY="${SSH_KEY:-~/.ssh/id_rsa}"
IMAGE_TAG="${1:-latest}"

SSH_CMD="ssh -o StrictHostKeyChecking=accept-new -i ${SSH_KEY} ${SERVER_USER}@${SERVER_HOST}"

echo ">>> 部署 IMAGE_TAG=${IMAGE_TAG} 到 ${SERVER_HOST}"

${SSH_CMD} << EOF
set -euo pipefail
cd /opt/tokenjoy

# 更新 IMAGE_TAG
sed -i "s/^IMAGE_TAG=.*/IMAGE_TAG=${IMAGE_TAG}/" .env

# 拉取最新镜像并重启
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d --remove-orphans

# 清理悬空镜像
docker image prune -f

echo "✓ 部署完成 (tag: ${IMAGE_TAG})"
docker compose -f docker-compose.prod.yml ps --format "table {{.Name}}\t{{.Status}}"
EOF
