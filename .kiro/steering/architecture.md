# 架构

默认在 apps/。路径无前缀 = apps/。

- `apps/` — TokenJoy，客户侧 LLM API 管控平台。前端 React19+TailwindV4+TanStackQuery，后端 Go+PG+Redis。`SUPPORT_SAAS` flag 区分私有化/SaaS。
- `apps/web/` — 官网，纯展示 React+TailwindV3，无路由，iframe 嵌入 App 认证。
- `sms/` — 内部供应商管理系统，独立 Go module，制品不出门。
- apps/sms 互不 import，跨产品仅 HTTP API，共享契约放 `packages/contracts/`。
