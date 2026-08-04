# Currency 变更记录（轻量方案）

## 目标

在 currencies 表上加 `updated_by_user_id`，表格里直接看到「谁最后改的」。不加 history 表，不加额外 API。

---

## 1. 数据模型

currencies 表新增一列：

```sql
ALTER TABLE currencies
    ADD COLUMN updated_by_user_id UUID;
```

- `updated_by_user_id`：操作人 user.id，前端通过 JOIN users 表展示名字

---

## 2. 后端改动

### 2.1 Store

`Currency` struct 加字段：

```go
type Currency struct {
    Code            string
    QuotaPerUnit    int64
    Enabled         bool
    UpdatedAt       time.Time
    UpdatedByUserID *uuid.UUID
}
```

所有写操作的 SQL 加上 `updated_by_user_id = $x`。

### 2.2 Handler

每个写操作（Create / Update / ToggleStatus）从 session 取操作人：

```go
session, _ := httpx.SessionFromContext(r.Context())
actorUserID := session.Member.UserID
```

传入 store 层写入。

### 2.3 CatalogSync

CatalogSync 的 upsert SQL **不涉及 `updated_by_user_id` 列**（不 SET、不覆盖）。该列仅由 platform admin 手动操作时写入。

`updated_at` 同理：sync 时保留 platform 端的原始值，不用 `NOW()` 覆盖。Local 看到的时间就是 SaaS 管理员实际操作的时间。

原因：
- SaaS 实例不跑 CatalogSync（它是 SOT），不存在覆盖问题
- Local 实例跑 CatalogSync 但不展示 platform currencies 管理页面（无 `platform:manage` 权限），`updated_by_user_id` 对 Local 无意义

### 2.4 API 响应

`currencyResponse` 新增字段：

```json
{
  "code": "CNY",
  "quotaPerUnit": 500000,
  "enabled": true,
  "updatedAt": "2025-01-15T10:30:00Z",
  "updatedByUserID": "uuid...",
  "updatedByName": "张三"
}
```

handler 读取时 JOIN users 表拿 name，返回给前端。不需要新 endpoint，现有 `GET /api/platform/currencies` 返回里直接带上。

---

## 3. 前端

### 3.1 类型

```ts
export interface PlatformCurrency {
  code: string
  quotaPerUnit: number
  enabled: boolean
  updatedAt: string
  updatedByName: string | null  // 新增，后端 JOIN 返回
}
```

### 3.2 表格

| 币种代码 | Quota/单位 | 状态 | 修改人 | 修改时间 | 操作 |
|---------|-----------|------|--------|---------|------|
| CNY | 500,000 | 启用 | 张三 | 01-15 10:30 | 编辑 · 启停 |
| USD | 70,000 | 启用 | — | 01-10 09:00 | 编辑 · 启停 |

- 「修改人」列：显示 `updatedByName`，null 则显示 `—`（系统同步或初始化）
- 「修改时间」列：相对时间或短格式，hover 显示完整时间戳

---

## 4. 查账够用吗？

当前方案只存最后一次修改人，不存历史。够用的场景：

- ✅ 知道当前 QPU 是谁设的
- ✅ lot 快照了 QPU，和当前值对比就知道是否变过
- ✅ 币种数量少（<10），变更频率极低

不够用的场景（未来可加 history 表升级）：

- ❌ 想看「QPU 从 X 改成 Y 是谁在几号做的」完整链条
- ❌ 多次修改后想回溯中间状态

> ponytail: 当前天花板是只存最后修改人。如果未来需要完整审计链，加 `currency_history` 表 + insert trigger 即可，不影响现有字段。

---

## 5. 工作量

| 步骤 | 内容 |
|------|------|
| 1 | schema 加一列 |
| 2 | store scan 补字段 |
| 3 | handler 4 个写操作传 actor user_id |
| 4 | list 查询 JOIN users 拿 name |
| 5 | 前端表格加列 |

半天搞定。
