# 组织架构同步 — 设计文档

> 状态：设计完成  
> 日期：2026-07-23  
> 范围：字段级同步策略 + 日常同步 + 换源迁移

---

## 1. 核心概念

### 1.1 字段同步模式

系统中每个成员字段属于且仅属于以下四种模式之一：

| 模式 | 含义 | 首次导入 | 后续同步 | 谁能改 |
|------|------|---------|---------|--------|
| **immutable** | 一次性写入 | 从源写入 | 不更新 | 仅 Admin |
| **user-owned** | 用户自管 | 从源写入 | 用户没改过→覆盖；改过→跳过 | 用户本人 |
| **sync-always** | 源始终权威 | 从源写入 | 始终覆盖 | 不可本地改 |
| **local-only** | 纯本地 | 不从源获取 | 不从源获取 | Admin |

### 1.2 字段分类表（Member 级）

| 字段 | 模式 | 说明 |
|------|------|------|
| EmployeeID | immutable | 工号，入职时确定，终身不变 |
| HireDate | immutable | 入职日期，同上 |
| Alias（显示名） | user-owned | 用户可能设置昵称/别名 |
| Avatar | user-owned | 用户可能自行更换头像 |
| JobTitle | sync-always | 职位由 HR 系统管理，本地不可改 |
| DepartmentID | sync-always | 部门归属由 HR 系统管理 |
| DepartmentName | sync-always | 跟随 DepartmentID |
| Roles | local-only | 本系统内的权限角色 |
| Status | local-only | 本系统内的启停状态 |
| PersonalBudget | local-only | 预算分配 |
| API Keys | local-only | 用户的密钥 |

> **关于 Phone/Email：** 这两个字段属于 User 表而非 Member 表。同步时仅用于定位或创建 User（见 1.4），不参与 OverrideFields 追踪。
>
> **关于 Manager：** 汇报关系是部门级属性（Department node 的 ManagerID），由部门同步逻辑处理，不在 Member 字段同步策略范围内。

### 1.3 OverrideFields

每个 Member 记录维护一个 `OverrideFields []string`，记录被用户/Admin 手动修改过的 user-owned 字段名。

- 字段未出现在列表中 → 同步可以覆盖
- 字段出现在列表中 → 同步跳过
- 用户/Admin 可执行"恢复同步"操作（从列表中移除，下次同步时远端值会覆盖回来）

### 1.4 User 解析逻辑（同步时创建/复用用户）

同步导入新成员时，通过 Phone/Email 定位或创建 User record：

1. 按 Phone 查找 users 表 → 找到则复用该 userID，**不更新** user 的 phone/email
2. Phone 未命中 → 按 Email 查找 → 找到则复用，**不更新**
3. 都未命中 → 创建新 User，写入 phone + email + name
4. Phone 和 Email 都为空 → 该成员导入失败，记录到 failures

Phone/Email 一旦写入 User 表后，由用户自己通过账户设置修改，同步流程不再覆盖。

---

## 2. 日常同步流程

### 2.1 同步决策逻辑

```
对每个远端成员 remote：
│
├─ 本地不存在（首次导入）
│   → 创建 Member，所有非 local-only 字段从 remote 写入
│   → OverrideFields = []
│
└─ 本地已存在 existing：
    │
    ├─ existing.Source == "manual"
    │   → 完全跳过（手动创建的成员不受同步影响）
    │
    └─ existing.Source == "imported"
        → 逐字段判断：

        immutable 字段：
          本地为空 → 写入
          本地有值 → 跳过

        user-owned 字段：
          字段 ∈ OverrideFields → 跳过
          字段 ∉ OverrideFields → 覆盖

        sync-always 字段：
          → 直接覆盖

        local-only 字段：
          → 不处理
```

### 2.2 本地修改时的追踪

| 操作者 | 修改 immutable 字段 | 修改 user-owned 字段 | 修改 sync-always 字段 |
|--------|-------------------|--------------------|--------------------|
| 用户本人 | 不允许 | 允许，自动加入 OverrideFields | 不允许（UI 不可编辑） |
| Admin | 允许 | 允许，自动加入 OverrideFields | 不允许（UI 不可编辑） |
| 同步 | 按规则 | 按规则 | 直接覆盖 |

### 2.3 "恢复同步值"操作

用户或 Admin 对 user-owned 字段可执行"恢复为同步值"：
1. 从 OverrideFields 中移除该字段
2. 下次同步时该字段会被远端值覆盖
3. UI 上显示"将在下次同步后恢复为 {platform} 中的值"

---

## 3. 换源迁移

### 3.1 场景

公司从飞书切换到钉钉（或反向，或切换到 WeCom/自建 LDAP）。两套系统的用户 ID 体系完全不同，需要把新源的人和本地已有成员对齐。

### 3.2 产品流程

换源是一个 **Admin 操作向导**，分四步：

---

#### Step 1：连接新数据源

Admin 配置新源的凭据，系统验证连通性。

- 输入：新平台类型 + 凭据
- 输出：连接成功/失败
- UI：与现有"数据源配置"页面复用

此时旧源仍生效，不中断现有同步。

---

#### Step 2：预览与自动匹配

系统拉取新源全量人员，与本地成员做自动匹配，生成匹配报告。

**匹配优先级（从高到低）：**

| 优先级 | 匹配依据 | 置信度 | 处理方式 |
|--------|---------|--------|---------|
| 1 | EmployeeID（工号） | 高 | 自动匹配 |
| 2 | Phone（手机号，归一化后） | 高 | 自动匹配 |
| 3 | Email（不区分大小写） | 高 | 自动匹配 |
| 4 | Name + Department | 低 | 需人工确认 |
| — | 无法匹配 | — | 需人工处理 |

> **Phone 归一化规则：** 去除 `+`、`-`、空格，取最后 11 位数字。例如 `+86-138-0000-1234` → `13800001234`。

**匹配规则：**
- 一个本地成员只能匹配一个远端人员（1:1）
- 如果多个远端人员匹配同一个本地成员，全部进入人工确认
- 匹配只考虑 Source=imported 的本地成员（manual 成员不参与）

**输出：匹配报告**

```
匹配报告：
├─ ✅ 自动匹配 (N 人)     — 高置信，可批量确认
├─ ⚠️ 待确认 (M 人)       — 低置信或冲突，需逐个处理
├─ ❌ 远端未匹配 (X 人)    — 新源有、本地没有
└─ 📌 本地未匹配 (Y 人)    — 本地有、新源没有
```

---

#### Step 3：人工审核与调整

Admin 在审核页面处理所有非自动匹配的情况：

**对"待确认"的人：**
- 确认匹配（接受系统建议）
- 手动指定另一个本地成员
- 标记为新增（本地没有此人，需要导入）

**对"远端未匹配"的人：**
- 手动关联到某个本地成员
- 标记为新增成员（走正常导入流程）
- 忽略（不导入此人）

**对"本地未匹配"的人：**
- 保留为本地管理（Source 改为 manual，后续同步不碰）
- 标记为已离职（Status 改为 inactive）
- 不处理（等后续手动处理）

**OverrideFields 处理选项：**

Admin 在确认匹配时可选择（全局或逐人）：

| 选项 | 行为 | 适用场景 |
|------|------|---------|
| 清空 OverrideFields（默认） | 新源数据刷入一次 user-owned 字段 | 换源是因为旧源数据不准 |
| 保留 OverrideFields | 用户手动修改的值不被新源覆盖 | 只是换平台，数据本身没变 |

---

#### Step 4：执行迁移

Admin 确认所有匹配后，点击"执行迁移"：

1. **停止旧源同步**（禁用定时任务）
2. **更新已匹配成员：**
   - ExternalID → 新源 ID
   - Source platform → 新平台
   - OverrideFields → 按 Admin 选择清空或保留
3. **处理新增成员：** 走正常 importFromProvider 流程
4. **处理本地未匹配：** 按 Admin 选择修改 Source 或 Status
5. **更新 OrgIntegration：** 平台、凭据切换为新源
6. **启动新源同步**（立即触发一次全量同步）
7. **记录迁移日志**（SyncLog type=migration）

**数据绑定是原子操作** — 步骤 2-4 在同一个数据库事务内执行，要么全部成功，要么全部回滚。步骤 1（停止旧源）和步骤 5-6（启动新源同步）在事务外执行：事务成功后停止旧调度、启动新调度；事务失败则旧源继续生效。

---

### 3.3 迁移状态机

```
idle → preparing → reviewing → executing → completed
                      │                        │
                      ▼                        ▼
                   cancelled                 failed → reviewing (可重试)
```

- `preparing`：系统在拉取新源数据并自动匹配
- `reviewing`：等待 Admin 审核（可保存草稿，不限时间）
- `executing`：正在执行绑定
- `completed`：迁移完成，新源生效

### 3.4 安全保障

| 风险 | 防护 |
|------|------|
| 误删大量成员 | 迁移流程不自动删除任何人，只改 Source/Status |
| 匹配错误 | 高置信也需 Admin 确认（至少"批量确认"一次） |
| 迁移中途失败 | 事务回滚，旧源继续生效 |
| 迁移后发现问题 | SyncLog 记录迁移前快照，支持人工排查 |
| API Keys / Budget 丢失 | 迁移只改 ExternalID/Source，不碰业务数据 |

### 3.5 迁移期间的系统行为

| 状态 | 旧源同步 | 新源同步 | 本地操作 |
|------|---------|---------|---------|
| preparing | 正常 | 不触发 | 正常 |
| reviewing | 暂停 | 不触发 | 正常 |
| executing | 已停止 | 执行中 | 锁定成员管理 |
| completed | 已停止 | 正常 | 正常 |

---

## 4. 数据模型

### 4.1 Member 变更

```diff
 type Member struct {
     ID             uuid.UUID
     UserID         uuid.UUID
     Alias          string
     Username       string
     EmployeeID     string
     JobTitle       string
     HireDate       string
     DepartmentID   uuid.UUID
     DepartmentName string
     Status         string
     Roles          []string
     Source         string
     ExternalID     *string
+    OverrideFields []string `json:"overrideFields,omitempty"`
 }
```

### 4.2 字段同步策略定义

```go
var FieldSyncPolicy = map[string]FieldSyncMode{
    "employeeId":      SyncModeImmutable,
    "hireDate":        SyncModeImmutable,
    "alias":           SyncModeUserOwned,
    "avatar":          SyncModeUserOwned,
    "jobTitle":        SyncModeAlways,
    "departmentId":    SyncModeAlways,
    "departmentName":  SyncModeAlways,
    "roles":           SyncModeLocalOnly,
    "status":          SyncModeLocalOnly,
    "personalBudget":  SyncModeLocalOnly,
}
```

### 4.3 迁移记录

```go
type SourceMigration struct {
    ID            uuid.UUID
    OldPlatform   Platform
    NewPlatform   Platform
    Status        string    // preparing, reviewing, executing, completed, cancelled, failed
    MatchReport   MatchReport
    AdminDecisions []MatchDecision
    CreatedAt     time.Time
    CompletedAt   *time.Time
}

type MatchReport struct {
    AutoMatched []MatchPair   // 高置信自动匹配
    NeedReview  []MatchPair   // 低置信需确认
    RemoteOnly  []RemoteMember // 新源有、本地无
    LocalOnly   []Member       // 本地有、新源无
}

type MatchPair struct {
    Remote      RemoteMember
    Local       Member
    MatchedBy   string  // "employee_id", "phone", "email", "name_dept"
    Confidence  string  // "high", "low"
}

type MatchDecision struct {
    RemoteExternalID string
    Action           string    // "bind", "create", "ignore"
    LocalMemberID    *uuid.UUID // bind 时指定
    KeepOverrides    bool       // 是否保留 OverrideFields
}
```

---

## 5. API 设计

### 5.1 日常同步（已有，需调整）

```
POST /api/org/sync/trigger         — 触发同步（内部逻辑加入 shouldSyncField）
GET  /api/org/sync/logs            — 同步日志
```

### 5.2 恢复同步值

```
DELETE /api/org/members/:id/overrides/:field   — 移除某字段的 override 标记
```

### 5.3 换源迁移

```
POST   /api/org/migration/prepare              — 连接新源 + 自动匹配，返回匹配报告
GET    /api/org/migration/current              — 获取当前迁移状态与报告
PUT    /api/org/migration/decisions            — Admin 提交审核决策
POST   /api/org/migration/execute              — 执行迁移
DELETE /api/org/migration/current              — 取消迁移
```

---

## 6. 前端页面

### 6.1 成员详情页 — 字段状态展示

| 字段类型 | UI 表现 |
|---------|---------|
| immutable（有值） | 灰色只读，hover 提示"此字段不可修改" |
| immutable（无值） | Admin 可编辑的输入框 |
| user-owned（未 override） | 可编辑，标签"同步中" |
| user-owned（已 override） | 可编辑，标签"已自定义"，旁边有"恢复同步"按钮 |
| sync-always | 灰色只读，标签"由 {飞书/钉钉} 管理" |
| local-only | 正常编辑，无特殊标记 |

### 6.2 换源向导页

四步向导，对应 Step 1-4：

**Step 1 — 配置新源**
- 选择平台（飞书/钉钉/WeCom/LDAP）
- 输入凭据
- 测试连接
- "下一步"

**Step 2 — 匹配预览**
- 加载中状态 → 自动匹配完成
- 展示匹配统计卡片（自动匹配 N 人 / 待确认 M 人 / ...）
- "下一步"进入审核

**Step 3 — 审核匹配**

三个 Tab：

Tab 1: 自动匹配
```
┌─────────────────────────────────────────────────────┐
│ ✅ 自动匹配 (32人)                                   │
│                                                     │
│ 全局设置: [清空自定义字段 ▼]                          │
│                                                     │
│ ┌─────────┬──────────┬────────────┬───────┐         │
│ │ 新源人员 │ 匹配本地  │ 匹配依据    │ 操作  │         │
│ ├─────────┼──────────┼────────────┼───────┤         │
│ │ 张三    │ 张三      │ 工号 A001  │ [✓][✗]│         │
│ │ 李四    │ 李四      │ 手机号     │ [✓][✗]│         │
│ └─────────┴──────────┴────────────┴───────┘         │
│                                                     │
│ [全部确认]                                           │
└─────────────────────────────────────────────────────┘
```

Tab 2: 待确认
```
┌─────────────────────────────────────────────────────┐
│ ⚠️ 待确认 (3人)                                      │
│                                                     │
│ 王五 (研发部, 新源)                                   │
│ 系统建议: → 王五 (研发部, 本地)  依据: 姓名+部门      │
│ [确认匹配] [手动选择本地成员 ▼] [标记为新增]          │
│                                                     │
│ 赵六 (产品部, 新源)                                   │
│ 系统建议: 无                                         │
│ [手动选择本地成员 ▼] [标记为新增] [忽略]              │
└─────────────────────────────────────────────────────┘
```

Tab 3: 本地未匹配
```
┌─────────────────────────────────────────────────────┐
│ 📌 本地有但新源没有 (2人)                             │
│                                                     │
│ 孙七 — 原飞书导入                                    │
│ [转为本地管理] [标记离职]                             │
│                                                     │
│ 周八 — 原飞书导入                                    │
│ [转为本地管理] [标记离职]                             │
└─────────────────────────────────────────────────────┘
```

**Step 4 — 确认执行**
- 展示操作摘要（将绑定 N 人、新增 M 人、转本地 X 人）
- 警告文案："执行后旧数据源将断开，此操作不可撤销"
- [执行迁移] 按钮
- 执行中展示进度
- 完成后跳转到同步日志页

---

## 7. 实施优先级

| 阶段 | 内容 | 改动量 |
|------|------|--------|
| **P0** | Member 加 OverrideFields + FieldSyncPolicy 定义 | ~30 行 |
| **P0** | importFromProvider 改用 shouldSyncField 判断 | ~40 行 |
| **P1** | UpdateMember 自动追踪 user-owned 变更 | ~20 行 |
| **P1** | 前端字段状态标记 + "恢复同步"按钮 | 1 个组件 |
| **P2** | 换源迁移后端（匹配 + 执行） | 1 个 service |
| **P2** | 换源向导前端 | 1 个页面 |

P0 是日常同步的基础能力，约 70 行后端改动。  
P2 是换源能力，等有实际换源需求时再做。
