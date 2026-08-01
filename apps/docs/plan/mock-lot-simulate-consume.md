# Mock Lot 与模拟消耗方案 — 剩余待办

> 相关：[Backend-预算.md](../Backend-预算.md) · [Backend-Ingest架构.md](../Backend-Ingest架构.md)
>
> **现状：核心方案已实现。** demo/trial/testing 账户可通过「模拟消耗」体验完整的调用→扣费→看板链路；Gateway 按 `CompanyType` 门控 test-model 准入（`gateway_service.go` 的 `isTestModelAllowed`）；`GET /keys/platform/{id}/simulate-bearer` 正式路由已上线（`handler/keys/handler.go` 的 `SimulateBearer` + `isSimulateAllowed`）；充值入口已限制 trial/demo；升级流程 `company.Service.UpgradeToStandard` 已实现（事务内改类型 + `ExpireMockLots` 过期 mock lot + 重算 wallet，事务提交后 invalidate precheck cache）。本文档只保留尚未落地的部分。

## 剩余待办

### 1. trial/demo 反向拦截：白名单被篡改后仍可能命中真实模型

`gateway_service.go` 已有 `isTestModelAllowed(companyType)` 拦截 test-model 被非 demo/trial/testing 账户调用。但**反向规则未实现**：trial/demo 账户如果通过 API 手动把真实模型加入 Key 白名单，Gateway 目前不会额外拦截（现有的模型白名单校验会放行，因为该模型确实在白名单里）。

需要补一条 Gateway 层规则：`companyType ∈ {trial, demo} && model != test-model → reject`，确保 trial/demo 账户物理上只能调用 test-model，不依赖白名单配置是否正确。

### 2. trial/demo 创建 Key 时白名单默认仅 test-model

当前创建 Platform Key 时，UI 和 API 层未对 trial/demo 账户强制将可选模型范围收窄到 test-model。建议在 Key 创建校验（`domain/keys/platform_key_create.go`）中补充：`companyType ∈ {trial, demo}` 时 `ModelWhitelist` 只能包含 test-model。

## 未来可选扩展（低优先级）

- **testing 环境 lot kind 过滤**：`ConsumeLotsLocked` 目前 mock/real lot 混合 FIFO，testing 环境如需严格隔离可加 `callType` 参数区分（`LotConsumer.ConsumeLotsLocked(ctx, st, co, amount, callType)`）。
- Mock Lot 续费（平台运营手动追加）
- 模拟消耗支持更多 mock 模型（test-embedding 等）
- 用量看板区分 mock/real 数据标签
- Trial 倒计时（天数 + 额度双触发升级提醒）
- 升级竞态完全防护：升级事务提交与 precheck cache invalidate 之间的极短窗口内，in-flight test-model 请求可能在 ingest 阶段因 mock lot 已过期走 overdraft（金额极小，可接受，可按 `model=test-model` 事后筛选清理）
