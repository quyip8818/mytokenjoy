# 模型定价管理：三端架构

## 一句话

官方管理平台掌握上游通道和计费权，用 NewAPI 原生 UI 管理定价。Local/SaaS 用 TJ 自己的 UI 展示价格 + 管理自定义模型。Local 的 hack 不影响平台账单。

---

## 1. 三个版本

| 版本 | 说明 | 模型管理界面 |
|------|------|-------------|
| **官方管理平台** | 发布模型价格、掌握上游 API Key、按 token 计费 | **NewAPI 原生 admin UI** |
| **Local 版本** | 私有化部署给客户 | **TJ 自研 UI** |
| **SaaS 版本** | TJ 公有云多租户 | **TJ 自研 UI** |

---

## 2. 三端模型管理界面详解

### 2.1 官方管理平台 — NewAPI 原生 UI

**用谁的**：NewAPI 自带的 admin 后台，零开发。

**为什么**：平台网关本身就是一个 NewAPI 实例。NewAPI 原生 UI 已覆盖全部需求：
- 模型定价编辑（model_ratio / completion_ratio / cache_ratio）
- Channel 管理（接入 OpenAI / Anthropic / 各上游）
- 上游价格同步
- 模型目录管理
- 按次计费 / 按量计费切换

**访问方式**：`https://gateway-admin.tokenjoy.internal/`（内网，运营人员直接登录）

**后续扩展**：当需要平台独有功能（定价版本发布、实例管理、账单）时，做独立平台前端，定价编辑仍调 NewAPI `/api/option/` API。

```
运营人员 → NewAPI admin UI → 编辑 model_ratio
                                    ↓
                           各 Local 定时同步拿到新价格
```

---

### 2.2 Local 版本 — TJ 自研 UI

**用谁的**：TJ 前端（`features/models/` 已有基础，增强即可）

**为什么不用 NewAPI UI**：
- Local 客户不应该看到 NewAPI 的管理后台（安全 + 体验）
- 权限体系不同（TJ 有自己的 session/permission）
- 功能需求不同：内置模型只读 + 自定义模型可编辑

**功能矩阵**：

| 功能 | 内置模型 | 自定义模型 |
|------|---------|-----------|
| 查看列表 + 价格 | ✅（只读，同步自平台） | ✅ |
| 编辑价格 | ❌（平台 SOT） | ✅ |
| 添加/删除模型 | ❌ | ✅ |
| 配置 endpoint / API Key | ❌ | ✅ |
| 启停 | ✅ | ✅ |
| 配置路由/白名单 | ✅ | ✅ |

**现有代码基础**（已有，无需重写）：
- `features/models/` — 模型列表页、创建/编辑 workflow
- `model-list-table.tsx` — 表格展示
- `model-create.tsx` / `model-edit.tsx` — 自定义模型 CRUD
- `IsCustom()` / `isCustomModel()` — 区分内置/自定义

**需要新增的**：
- 内置模型的价格展示列（`ListModelsWithPricing`）
- 内置模型编辑时价格字段只读
- 同步状态指示器（版本号 + 上次同步时间）

---

### 2.3 SaaS 版本 — TJ 自研 UI

**用谁的**：与 Local 相同的 TJ 前端，但功能更少。

**功能矩阵**：

| 功能 | SaaS |
|------|------|
| 查看内置模型列表 + 价格 | ✅（只读） |
| 自定义模型 | ❌（SaaS 不支持自带模型） |
| 编辑价格 | ❌ |
| 同步状态 | ❌（不展示，平台直接管控） |

**现有代码已处理**：`isSelfHosted` 分支已经隐藏了 SaaS 不需要的功能。

---

## 3. 三端对比总览

| | 官方管理平台 | Local | SaaS |
|---|---|---|---|
| **界面** | NewAPI 原生 admin UI | TJ 自研 UI | TJ 自研 UI |
| **模型定价编辑** | ✅ 完整编辑 | ❌ 内置只读 / ✅ 自定义可编辑 | ❌ 全部只读 |
| **Channel 管理** | ✅（上游供应商接入） | ❌ 不暴露 | ❌ 不暴露 |
| **自定义模型** | — | ✅ 客户自己的模型 | ❌ |
| **开发成本** | 0（直接用 NewAPI UI） | 低（现有代码增强） | 0（复用 Local UI 的子集） |
| **用户** | TJ 运营团队 | Local 客户管理员 | SaaS 客户 |

---

## 4. 架构全景

```
┌─────────────────────────────────────────────────────────────────────┐
│                      官方管理平台                                      │
│                                                                     │
│  ┌────────────────────────┐    ┌──────────────────────────────┐    │
│  │ NewAPI 原生 admin UI    │    │ 平台网关 NewAPI                │    │
│  │ (运营人员操作定价)       │───▶│ - 持有上游 API Key            │    │
│  └────────────────────────┘    │ - model_ratio = 计费 SOT     │    │
│                                │ - 按 token 计量 → 出账单      │    │
│                                └──────────────▲───────────────┘    │
│                                               │                     │
└───────────────────────────────────────────────┼─────────────────────┘
                                                │
                  ┌─────────────────────────────┼──────────────────┐
                  │                             │                  │
    ┌─────────────┴──────────┐    ┌────────────┴───────────────┐
    │     Local 实例 A        │    │     Local 实例 B            │
    │                        │    │                            │
    │  TJ 前端 (自研 UI)      │    │  TJ 前端 (自研 UI)          │
    │  ├ 内置模型：只读价格   │    │  ├ 内置模型：只读价格       │
    │  ├ 自定义模型：可编辑   │    │  ├ 自定义模型：可编辑       │
    │  └ 同步状态             │    │  └ 同步状态                 │
    │                        │    │                            │
    │  TJ 后端               │    │  TJ 后端                   │
    │  ├ pricing syncer      │    │  ├ pricing syncer          │
    │  └ 定时拉取平台价格     │    │  └ 定时拉取平台价格         │
    │                        │    │                            │
    │  NewAPI (内网)          │    │  NewAPI (内网)              │
    │  └ channel → 平台网关   │    │  └ channel → 平台网关       │
    └────────────────────────┘    └────────────────────────────┘
```

---

## 5. 计费防 hack

**原则：计费权在平台，不在 Local。**

平台网关掌握上游 API Key。Local 的所有 AI 请求经过平台网关，平台按自己的 ratio 计费。

| Local 用户可能做的 | 平台账单影响 |
|-------------------|-------------|
| 改本地 NewAPI 的 model_ratio | ❌ 不影响 |
| 停止同步平台价格 | ❌ 不影响 |
| 修改 TJ 后端代码 | ❌ 不影响 |
| 用自己的 key 替换 channel | 流量不走平台 → 平台看到为 0 → 合同问题 |

不需要签名、防篡改、客户端校验。架构本身保证计费在平台侧。

---

## 6. 定价同步（Local）

Local 定时同步平台价格，用于本地展示和内部预算管理。

```go
// 每 10 分钟
func (s *Syncer) syncOnce(ctx context.Context) {
    latest, err := s.platform.GetLatestPricing(ctx)
    if err != nil { slog.Warn(...); return }
    if latest.Version == s.lastVersion { return }
    
    s.adminport.UpdateOption(ctx, "ModelRatio", marshal(latest.ModelRatio))
    s.adminport.UpdateOption(ctx, "CompletionRatio", marshal(latest.CompletionRatio))
    // ...
    s.lastVersion = latest.Version
}
```

同步失败不影响核心功能 — Local 继续用旧价格展示，平台计费不受影响。

---

## 7. 实施步骤

| 步骤 | 说明 | 工作量 |
|------|------|--------|
| 1. 平台：部署网关 NewAPI | 配置上游 channel，运营人员用自带 UI 管理定价 | 运维配置 |
| 2. Local：NewAPI channel 指向平台网关 | 部署配置变更 | 运维配置 |
| 3. 平台：暴露 `GET /api/v1/pricing/latest` | Local syncer 的数据源 | 小 |
| 4. Local 后端：pricing syncer | 定时拉取 → 写本地 NewAPI option | 小 |
| 5. Local 后端：adminport 扩展 | GetOption / UpdateOption | 小 |
| 6. Local 前端：模型列表增强 | 内置模型价格列 + 同步状态 | 小 |
| 7. 平台：账单模块 | 从网关 log 按实例汇总出账 | 中 |

---

## 8. 决策记录

| 决策 | 理由 |
|------|------|
| 平台模型管理用 NewAPI 原生 UI | 功能完全匹配，零开发成本 |
| 客户侧用 TJ 自研 UI | 权限、体验、安全需要自控 |
| 不复用 NewAPI UI 给客户 | 暴露 NewAPI 管理后台不安全且体验割裂 |
| 计费在平台网关侧 | 架构保证，无需防 hack |
| Local 同步是强制的 | 保证本地展示与平台一致（但不影响计费） |
| 自定义模型价格 Local 可编辑 | 客户自己的模型，平台不管 |
