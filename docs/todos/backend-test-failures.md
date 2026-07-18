# Backend 测试失败记录

> **状态：✅ 全部已修复**（2026-07-18）

## 背景

UUID v7 迁移期间曾出现 46 个测试失败（分布在 17 个 package），主要模式：

1. 测试 fixture 传入旧格式 ID（`"dept-5"`、`"approval-1"` 等非 UUID 值）
2. Gateway precheck allowlist 逻辑依赖旧 int64 model ID
3. Org sync authz_revision bump 调用路径变更
4. Handler 层新增 UUID 格式校验但测试未同步

## 当前状态

- `go test -tags testhook ./...` — **全部通过**（个别 flaky 为 DB 并发竞争，与 UUID 迁移无关）
- `pnpm test`（前端）— **47 files / 156 tests 全部通过**

所有修复已合入 main。此文档仅作历史记录。
