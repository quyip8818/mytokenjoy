# 组织模块代码审查报告

**审查日期:** 2026-07-07
**审查范围:** 数据源、组织架构、角色管理三大模块
**涉及路径:** `apps/frontend/src/features/org/`, `apps/backend/internal/domain/org/`

---

## 总体评估

三个模块共发现 **6 个高危问题、14 个中危问题、12 个低危问题**。

核心风险集中在两个模式缺陷上:

1. **部分更新 vs 全量替换不匹配** — 前端发送部分字段，后端按全量结构体覆盖，导致字段丢失（影响组织架构模块，会清除成员角色和状态）
2. **权限保护缺口** — 角色管理模块的 `AddRoleMember` 和 `UpdateRole` 缺少对预设角色的保护，允许权限提升攻击

数据源模块的主要问题是"测试字段映射"端点返回硬编码假数据，使向导流程的验证步骤形同虚设。

---

## 一、数据源模块

### 高危

#### 1. 字段映射测试接口返回硬编码假数据

- **文件:** `apps/backend/internal/domain/org/remote/field_mappings.go` (L56-84)
- **影响:** 前端依赖此接口的返回结果来决定"下一步"按钮是否可用。由于后端始终返回成功和假数据，用户即使配置了错误的映射也会收到"测试通过"的误导，直到实际同步时才会发现问题。
- **建议修复:** 使用存储的字段映射对远端 API 发起真实的小范围查询（如取 1 条记录），返回实际转换结果。添加错误场景的响应结构。

#### 2. SyncConfig 缺少输入验证，负阈值绕过安全保护

- **文件:** `apps/backend/internal/domain/org/remote/sync.go` (L174-181)
- **影响:** `FrequencyHours` 可为负数；`DeleteMemberThreshold` / `DeleteDepartmentThreshold` 为负时，阈值判断 `len(diff) > threshold` 恒为真，相当于禁用同步安全保护。攻击者或误操作可导致全量删除不被拦截。
- **建议修复:** 在 `UpdateSyncConfig` 入口添加验证：频率 >= 1，阈值 >= 0，StartTime 匹配 HH:MM 格式。

### 中危

#### 3. 同步表单遗漏 notifyIm 字段

- **文件:** `apps/frontend/src/features/org/components/data-source/step-sync-schedule.tsx`
- **影响:** 每次保存配置时 notifyIm 被重置为 false，用户无法开启 IM 通知。
- **建议修复:** 在表单中添加 notifyIm 开关控件，并纳入 defaultValues。

#### 4. 字段映射保存失败无反馈

- **文件:** `apps/frontend/src/features/org/components/data-source/step-field-mapping.tsx` (L97-105)
- **影响:** `saveFieldMappings` 抛异常时用户无任何提示，按钮恢复可用但数据未保存。
- **建议修复:** 在 catch 块中调用 `toast.error()`，与 StepSyncSchedule 保持一致。

#### 5. 字段映射空数组无空状态提示

- **文件:** `apps/frontend/src/features/org/components/data-source/step-field-mapping.tsx`
- **影响:** 新环境中返回空映射时，用户卡在无行、无提示的表格上，无法前进。
- **建议修复:** 添加空状态 UI，引导用户手动添加映射或从模板初始化。

#### 6. 导入事务后的操作失败导致状态不一致

- **文件:** `apps/backend/internal/domain/org/remote/import.go` (L247-254)
- **影响:** 主导入事务已提交，但 `SetImportFailures` 或 `EnqueueModelLimitsForDepartments` 失败时函数返回 error，调用方以为导入失败但数据已写入。
- **建议修复:** 将后续操作纳入同一事务，或改为 best-effort 并记录日志而非返回 error。

#### 7. 阈值为 0 时阻止所有删除操作

- **文件:** `apps/backend/internal/domain/org/remote/sync.go` (L123-134)
- **影响:** 默认零值配置下，任何涉及删除的同步都被拦截。用户困惑于"同步成功但成员没变化"。
- **建议修复:** 将阈值语义改为 `>=` 或使用 -1 表示"不限制"，并在前端明确标注含义。

#### 8. 平台切换未清除旧字段映射

- **文件:** `apps/backend/internal/domain/org/remote/datasource.go` (L68-72)
- **影响:** 从飞书切换到钉钉后，旧的飞书映射仍存在且通过平台匹配检查，导致字段语义错位。
- **建议修复:** 在 `UpdateDataSource` 的 force 分支中调用 `ClearFieldMappings()`。

### 低危

#### 9. 同步日志全量加载无分页

- **文件:** `apps/backend/internal/domain/org/remote/sync.go` (L70-86)
- **建议修复:** 添加 store 方法只查询最近一条 scheduled 类型日志。

#### 10. 表单注册全平台字段造成冗余验证项

- **文件:** `apps/frontend/src/features/org/components/data-source/step-credentials.tsx`
- **建议修复:** 在平台切换时调用 `form.reset()` 清理不相关字段。

---

## 二、组织架构模块

### 高危

#### 1. 成员更新：部分字段发送 + 全量替换 = 数据丢失

- **前端:** `apps/frontend/src/features/org/hooks/use-structure-page.ts` (L98)
- **后端:** `apps/backend/internal/domain/org/structure/member.go` (L109-123)
- **影响:** 编辑成员姓名/手机后保存，该成员的 roles、status、source、personalQuota、companyId 全部被清零。roles 被清空意味着成员失去所有权限，无法使用系统。这是当前最紧急的生产级 bug。
- **建议修复:**
  - **方案 A（推荐）:** 后端改用 merge 更新 — 先读取现有 member，仅覆盖请求中非零值字段。
  - **方案 B:** 前端在提交前先 GET 完整 member，合并编辑字段后 PUT 完整对象。

### 中危

#### 2. 前端表单收集后端不存在的字段

- **文件:** `apps/frontend/src/features/org/components/structure/member-form-dialog.tsx`
- **影响:** username、employeeId、jobTitle、hireDate 在后端 `types.Member` 中无对应字段，数据静默丢弃。用户以为信息已保存。
- **建议修复:** 短期移除表单中未支持字段并加 TODO；长期在后端 Member 结构体中添加扩展字段。

#### 3. CreateMember 缺少事务保护

- **文件:** `apps/backend/internal/domain/org/structure/member.go` (L60)
- **影响:** SetMembers 成功但 persistRecalculatedMemberCounts 失败时，部门成员数不一致。
- **建议修复:** 参照 CreateDepartment 使用 `WithTx` 包装。

#### 4. 部门成员计数包含已停用成员

- **文件:** `apps/backend/internal/pkg/org/org_nodes.go` (L102)
- **影响:** 含已停用成员的部门无法删除（前端依据 memberCount > 0 阻止）。前后端判断逻辑不一致（后端 HasDirectActiveMembers 排除 inactive，前端用 memberCount 不排除）。
- **建议修复:** `RecalcOrgNodeMemberCounts` 过滤 `status != "inactive"` 的成员。

#### 5. 部门/成员 ID 使用毫秒时间戳，并发冲突风险

- **文件:** `apps/backend/internal/domain/org/structure/department.go` (L29), `member.go` (L76)
- **影响:** 同一毫秒内的并发请求会生成相同 ID。
- **建议修复:** 改用 UUID 或 ULID；至少添加随机后缀。

#### 6. BatchImport 未重算部门成员计数

- **文件:** `apps/backend/internal/domain/org/structure/member.go` (L260)
- **影响:** 批量导入后部门树显示的成员数过时。
- **建议修复:** 在 BatchImport 末尾调用 `persistRecalculatedMemberCounts`。

#### 7. 部门树操作无乐观锁

- **影响:** 两个管理员同时修改树结构时最后写入者覆盖前者，无冲突检测。
- **建议修复:** 添加 version 字段，更新时校验版本号。

### 低危

#### 8. 删除确认文案与实际行为不符

- **文件:** `apps/frontend/src/features/org/hooks/use-structure-page.ts` (L139)
- **建议修复:** 文案改为"成员将被停用，可由管理员重新激活"。

#### 9. 部门搜索过滤逻辑不一致

- **文件:** `apps/frontend/src/features/org/components/structure/department-panel.tsx` (L72-83)
- **影响:** 匹配父节点时子节点未被过滤，视觉上信息噪声较大。

#### 10. MembersInvite handler 丢弃请求体

- **文件:** `apps/backend/internal/http/handler/org/member.go` (L82)
- **影响:** 当前返回 501，不影响运行；但实现时需重构 handler。

#### 11. UpdateMemberStatus 未触发计数重算

- **文件:** `apps/backend/internal/domain/org/structure/member.go` (L141)
- **影响:** 当前被 RecalcOrgNodeMemberCounts 计入 inactive 成员掩盖，修复计数逻辑后会暴露。

---

## 三、角色管理模块

### 高危

#### 1. AddRoleMember 无角色提升保护

- **文件:** `apps/backend/internal/domain/org/structure/role.go` (L143)
- **影响:** 任何拥有 `org:roles` 权限的用户可通过 `POST /roles/{roleId}/members` 将自己添加到"超级管理员"角色，绕过 `validateRolesNotEscalated` 检查。这是权限提升漏洞。
- **建议修复:** 在 AddRoleMember 开头检查目标角色是否为 preset 类型且为受保护角色（超级管理员、组织管理员），如是则拒绝或要求更高权限。

#### 2. UpdateRole 未阻止修改预设角色

- **文件:** `apps/backend/internal/domain/org/structure/role.go` (L42)
- **影响:** 攻击者可修改"超级管理员"角色的权限列表，降低其权限或重命名角色制造混乱。前端虽隐藏了编辑按钮，但 API 层无保护。
- **建议修复:** 添加 `if role.Type == "preset" { return ErrCannotModifyPresetRole }`，与 DeleteRole 保持一致。

### 中危

#### 3. CreateRole 不检查角色名重复

- **文件:** `apps/backend/internal/domain/org/structure/role.go` (L19)
- **影响:** 由于成员的 roles 字段存储角色名称字符串，重复名称会导致权限解析歧义。
- **建议修复:** 在创建前查询是否存在同名角色。

#### 4. AddRoleMember 对不存在的成员静默成功

- **文件:** `apps/backend/internal/domain/org/structure/role.go` (L143)
- **影响:** API 返回 200 但实际未修改任何数据，前端显示"添加成功"但成员未出现。
- **建议修复:** 循环未命中时返回 `ErrMemberNotFound`。

#### 5. 无法阻止移除最后一个超级管理员

- **文件:** `apps/backend/internal/domain/org/structure/role.go`
- **影响:** 移除超级管理员角色的最后一个成员后，组织失去管理权限，无法恢复。
- **建议修复:** RemoveRoleMember 检查：若目标角色为超级管理员且当前成员数为 1，返回错误。

#### 6. DeleteRole 的 SetMembers 和 SetRoles 缺少事务

- **文件:** `apps/backend/internal/domain/org/structure/role.go` (L67)
- **影响:** SetMembers 成功后 SetRoles 失败，成员丢失角色引用但角色仍存在。
- **建议修复:** 使用 `WithTx` 包装两步操作。

#### 7. 前端异步操作无 try/catch

- **文件:** `apps/frontend/src/features/org/hooks/use-roles-page.ts` (L80, 90, 109, 116)
- **影响:** API 失败时 Promise rejection 未处理，用户无错误反馈，弹窗状态可能不正确。
- **建议修复:** 所有 async handler 添加 try/catch + toast.error。

### 低危

#### 8. CreateRole/UpdateRole 不验证空名称

- **建议修复:** 添加 `strings.TrimSpace(name)` 非空校验。

#### 9. 前端 handleEditRole 未阻止编辑预设角色

- **建议修复:** 添加 `if (role.type === 'preset') return` 防御性检查。

#### 10. 角色表单允许提交零权限

- **建议修复:** 添加 form 验证规则，要求至少选中一项权限。

#### 11. 添加成员弹窗未自动关闭

- **文件:** `apps/frontend/src/features/org/components/role-member-table.tsx` (L226)
- **建议修复:** onAdd 回调完成后关闭弹窗或刷新搜索结果。

---

## 四、跨模块共性问题

| 模式缺陷 | 影响模块 | 说明 |
|----------|---------|------|
| 事务边界不完整 | 数据源、组织架构、角色管理 | 多步 store 操作未用 WithTx 包装，中途失败导致不一致 |
| 前端异步无错误处理 | 数据源（部分）、角色管理 | async handler 缺少 try/catch，用户无失败反馈 |
| ID 生成用毫秒时间戳 | 组织架构 | dept/member ID 并发碰撞风险 |
| 预设角色保护不一致 | 角色管理 | Delete 有保护，Update/AddMember 无保护 |
| 部门成员计数逻辑 | 组织架构 | 多处修改成员后未调用 recalc，且 recalc 不过滤 inactive |

---

## 五、优先修复行动计划（Top 10）

| 优先级 | 问题 | 模块 | 原因 |
|--------|------|------|------|
| **P0** | 成员更新全量替换导致角色/状态丢失 | 组织架构 | 生产环境每次编辑成员都会清除权限，影响所有用户 |
| **P0** | AddRoleMember 无角色提升保护 | 角色管理 | 安全漏洞，允许任意权限提升 |
| **P0** | UpdateRole 可修改预设角色 | 角色管理 | 安全漏洞，可篡改超级管理员权限定义 |
| **P1** | SyncConfig 负阈值绕过删除保护 | 数据源 | 可导致同步全量删除成员不被拦截 |
| **P1** | 阻止移除最后一个超级管理员 | 角色管理 | 可导致组织永久失去管理权限 |
| **P1** | CreateRole 角色名重复检查 | 角色管理 | 权限解析歧义，影响授权正确性 |
| **P2** | 字段映射测试返回假数据 | 数据源 | 用户配置错误无法被发现，增加运维成本 |
| **P2** | 部门成员计数包含 inactive | 组织架构 | 前后端判断不一致，用户无法删除空部门 |
| **P2** | 平台切换未清除旧映射 | 数据源 | 切换平台后同步使用错误字段映射 |
| **P2** | CreateMember/DeleteRole 添加事务保护 | 组织架构/角色 | 数据一致性基础保障 |

---

*P0 建议在本迭代内修复，P1 在下一迭代完成，P2 纳入技术债务持续改进。*
