# NewAPI Token 归属企业钱包用户

| | |
| --- | --- |
| 状态 | 方案讨论（未实施） |
| 关联 | [Backend.md](./Backend.md) §4.2、[Backend-架构.md](./Backend-架构.md)、[seed/FIXES.md](../apps/backend/seed/FIXES.md) FIX-SEED-004 |
| 上游钉扎 | `apps/newapi/patches/new-api/UPSTREAM_REF` |

## 1. 问题

### 1.1 现象

本地 / SaaS bootstrap 后：

| 事实 | 典型值 |
| --- | --- |
| Postgres `companies.newapi_wallet_user_id` | `2`（用户名 `company-2`） |
| NewAPI `tokens.user_id`（Platform Key 对应 Token） | `1`（root / admin） |

文档与产品意图要求：**NewAPIKey 的 `user_id` = 该企业的 `newapi_wallet_user_id`**（见 Backend.md：「NewAPIKey 创建（SaaS）：`user_id` = `newapi_wallet_user_id`」）。

现状与文档不一致：钱包配额同步打在企业用户上，Token 却挂在运营账号下。

### 1.2 根因（合同误解，非 seed 随机写错）

```
TokenJoy domain
  └─ CreateTokenInput.UserID = newapi_wallet_user_id   // 意图正确
       └─ HTTP JSON: { "user_id": 2, "name": "tokenjoy:plk-1", ... }
            └─ Headers: Authorization=Bearer <ADMIN_AT>, New-Api-User: 1
                 └─ NewAPI AddToken:
                      UserId: c.GetInt("id")   // 只取鉴权用户 = 1
                      // body.user_id 被忽略
```

要点：

1. Backend **已经**把企业钱包用户 id 放进请求体。
2. 调用身份始终是全局 `NEW_API_ADMIN_TOKEN`（运营用户，通常 id=1）；`New-Api-User` 必须与 AT 所属用户一致。
3. 上游 `AddToken`（钉扎版本）**不读取** body 中的 `user_id`，落库一律用当前鉴权 `id`。
4. 因此出现「钱包用户是 2、Token 属于 1」——可以跑通（Token `remain_quota` + root 额度），但多租户归属语义是错的。

这不是「忘了传企业 id」，而是 **传了但对方 API 不支持代建**。

### 1.3 为什么不能「直接改用企业 AT」

若要用企业用户身份调用 `POST /api/token/`，必须同时具备：

- 属于该企业钱包用户的 **Access Token**（不是 Platform Key `sk-…`）
- Header `New-Api-User` = 该用户 id

当前缺口：

- `CreateUser` 时密码是临时随机串，**未作为可再登录凭证持久化**（也不应明文长期存）。
- 上游没有稳妥的「Admin 代发任意用户 AT」合同；`GenerateAccessToken` 面向当前登录用户。
- 因此 sync / bootstrap **拿不到**企业用户的 Admin AT，只能一直用运营 AT。

「传企业 token」在现有凭证模型下不可行，除非另建 impersonation / 凭证体系。

---

## 2. 目标语义（SSOT）

与现有主文档对齐，修复后应满足：

1. **企业隔离**：每个企业对应唯一 `newapi_wallet_user_id`；该企业所有 Platform Key → NewAPI Token 的 `user_id` 等于该值。
2. **主账仍在 Postgres**：lot / `wallet_remain` / ledger 为金钱 SSOT；NewAPI `users.quota` 与 Token `remain_quota` 仍为派生/分配视图。
3. **wallet_sync**：继续按企业用户校准 NewAPI 用户配额；与 Token 所属用户一致，避免「改 A 的钱包、Key 挂在 B」。
4. **运维面**：Backend 仍可用一个运营 Admin Token 完成管理操作；不要求把企业密码塞进 Backend 配置。

非目标（本方案不展开）：

- 让企业终用户直接登录 NewAPI Console。
- 把 NewAPI `users.quota` 升格为金钱 SSOT。

---

## 3. 方案对比

### 方案 A — NewAPI 补丁：Admin 代建 Token 并指定 `user_id`（推荐）

**做法**

在 `apps/newapi/patches/` 扩展 `AddToken`（或新增 Admin 专用路由）：

- 当调用者 `role >= Admin`（或 Root）时，允许 body（或独立字段）指定目标 `user_id`。
- 校验：目标用户存在；调用者有权管理该用户（不可抬权到高于自己）。
- 落库：`Token.UserId = 指定的 user_id`（而非 `c.GetInt("id")`）。
- 普通用户调用行为不变（仍只能给自己建 Token）。

Backend 适配层保持现有 `CreateTokenInput.UserID`；去掉「假装 body 已生效」的错觉，改为依赖 **已补丁上游** 的合同，并用合同测试钉死。

**优点**

- 与 Backend.md / 计费文档意图一致。
- 不存企业密码、不做会话扮演。
- 运营 AT 模型不变，运维简单。
- wallet_sync 与 Token 归属同一 NewAPI 用户，语义闭合。

**缺点**

- 必须维护 fork；升级 `UPSTREAM_REF` 要 rebase。
- 需权限与审计（谁代谁建了 Token）。

**适用**：继续以「一企一 NewAPI 钱包用户」为多租户边界——与当前文档一致时的默认选择。

---

### 方案 B — 改文档 / 记账模型：承认 Token 挂在运营账号

**做法**

正式声明：

- NewAPI Token **统一属于**运营 Admin 用户（或共享技术账号）。
- 企业隔离 **只** 发生在 Postgres（公司、lot、mapping、`company_id`）。
- `newapi_wallet_user_id` 仅用于配额派生（若仍需要），或进一步弱化为可选。
- 删除 / 改写「Token.user_id = wallet user」的表述；rebalance / sync 文档写清「User quota vs Token remain」职责。

**优点**

- 对齐现状，改动小；上游无 fork。
- 实现简单。

**缺点**

- 与已写文档、SaaS 开户故事冲突，需全量改文档与心智。
- NewAPI 控制台侧按用户看 Token 时，所有企业混在运营账号下，运维辨识差。
- 若未来依赖 NewAPI 用户级限流/账单，隔离会变糊。

**适用**：产品明确放弃「Token 必须属企业用户」，只把 NewAPI 当执行网关。

---

### 方案 C — Impersonation：用企业用户身份调 Admin API

**做法**

CreateUser 后保存可轮换的凭据（或 root 代发 AT），建 Token / 改 Token 时切换 `Authorization` + `New-Api-User` 为企业钱包用户。

**优点**

- 不改 AddToken 归属逻辑，也能让 Token 落在企业用户下。

**缺点**

- 要设计密钥保管、轮换、泄露面（显著大于方案 A）。
- 每个企业一次登录/发 AT，失败重试与密钥失效复杂。
- 与「Admin Token 仅存 Backend 环境变量」的安全叙述冲突。

**结论**：不推荐作为主路径；仅当无法 fork NewAPI 时的权宜。

---

## 4. 推荐结论

| 优先级 | 方案 | 建议 |
| --- | --- | --- |
| **首选** | **A · Admin 代建 + 指定 user_id** | 与现有产品文档和租户模型一致，安全模型干净 |
| 备选 | B · 改文档接受挂运营账号 | 仅当产品否决「Token 属钱包用户」 |
| 不建议 | C · 企业 AT impersonation | 密钥与运维成本过高 |

**拍板问题（实施前需产品确认一句）**

> 企业隔离是否必须体现在 NewAPI `tokens.user_id` 上？  
> — 是 → 方案 A  
> — 否，只信 Postgres → 方案 B  

在未拍板前，**不要**再在业务层用「body 写了 user_id」当作已归属企业用户。

---

## 5. 方案 A 实施纲要（待批准后执行）

### 5.1 NewAPI patch

1. 扩展 `AddToken`（或 `POST /api/token/admin`）：Root/Admin 可读目标 `user_id`。
2. 权限：目标用户存在；`canManageTargetRole`；禁止给 Root 乱挂（按现有用户管理约束）。
3. 响应：至少返回 `id`（若继续空 data，Backend 已有按 `name` 查找的兜底，可保留但要测代建后的 list 过滤——**注意**：当前 `GET /api/token/` 只列 **当前用户** 的 Token；代建到 user=2 后，用 root AT **按 name 列举可能找不到**。  
   → patch 必须同时保证其一：  
   - Create 响应带 `data.id` / `data.key`，或  
   - Admin 可按 user / name 查询任意用户 Token。
4. 升级说明写入 `apps/newapi/patches/new-api/` 与 UPSTREAM_REF 变更记录。

### 5.2 Backend 适配层

1. `CreateToken` 合同测试：断言请求带目标 `user_id`，且创建结果 `Token.UserID == walletUserID`（对 mock/夹具或集成 NewAPI）。
2. Create 后校验：`token.UserID` 必须等于输入；否则失败并告警（防止再静默挂到 root）。
3. `findTokenByName`：若 list 仍按当前用户过滤，改为使用 Create 响应 id，或 Admin 搜索 API。
4. seed / bootstrap / SaaS `CreateCompany` 路径共用同一 adminport 语义，无分叉。

### 5.3 数据修复（一次性）

已错误挂在 root 下的 Token：

1. **推荐**：disable/delete 旧 Token → 按正确 `user_id` 重建 → 更新 `platform_key_mappings` + `key_hash`。  
2. 或 DB 层改 `tokens.user_id`（若上游无转移 API）——需同步清 Redis Token 缓存，风险较高，仅本地可接受。

本地可用 `make dev-bootstrap` / 映射重建流程覆盖；生产需迁移 runbook。

### 5.4 验收

| # | 条件 |
| --- | --- |
| 1 | 新建企业 Platform Key 后，NewAPI Token.`user_id` == `companies.newapi_wallet_user_id` |
| 2 | `wallet_sync` 调整的用户与 Token 所属用户为同一 id |
| 3 | 合同测试覆盖「Admin 代建 + 指定 user_id」；`UpstreamRef` 与 pin 文件一致 |
| 4 | 文档句「NewAPIKey 创建：user_id = newapi_wallet_user_id」与实现一致，无需改口 |
| 5 | 不引入企业密码 / 长期 AT 入库 |

---

## 6. 与近期相关修复的关系

下列问题已在适配层按上游合同修过或加固，**不代替**本归属问题：

| 主题 | 说明 |
| --- | --- |
| CreateUser 空 data | 用 search 解析真实 id，禁止写入 `newapi_wallet_user_id=0` |
| UpdateToken 整行覆盖 | GET→merge→PUT，避免 `expired_time=0` |
| TopUp 路径 | `/api/user/manage` `add_quota` |
| Channel create | `{mode:single, channel:{…}}`；abilities → `POST /api/channel/fix` |

Token **用户归属** 是独立的产品合同问题；修完 HTTP 形状后，若无方案 A/B 拍板，归属仍会错。

---

## 7. 决策记录（留空）

| 日期 | 决策 | 备注 |
| --- | --- | --- |
| | 待选 A / B | |

批准方案后，再改代码与测试；本文仅作方案基准。
