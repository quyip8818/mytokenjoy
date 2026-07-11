# Backend 命名统一

命名约定（与 [Backend-架构.md](./Backend-架构.md) §0 一致）。

---

## 约定

| 项 | 取值 |
| -- | ---- |
| Gateway 开关 | `NEW_API_GATEWAY_ENABLED` → `GatewayEnabled` |
| SaaS 共享 group | `PLATFORM_SHARED_NEW_API_GROUP` |
| Deps | `Gateway`（类型 `GatewayService`） |
| 平台创建 ProviderKey | `CreateProviderKeyForPlatform` |
| org 本地结构 | `structure.LocalService` |
| Go 缩写 | `ID` / `API`（生成器缩写表） |
| Store 接口 / 实现文件 | `*_repo.go` |

领域词汇：`Gateway` / `NewAPISync` / `PlatformKey` / `ProviderKey` / `NewAPIKey` / `PlatformKeyMapping` / `AsyncJobs`。不用 Relay；领域不用 Token 指 Key（JWT/session 写全称；LLM `inputTokens` 与厂商 Admin API 字面量除外）。

---

## 明确不做

| 不做 | 原因 |
| ---- | ---- |
| 改 JSON tag / DB 列 / HTTP path | 对外契约 |
| `PlatformKey` → TokenJoyKey | §0 词汇 |
| domain → `me` | HTTP 门面 ≠ 领域包名 |
| `LogStore` → `LogRepository` | 日志库与主库刻意区分 |
| Overrun/Rebalance 接口与实现同词 | 端口按能力命名 |

---

## 接受的双名

| Domain | HTTP / 其它 |
| ------ | ----------- |
| `memberanalytics` | `handler/me`、`/api/me` |
| — | `LogStore`；`OverrunProcessor` / `Rebalancer` |
