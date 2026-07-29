# Wallet 同步机制

TokenJoy 是钱包余额的唯一真实来源（SOT）。NewAPI 的 user wallet 是它的实时镜像。

## 核心原则

```
TokenJoy wallet_remain_quota = SOT（唯一写入点）
NewAPI user.quota            = 镜像（每次变更后 override）
```

每次 TokenJoy 的 `wallet_remain_quota` 发生变更，都会在事务提交后 best-effort 调用 NewAPI `set_quota`（mode=override）把绝对值覆盖过去。

---

## 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                          TokenJoy                                │
│                                                                 │
│  ┌──────────────┐         ┌──────────────────────┐              │
│  │ 充值路径      │         │ 消费路径（Ingest）    │              │
│  │              │         │                      │              │
│  │ CreditFromLot│         │ ConsumeLotsLocked    │              │
│  │ (billing svc)│         │ (ingest service)     │              │
│  └──────┬───────┘         └──────────┬───────────┘              │
│         │                            │                          │
│         │  TX: ApplyWalletDelta      │  TX: SetWalletRemainQuota│
│         │                            │                          │
│         ▼                            ▼                          │
│  ┌─────────────────────────────────────────────┐                │
│  │        companies.wallet_remain_quota         │  ← SOT        │
│  └─────────────────────────────────────────────┘                │
│         │                            │                          │
│         │  Post-commit (best-effort) │                          │
│         ▼                            ▼                          │
│  ┌─────────────────────────────────────────────┐                │
│  │    ManageUser("set_quota", value)            │                │
│  │    mode: "override"                          │                │
│  └──────────────────────┬──────────────────────┘                │
│                         │                                       │
└─────────────────────────┼───────────────────────────────────────┘
                          │ HTTP POST /api/user/manage
                          ▼
              ┌───────────────────────┐
              │       NewAPI          │
              │   user.quota = value  │  ← 纯镜像
              └───────────────────────┘
```

---

## 两条写入路径

### 1. 充值路径

```
用户充值 → billing.CreditFromLot()
        → TX: insert lot + ApplyWalletDelta
        → TX commit
        → syncWalletBestEffort: set_quota(新余额)
```

触发时机：`PlatformRecharge` / `PlatformGift` / `PlatformAdjust` / `ConfirmPayment`

### 2. 消费路径

```
NewAPI consume log → IngestRaw()
        → TX: LockForUpdate + ConsumeLotsLocked + SetWalletRemainQuota
        → TX commit
        → post-commit: set_quota(新余额)
```

触发时机：每条 consume log 入账时

### 3. 升级路径（边缘）

```
UpgradeToStandard → TX: ExpireMockLots + SetWalletRemainQuota
                  → TX commit
                  → set_quota(新余额)
```

---

## 失败处理

| 场景 | 行为 |
|------|------|
| NewAPI 不可达 | warn log，不阻塞主流程 |
| 多次 override 同一个值 | 幂等，无副作用 |
| 两次 ingest 并发（同一 company） | company row lock 保证串行，override 有序 |
| NewAPI wallet 被自身消费扣减 | 下一次 ingest override 会覆盖回来 |

**偏差方向始终安全：** NewAPI wallet 可能暂时偏高（用户能多用一点），不会出现"有钱但被拒"的情况。

---

## NewAPI API 调用格式

```json
POST /api/user/manage
{
  "id": <walletUserID>,
  "action": "add_quota",
  "mode": "override",
  "value": <wallet_remain_quota 绝对值>
}
```

注意：NewAPI 的 action 固定是 `"add_quota"`，通过 `mode` 字段区分语义：
- `"add"` — 增量
- `"override"` — 覆盖为绝对值

---

## 代码位置

| 模块 | 文件 | 职责 |
|------|------|------|
| ManageUser client | `integration/newapi/user.go` | HTTP 调用封装，set_quota → mode=override |
| 充值同步 | `domain/billing/lot_confirm.go` | `syncWalletBestEffort` |
| 消费同步 | `domain/usage/ingest.go` | post-commit block |
| 升级同步 | `domain/company/service.go` | `UpgradeToStandard` post-commit |
| SOT 写入 | `store/postgres/company_repo.go` | `ApplyWalletDelta` / `SetWalletRemainQuota` |
