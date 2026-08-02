#!/usr/bin/env bash
# 本地 Mac 构建 linux/amd64 镜像 → scp 到 ECS → 加载
# 用法: ./deploy/push-images.sh <ECS_IP>
# ECS 上之后跑 init.sh，它会检测镜像已存在直接启动
set -euo pipefail

SERVER_HOST="${1:-${DEPLOY_HOST:?用法: ./deploy/push-images.sh <ECS_IP> 或设置 DEPLOY_HOST}}"
SERVER_USER="${DEPLOY_USER:-root}"
SSH_KEY="${SSH_KEY:-~/.ssh/id_ed25519_aliyun}"
SSH="ssh -o StrictHostKeyChecking=accept-new -i ${SSH_KEY} ${SERVER_USER}@${SERVER_HOST}"
SCP="scp -o StrictHostKeyChecking=accept-new -i ${SSH_KEY}"
IMAGES_DIR="/tmp/tokenjoy-images"
REMOTE_DIR="/tmp/tokenjoy-images"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

SERVICES=(newapi apps-backend apps-frontend sms-backend sms-frontend web)
DC="docker compose -f deploy/docker-compose.yml"

echo "═══ 构建并推送镜像到 ${SERVER_HOST} ═══"

# ─── 1. 本地构建 linux/amd64 ─────────────────────────────────
echo ""
echo ">>> [1/3] 构建镜像 (linux/amd64)..."
for svc in "${SERVICES[@]}"; do
  echo "  构建 ${svc}..."
  DOCKER_DEFAULT_PLATFORM=linux/amd64 $DC build "${svc}"
done

# ─── 2. 导出压缩 ─────────────────────────────────────────────
echo ""
echo ">>> [2/3] 导出镜像..."
rm -rf "${IMAGES_DIR}" && mkdir -p "${IMAGES_DIR}"
for svc in "${SERVICES[@]}"; do
  IMG="deploy-${svc}:latest"
  echo "  ${svc} → ${IMG}"
  docker save "${IMG}" | gzip > "${IMAGES_DIR}/${svc}.tar.gz"
done
echo "  总大小: $(du -sh ${IMAGES_DIR} | cut -f1)"

# ─── 3. 传输 + 加载 ──────────────────────────────────────────
echo ""
echo ">>> [3/3] 传输到 ECS..."
${SSH} "mkdir -p ${REMOTE_DIR} /opt/mytokenjoy"

# 上传 deploy 目录（ECS 运行所需的全部文件，排除本地 secrets）
echo "  上传 deploy/ 目录..."
rsync -az \
  --exclude='ssl/*.pem' --exclude='ssl/*.key' \
  --exclude='push-images.sh' \
  --exclude='docker-compose.local.yml' --exclude='dockerfiles' \
  --exclude='README.md' \
  --exclude='env/secret.env' \
  -e "ssh -o StrictHostKeyChecking=accept-new -i ${SSH_KEY}" \
  deploy/ "${SERVER_USER}@${SERVER_HOST}:/opt/mytokenjoy/deploy/"

# 上传镜像
echo "  上传镜像文件..."
${SCP} ${IMAGES_DIR}/*.tar.gz "${SERVER_USER}@${SERVER_HOST}:${REMOTE_DIR}/"

echo "  加载镜像..."
${SSH} << 'EOF'
set -euo pipefail
for f in /tmp/tokenjoy-images/*.tar.gz; do
  echo "  加载 $(basename $f)..."
  gunzip -c "$f" | docker load
done
rm -rf /tmp/tokenjoy-images
echo "  ✓ 完成"
EOF

rm -rf "${IMAGES_DIR}"

echo ""
echo "═══ 镜像已就绪 ═══"
echo "ECS 上执行: cd /opt/mytokenjoy && ./deploy/init.sh"
