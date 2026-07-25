#!/bin/bash
# 本地首次启动：等待 NewAPI 就绪 + 初始化 SMS 数据库
# 前提：已运行 pnpm infra 或 docker compose up -d
set -euo pipefail

SCRIPT_DIR="$(dirname "$0")"

echo "等待 NewAPI (sms) 就绪..."
for i in $(seq 1 30); do
  if curl -sf http://localhost:3020/api/status > /dev/null 2>&1; then
    echo "✓ NewAPI ready"
    break
  fi
  sleep 2
done

echo "初始化 SMS 数据库..."
cd "$SCRIPT_DIR/../../backend"
psql "$(grep DATABASE_URL .env.development | cut -d= -f2-)" -f schema.sql 2>/dev/null || true
echo "✓ done — 访问 http://localhost:3020 设置 NewAPI admin 账号"
