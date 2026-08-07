---
inclusion: fileMatch
fileMatchPattern: "**/internal/app/**,**/internal/http/**,**/internal/domain/**,**/internal/worker/**"
---

# SQL 访问规则

- `internal/app/`、`internal/http/`、`internal/domain/`、`internal/worker/` 禁止直接写 INSERT/UPDATE/DELETE SQL
- 业务表写操作必须通过 `store/` repo 层（`store.BillingRepository`、`store.ModelsRepository` 等）
- `seed/` 层例外（DB 初始化工具，允许 raw SQL）
- `internal/app/setup_server.go` 现有 raw SQL 为历史遗留（标注了 ponytail 注释），后续应收口到 repo
- 新增写操作时，如果 store 实例不可用，必须在注释中标明对应的 repo 方法名，便于 schema 变更时同步更新
