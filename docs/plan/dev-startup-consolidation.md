# 本地开发启动脚本合并（已实现）

## 改动摘要

将 `docker:reset` / `start:local` / `start:local:empty` / `start:saas` 合并为：

| 命令 | 职责 |
|------|------|
| `pnpm reset [local\|saas] [--empty]` | wipe volume + infra + 按模式修改 `.env.development` + seed |
| `pnpm start` | 纯启动前后端，无参数 |

---

## 用法

```bash
pnpm reset                # 单公司 + demo 数据（默认）
pnpm reset local --empty  # 单公司 + 只有 company + admin
pnpm reset saas           # 多租户 + demo 数据
pnpm reset saas --empty   # 多租户骨架

pnpm start                # 启动（读取 .env.local-mode 中的 SUPPORT_SAAS）
pnpm start:lite           # 轻量：只 postgres + backend + frontend
```

---

## 实现细节

### reset.sh

1. 解析参数（mode + --empty）
2. 用 `set_env` helper 更新 `.env.development` 中的模式 key：
   - `BOOTSTRAP_MODE`（demo / prod）
   - `SUPPORT_SAAS`（true / false）
   - `BOOTSTRAP_CONFIG_PATH`（仅 local --empty）
   - `PLATFORM_BOOTSTRAP_*`（仅 saas）
3. `docker-compose down -v` → `up postgres` → bootstrap-local-after-reset → redis FLUSHALL
4. `pnpm -F @tokenjoy/backend dev-bootstrap`（Makefile source `.env.development` → 按新值 seed）
5. 写 `.env.local-mode`（供 start 读取 SUPPORT_SAAS）

### start.sh

1. 检查 `.env.local-mode` 存在
2. `set -a && source .env.local-mode`（export SUPPORT_SAAS）
3. ensure-infra + concurrently 启动 backend / frontend / mock

### 文件变更

| 文件 | 操作 |
|------|------|
| `scripts/dev/reset.sh` | 重写 |
| `scripts/dev/start.sh` | 新建 |
| `scripts/dev/start-local.sh` | 删除 |
| `scripts/dev/start-local-empty.sh` | 删除 |
| `scripts/dev/start-saas.sh` | 删除 |
| `scripts/dev.sh` | 重写 dispatcher |
| `package.json` | 精简 scripts |
| `.gitignore` | 添加 `.env.local-mode` |

---

## 迁移对照

| 旧命令 | 新等价 |
|--------|--------|
| `pnpm docker:reset` | `pnpm reset` |
| `pnpm start:local` | `pnpm reset local && pnpm start` |
| `pnpm start:local:empty` | `pnpm reset local --empty && pnpm start` |
| `pnpm start:saas` | `pnpm reset saas && pnpm start` |
| `pnpm start:lite` | `pnpm start:lite` |

日常：reset 一次，反复 `pnpm start`。换模式时重新 reset。
