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
| status | 状态 |
| source | 来源 |

### 删除的字段
- `member.username` → 删除。alias 就是昵称，不需要两个名字字段。
- `types.Member.Phone/Email` 作为 input-only 字段保留（用于 CreateMember 时解析/创建 user），但不再在 member 列表响应中返回。

## API 变更

### Member 列表响应（GET /org/members）
新增 `name` 字段（来自 users 表 JOIN）：

```json
{
  "id": "...",
  "alias": "小曲",       // 昵称（member 表）
  "name": "曲一鹏",     // 姓名（users 表）
  "phone": "138...",    // 手机（users 表）
  "email": "q@x.com",  // 邮箱（users 表）
  "departmentId": "...",
  "employeeId": "E001",
  "jobTitle": "工程师",
  ...
}
```

phone/email/name 通过 JOIN users 表返回，只读展示。

### 创建成员（POST /org/members）
请求体只包含 member 字段：
- `alias` → `member.alias`（昵称）
- `departmentId` → `member.department_id`
- `employeeId` → `member.employee_id`
- `jobTitle` → `member.job_title`
- `hireDate` → `member.hire_date`（如果需要）

**不传 name/phone/email**。用户身份通过独立的 user API 管理。

### 创建用户（POST /org/users 或复用现有 resolveOrCreateUser）
前端在创建成员前先创建/解析 user：
- `name` → `user.name`
- `phone` → `user.phone`
- `email` → `user.email`

返回 `userId`，前端再用这个 userId 创建 member。

或者更简单：**CreateMember 接收 `userId` 参数**（前端先确保 user 存在），member API 本身不碰 user 字段。

### 更新成员（PUT /org/members/:id）
只更新 member 字段：alias、departmentId、employeeId、jobTitle、hireDate。

### 更新用户（PUT /org/users/:id 或 PATCH）
独立 API 更新 user 字段：name、phone、email。

### 删除 username 字段
- 后端 `types.Member.Username` 字段删除
- 数据库 members 表无 username 列（当前没有，只在 Go struct 里）
- 前端 Member 类型删除 `username`

## 前端表单映射

### 编辑成员表单
| UI 标签 | 字段名 | 存储位置 |
|---------|--------|----------|
| 姓名 * | name | users.name |
| 昵称 | alias | members.alias |
| 手机号 | phone | users.phone |
| 邮箱 | email | users.email |
| 工号 * | employeeId | members.employee_id |
| 职位 | jobTitle | members.job_title |
| 入职时间 | hireDate | 不持久化（或加列） |
| 主部门 * | departmentId | members.department_id |

### 接受邀请表单
| UI 标签 | 字段名 | 存储位置 |
|---------|--------|----------|
| 昵称 | alias | members.alias |
| 密码 | password | users.password_hash |
| 确认密码 | - | 前端校验 |

## 后端改动清单

1. **members SQL query**：JOIN users 表返回 name/phone/email（只读展示）
2. **types.Member**：删除 `Username`、`Phone`、`Email` 字段（不再作为 input-only）
3. **CreateMember**：只接收 member 字段 + `userId`（调用方负责先创建 user）
4. **UpdateMember**：只更新 member 字段，不碰 user
5. **新增 user API**：`POST /org/users`（创建）、`PUT /org/users/:id`（更新 name/phone/email）
6. **前端 Member 类型**：删除 `username`；`name`/`phone`/`email` 为只读展示字段
7. **member-form-dialog**：拆分为两步或并行调用（user 字段走 user API，member 字段走 member API）
8. **invite-accept**：保持当前行为（只改 member.alias + 设密码）
9. **Seed 数据**：删除重建（破坏性迁移，`pnpm reset`）

## 破坏性变更

- members 列表 API 响应新增 `name` 字段（来自 user）
- 删除 `username` 字段
- 前端表单字段映射变更
- seed 数据需要重建
