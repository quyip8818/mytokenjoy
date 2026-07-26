# Member/User 边界重构

## 问题

当前 `types.Member` struct 混合了两种数据：
- **member 自有字段**（alias、departmentId、employeeId 等）——公司内的职位身份
- **user 字段**（phone、email、name）——全局用户身份

前端编辑成员表单把两者混在一起，标签和数据映射混乱：
- "姓名" input 实际存的是 `member.alias`
- "昵称" input 实际存的是 `member.username`
- `user.name`（真实姓名）从未被前端使用

## 目标数据模型

### users 表（全局用户身份）
| 字段 | 说明 |
|------|------|
| name | 用户真实姓名（跨公司通用） |
| phone | 手机号（登录凭证） |
| email | 邮箱（登录凭证） |
| password_hash | 密码 |

### members 表（公司内成员身份）
| 字段 | 说明 |
|------|------|
| alias | 昵称/显示名（公司内） |
| department_id | 部门 |
| employee_id | 工号 |
| job_title | 职位 |
| hire_date | 入职时间 |
| status | 状态 |
| source | 来源 |

### 删除的字段
- `member.username` → 删除。alias 就是昵称，不需要两个名字字段。
- `types.Member.Phone/Email/Name` → 语义从"input-only 写入字段"改为"只读 JOIN 展示字段"。创建和更新不再通过 Member struct 传递，而是通过独立的 User API 或嵌套的 `user` 子结构。列表通过 JOIN users 表只读返回。

## API 总览

重构后管理员所有操作保持在 `/org/members` 下，不引入顶级 `/org/users`。

| API | 用途 | 变更 |
|-----|------|------|
| `GET /org/members` | 成员列表 | JOIN users 返回 name/phone/email（只读） |
| `POST /org/members` | 创建成员 | body 改为 `{user:{...}, member:{...}}` resolveOrCreateUserMember 语义 |
| `PUT /org/members/:id` | 更新 member 字段 | 只含 alias/departmentId/employeeId/jobTitle/hireDate |
| `PUT /org/members/:id/user` | 管理员改别人的 user 字段 | **新增**，含 name/phone/email |
| `DELETE /org/members` | 删除成员 | 不变 |
| `PUT /org/members/status` | 批量状态 | 不变 |
| `POST /org/members/transfer` | 批量转部门 | 不变 |
| `POST /org/members/:id/invite-link` | 获取邀请链接 | 不变 |
| `POST /org/members/batch-invite` | 批量重发邀请 | 不变 |
| `POST /org/members/batch-import` | CSV 导入 | 内部 resolveOrCreate，接口不变 |
| `PUT /me/profile` | 用户改自己 | 保持现状（name/avatar/alias） |

## API 详细

### Member 列表响应（GET /org/members）
JOIN users 表返回只读展示字段：

```json
{
  "id": "...",
  "alias": "小曲",       // 昵称（member 表）
  "name": "曲一鹏",     // 姓名（users 表，JOIN 只读）
  "phone": "138...",    // 手机（users 表，JOIN 只读）
  "email": "q@x.com",  // 邮箱（users 表，JOIN 只读）
  "departmentId": "...",
  "employeeId": "E001",
  "jobTitle": "工程师",
  "hireDate": "2024-03-01",
  ...
}
```

### 创建成员（POST /org/members）— resolveOrCreateUserMember
一个 API 原子完成 user + member 创建：

```json
{
  "user": { "name": "曲一鹏", "phone": "138...", "email": "q@x.com" },
  "member": { "alias": "小曲", "departmentId": "...", "employeeId": "E001", "jobTitle": "工程师", "hireDate": "2024-03-01" }
}
```

后端流程：
1. resolveOrCreate user（按 phone/email 查找，不存在则创建）
2. 创建 member（关联 userId，status=pending）
3. 创建 invite + 发送邀请通知

`member.alias` 可选，缺省复制 `user.name`。

### 更新成员（PUT /org/members/:id）
只更新 member 字段：

```json
{ "alias": "小曲", "departmentId": "...", "employeeId": "E001", "jobTitle": "工程师", "hireDate": "2024-03-01" }
```

### 更新成员关联用户（PUT /org/members/:id/user）— 新增
管理员修改成员对应的 user 信息：

```json
{ "name": "曲一鹏", "phone": "138...", "email": "q@x.com" }
```

后端通过 member.user_id 找到对应 user 更新。

**唯一性冲突策略**：如果新 phone/email 已被其他 user 占用，直接报错"该手机号/邮箱已被其他用户使用"，不做 merge。避免管理员无意间把 member 关联到别的 user。

### 用户自改（PUT /me/profile）
保持不变，已有 name/avatar/alias 支持。

## 前端表单映射

### 编辑成员表单
| UI 标签 | 字段名 | API | 存储位置 |
|---------|--------|-----|----------|
| 姓名 * | name | PUT /org/members/:id/user | users.name |
| 昵称 | alias | PUT /org/members/:id | members.alias |
| 手机号 | phone | PUT /org/members/:id/user | users.phone |
| 邮箱 | email | PUT /org/members/:id/user | users.email |
| 工号 * | employeeId | PUT /org/members/:id | members.employee_id |
| 职位 | jobTitle | PUT /org/members/:id | members.job_title |
| 入职时间 | hireDate | PUT /org/members/:id | members.hire_date |
| 主部门 * | departmentId | PUT /org/members/:id | members.department_id |

前端提交时串行调用两个 API：先 `PUT /org/members/:id/user`（user 字段），成功后再 `PUT /org/members/:id`（member 字段）。任一失败直接报错，不继续后续调用。

### 添加成员表单
所有字段一次提交到 `POST /org/members`（`{user, member}` 嵌套结构）。

### 接受邀请表单
| UI 标签 | 字段名 | 存储位置 |
|---------|--------|----------|
| 昵称 | alias | members.alias |
| 密码 | password | users.password_hash |
| 确认密码 | - | 前端校验 |

## 后端改动清单

1. **members SQL query**：JOIN users 表返回 name/phone/email（只读展示）
2. **types.Member**：删除 `Username`；`Phone`/`Email`/`Name` 改为只读 JOIN 展示字段
3. **CreateMember（重写）**：改为 resolveOrCreateUserMember 语义。接收 `{user, member}` 嵌套体，内部原子完成 resolveOrCreate user + 创建 member + 创建 invite + 发通知
4. **UpdateMember**：只更新 member 字段（alias、departmentId、employeeId、jobTitle、hireDate），移除 phone/email 逻辑
5. **新增 `PUT /org/members/:id/user`**：管理员更新成员的 user 信息（name/phone/email）。phone/email 唯一性冲突时报错，不做 user merge
6. **BatchImport**：内部直接调 resolveOrCreateUser 拿 userId，不通过 Member struct
7. **remote/import.go**（飞书/钉钉同步）：同上
8. **HTTP handler（member.go）**：CreateMember body 改为 `{user, member}` 嵌套结构；新增 MemberUserUpdate handler
9. **前端 Member 类型**：删除 `username`；列表响应新增只读 `name`/`phone`/`email`（来自 JOIN）
10. **member-form-dialog**：创建时提交 `{user, member}` 到 POST；更新时串行调用 user PUT → member PUT，任一失败直接报错不继续
11. **use-structure-page**：更新逻辑拆分两个串行 API 调用
12. **invite-accept**：保持当前行为（只改 member.alias + 设密码）
13. **schema.sql**：members 表增加 `hire_date` 列
14. **Seed 数据**：删除重建（破坏性迁移，`pnpm reset`）

## 不需要改的

- `resolveOrCreateUser` 函数保留（CreateMember 和 BatchImport 内部使用）
- `store.MemberAuthz`（已 JOIN users 取 UserName）
- `SessionContext.User.Name`（已正确从 users 表读）
- `PUT /me/profile`（用户自改，保持现状）
- invitetoken / invite 流程
- feishu/dingtalk adapter（内部处理）

## 破坏性变更

- `POST /org/members` 请求体从 flat 改为 `{user, member}` 嵌套
- `PUT /org/members/:id` 请求体移除 phone/email/username
- 列表响应新增 name/phone/email（JOIN），移除 username
- 新增 `PUT /org/members/:id/user`
- 前端 Member 类型定义变更
- 前端表单提交逻辑变更（创建一次调用，更新串行两次调用）
- schema.sql 增加 hire_date 列
- seed 数据需要重建（`pnpm reset`）
