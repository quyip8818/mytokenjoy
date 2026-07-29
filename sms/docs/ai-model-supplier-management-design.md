# AI 大模型供应商管理系统 · 设计文档

日期：2026-07-23（按实际代码重写）

## 1. 背景与目标

企业内部供应商管理系统，管理对象为 AI 大模型供应商（OpenAI、Anthropic、阿里云、DeepSeek 等厂商）及其提供的模型服务。

### 目标

- 统一管理大模型厂商档案、联系人、合作状态
- 管理合同与采购订单，合同原文电子归档，到期预警
- 五维度绩效评估对供应商打分评级，辅助采购决策
- 登录认证 + 多角色权限控制

### 使用角色

| 角色 | 权限范围 |
|------|---------|
| admin 管理员 | 全部功能 + 用户管理 + 评估权重配置 |
| buyer 采购员 | 供应商/模型/合同/订单维护、评估打分 |
| viewer 只读 | 仅查看各模块数据、下载合同附件 |

## 2. 技术选型

| 层 | 技术 |
|---|---|
| 前端 | React 19 + Vite 8 + TanStack Query + Zustand + Radix UI + Tailwind CSS 4 + react-router v7 |
| 后端 | Go 1.23 + chi (HTTP 路由) + pgx v5 (PostgreSQL 驱动) + golang-jwt/jwt v5 |
| 数据库 | PostgreSQL（原生 SQL，无 ORM） |
| 认证 | JWT Access Token (短期) + Refresh Token (Session 表，httpOnly cookie) |
| Monorepo | pnpm workspace，sms/backend + sms/frontend（位于 mytokenjoy monorepo） |

### 选型理由

- Go：编译型语言，类型安全，部署一个二进制，并发性能好
- pgx 原生 SQL：完全掌控查询，避免 ORM 魔法；domain 层通过 Store interface 解耦
- React + TanStack Query：服务器状态管理成熟，与 Zustand 配合处理客户端状态
- 双 Token 认证：比纯 JWT 更安全，access token 短期失效，refresh token 支持服务端吊销

## 3. 整体架构

```
mytokenjoy/sms/
├── backend/                          # Go 后端
│   ├── cmd/server/main.go            # 入口：加载配置 → 构建 App → 启动 HTTP
│   ├── cmd/seed/main.go              # 种子数据脚本
│   ├── schema.sql                    # DDL（幂等，IF NOT EXISTS）
│   └── internal/
│       ├── config/                   # 环境变量配置（caarlos0/env）
│       ├── app/app.go                # 依赖注入组装：pool → store → services → router
│       ├── domain/                   # 业务逻辑层（每个 domain 一个包）
│       │   ├── types/                # 共享内核：models.go, enums.go, errors.go
│       │   ├── auth/                 # 登录/刷新/登出
│       │   ├── supplier/             # 供应商 + 联系人
│       │   ├── contract/             # 合同 + 附件
│       │   ├── order/                # 采购订单
│       │   ├── evaluation/           # 绩效评估 + 权重
│       │   ├── dashboard/            # 仪表盘聚合
│       │   └── user/                 # 用户管理
│       ├── http/
│       │   ├── handler/              # HTTP 适配层（每个 domain 一个子包）
│       │   ├── middleware/           # auth.go, cors.go, role.go
│       │   ├── helpers/              # 参数解析、错误映射
│       │   ├── response/             # 统一响应格式
│       │   └── deps/deps.go          # 依赖容器 struct
│       └── store/
│           ├── pool.go               # pgxpool 初始化
│           ├── errors.go             # SQL 错误到 domain 错误转换
│           └── postgres/             # 全部 SQL 查询实现
├── frontend/                         # React 前端
│   └── src/
│       ├── api/                      # HTTP 客户端（fetch 封装 + 各 domain API）
│       ├── config/enums.ts           # 前端状态枚举/标签映射
│       ├── features/                 # 领域特性包（hooks + components + lib）
│       │   ├── session/              # 登录态管理
│       │   ├── query/                # TanStack Query 封装 + queryKeys
│       │   ├── suppliers/            # 供应商
│       │   ├── contracts/            # 合同
│       │   ├── orders/               # 订单
│       │   ├── evaluations/          # 评估
│       │   ├── dashboard/            # 仪表盘
│       │   └── users/                # 用户管理
│       ├── routes/                   # 页面入口（从 features 导入）
│       ├── components/ui/            # 原子组件（无业务语义）
│       └── components/layout/        # 布局壳
└── docs/                             # SMS 产品文档
```

### 分层约束

- `cmd/` 仅入口 + 启动编排，禁止业务逻辑
- `domain/{name}/` 定义 `Store` interface，不依赖具体数据库实现（依赖倒置）
- `domain/types/` 为共享内核，可被所有 domain 引用
- 跨 domain 调用通过 exported interface 或 value types，不引用对方内部实现
- 前端 features 之间只通过 `index.ts` barrel 引用，禁止 deep import

## 4. 数据模型（10 张表）

### 4.1 roles 角色

| 字段 | 类型 | 说明 |
|---|---|---|
| id | SERIAL PK | |
| code | VARCHAR(32) UNIQUE | admin / buyer / viewer |
| name | VARCHAR(64) | 角色名称 |

预置三条。仅做展示用途，实际权限判断读 `users.role` 字段。

### 4.2 users 用户

| 字段 | 类型 | 说明 |
|---|---|---|
| id | UUID PK | gen_random_uuid() |
| username | VARCHAR(64) UNIQUE | 登录名 |
| password_hash | TEXT | bcrypt 加密 |
| real_name | VARCHAR(64) | 姓名 |
| email | VARCHAR(128) | 邮箱 |
| role | VARCHAR(32) | 角色 code（admin/buyer/viewer） |
| status | SMALLINT DEFAULT 1 | 1 启用 / 0 停用 |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

注：role 直接存字符串而非 FK，避免不必要的 join。status=0 的用户登录被拒绝。

### 4.3 sessions 会话

| 字段 | 类型 | 说明 |
|---|---|---|
| token | VARCHAR(64) PK | refresh token（UUID） |
| user_id | UUID FK → users | |
| expires_at | TIMESTAMPTZ | 过期时间 |
| created_at | TIMESTAMPTZ | |

Refresh token 存服务端，支持主动吊销和轮换。

### 4.4 suppliers 供应商

| 字段 | 类型 | 说明 |
|---|---|---|
| id | SERIAL PK | |
| name | VARCHAR(128) | 厂商名称 |
| code | VARCHAR(64) UNIQUE | 厂商编码 |
| category | VARCHAR(64) | 国内厂商 / 国外厂商 |
| website | VARCHAR(256) | 官网 |
| status | VARCHAR(32) | potential / active / frozen / blacklisted |
| description | TEXT | 备注 |
| created_by | UUID FK → users | 创建人 |
| created_at / updated_at | TIMESTAMPTZ | |

### 4.5 supplier_contacts 联系人

| 字段 | 类型 | 说明 |
|---|---|---|
| id | SERIAL PK | |
| supplier_id | INT FK → suppliers | CASCADE 删除 |
| name | VARCHAR(64) | 姓名 |
| position | VARCHAR(64) | 职务 |
| phone | VARCHAR(32) | 电话 |
| email | VARCHAR(128) | 邮箱 |
| is_primary | BOOLEAN | 是否主要联系人 |
| created_at | TIMESTAMPTZ | |

### 4.6 contracts 合同

| 字段 | 类型 | 说明 |
|---|---|---|
| id | SERIAL PK | |
| supplier_id | INT FK → suppliers | |
| contract_no | VARCHAR(64) UNIQUE | 合同编号 |
| title | VARCHAR(256) | 合同标题 |
| amount | NUMERIC(14,2) | 合同金额 |
| sign_date | DATE | 签订日期 |
| start_date | DATE | 生效日期 |
| end_date | DATE | 到期日期 |
| status | VARCHAR(32) | draft / active / expired / terminated |
| remarks | TEXT | 备注 |
| created_by | UUID FK → users | 创建人 |
| created_at / updated_at | TIMESTAMPTZ | |

### 4.7 contract_attachments 合同附件

| 字段 | 类型 | 说明 |
|---|---|---|
| id | SERIAL PK | |
| contract_id | INT FK → contracts | CASCADE 删除 |
| file_name | VARCHAR(256) | 原始文件名 |
| file_path | VARCHAR(512) | 服务器存储路径 |
| file_size | BIGINT | 文件大小（字节） |
| uploaded_by | UUID FK → users | 上传人 |
| created_at | TIMESTAMPTZ | |

存储：本地磁盘 `$UPLOAD_DIR/`（默认 `./uploads`），文件名 UUID + 原始扩展名。删除合同时 CASCADE 删附件记录。

### 4.8 purchase_orders 采购订单

| 字段 | 类型 | 说明 |
|---|---|---|
| id | SERIAL PK | |
| order_no | VARCHAR(64) UNIQUE | 订单编号 |
| supplier_id | INT FK → suppliers | |
| contract_id | INT FK → contracts | 可空，关联合同 |
| total_amount | NUMERIC(14,2) | 采购金额 |
| order_date | DATE | 下单日期 |
| status | VARCHAR(32) | pending / approved / delivered / completed / cancelled |
| description | TEXT | 说明 |
| created_by | UUID FK → users | 创建人 |
| created_at / updated_at | TIMESTAMPTZ | |

**状态流转**（后端强制校验）：
```
pending → approved → delivered → completed
    ↘        ↘         ↘
    cancelled  cancelled  cancelled
```

### 4.9 evaluations 绩效评估

| 字段 | 类型 | 说明 |
|---|---|---|
| id | SERIAL PK | |
| supplier_id | INT FK → suppliers | CASCADE 删除 |
| evaluator_id | UUID FK → users | 评估人 |
| period | VARCHAR(32) | 评估周期（如 2026-Q3） |
| quality | INT 0–100 | 模型质量 |
| performance | INT 0–100 | 响应性能 |
| price | INT 0–100 | 价格成本 |
| service | INT 0–100 | 服务支持 |
| compliance | INT 0–100 | 合规安全 |
| total_score | NUMERIC(5,2) | 综合分（自动计算） |
| grade | VARCHAR(2) | A / B / C / D（自动生成） |
| comment | TEXT | 评语 |
| created_at | TIMESTAMPTZ | |

唯一约束：`UNIQUE (supplier_id, period)`。

### 4.10 evaluation_weights 评估权重

| 字段 | 类型 | 说明 |
|---|---|---|
| id | SERIAL PK | |
| dimension | VARCHAR(32) UNIQUE | quality / performance / price / service / compliance |
| weight | INT 0–100 | 权重（%），五项合计 100 |

预置：质量 30、性能 20、价格 20、服务 20、合规 10。仅 admin 可修改。

## 5. 认证与权限

### 双 Token 认证流程

1. `POST /api/auth/login` → 验证密码 + 检查 status → 签发 access token (JWT, 15min) + 创建 session (refresh token, 7天，httpOnly cookie)
2. 前端请求携带 `Authorization: Bearer {accessToken}`
3. Access token 过期时前端调用 `POST /api/auth/refresh`（cookie 携带 refresh token）→ 轮换 refresh token + 签发新 access token
4. `POST /api/auth/logout` → 删除 session 记录

### JWT Payload

```json
{ "id": "uuid", "username": "admin", "role": "admin", "exp": 1753000000 }
```

### 角色权限矩阵

| 功能 | admin | buyer | viewer |
|---|---|---|---|
| 供应商/联系人/合同/订单 增删改 | ✅ | ✅ | ❌ |
| 合同附件上传/删除 | ✅ | ✅ | ❌ |
| 合同附件下载 | ✅ | ✅ | ✅ |
| 绩效评估打分 | ✅ | ✅ | ❌ |
| 评估权重配置 | ✅ | ❌ | ❌ |
| 用户管理 | ✅ | ❌ | ❌ |
| 仪表盘/列表/详情 | ✅ | ✅ | ✅ |

实现：`middleware.Auth` 解析 JWT 注入 `AuthUser`，`middleware.RequireRole("admin", "buyer")` 在路由层控制写权限。前端按角色隐藏按钮（体验优化，安全以后端为准）。

## 6. 绩效评估逻辑

- 五维度各 0–100 分，百分制
- 综合分：`total_score = Σ(维度分 × 该维度权重) / 100`
- 评级规则：`A ≥ 90`，`80 ≤ B < 90`，`60 ≤ C < 80`，`D < 60`
- 提交时后端自动计算 total_score 和 grade
- 前端打分表单实时预览（计算逻辑一致，后端结果为准）
- 同一供应商同一周期唯一约束，不可重复评估

## 7. API 设计

统一前缀 `/api`，统一 JSON 响应。分页参数 `page` / `pageSize`，返回 `{ items, total, page, pageSize }`。

```
POST   /api/auth/login              登录
POST   /api/auth/refresh            刷新 token
POST   /api/auth/logout             登出
GET    /api/auth/profile            当前用户信息

GET    /api/dashboard/summary       仪表盘（统计卡片 + 到期合同 + 评级分布 + 模型分布）

GET    /api/suppliers               列表（keyword/status/category 筛选，分页）
GET    /api/suppliers/options       下拉选项（id+name）
GET    /api/suppliers/{id}          详情（含联系人 + 合同 + 订单 + 评估）
POST   /api/suppliers               新建
PUT    /api/suppliers/{id}          更新
DELETE /api/suppliers/{id}          删除（有关联合同/订单时拒绝）
POST   /api/suppliers/{id}/contacts         添加联系人
PUT    /api/suppliers/{id}/contacts/{cid}   更新联系人
DELETE /api/suppliers/{id}/contacts/{cid}   删除联系人

GET    /api/contracts               列表（keyword/supplierId/status 筛选）
GET    /api/contracts/{id}          详情（含附件列表）
POST   /api/contracts               新建
PUT    /api/contracts/{id}          更新
DELETE /api/contracts/{id}          删除（有关联订单时拒绝）
POST   /api/contracts/{id}/attachments              上传附件（multipart）
GET    /api/contracts/{id}/attachments/{aid}/download  下载附件
DELETE /api/contracts/{id}/attachments/{aid}         删除附件

GET    /api/purchase-orders         列表（keyword/supplierId/status 筛选）
POST   /api/purchase-orders         新建
PUT    /api/purchase-orders/{id}    更新（含状态流转校验）
DELETE /api/purchase-orders/{id}    删除

GET    /api/evaluations             列表（supplierId/period 筛选）
POST   /api/evaluations             提交评估
PUT    /api/evaluations/{id}        修改评估
DELETE /api/evaluations/{id}        删除评估
GET    /api/evaluations/weights     权重查看
PUT    /api/evaluations/weights     权重配置（admin）

GET    /api/users                   用户列表（admin）
POST   /api/users                   新建用户（admin）
PUT    /api/users/{id}              更新用户（admin）
DELETE /api/users/{id}              删除用户（admin）
GET    /api/users/roles             角色列表
```

### 错误处理

后端 `helpers.HandleDomainError` 将 domain 错误映射到 HTTP 状态码：
- `ErrValidation` → 400
- `ErrUnauthorized` → 401
- `ErrNotFound` → 404
- `ErrConflict`（唯一约束冲突）→ 409
- `ErrHasRefs`（有关联数据）→ 409
- 未知错误 → 500

前端通过 fetch client 统一拦截：401 触发 token 刷新或跳转登录，其他错误 toast 提示。

## 8. 合同附件方案

- 上传限制 32MB（ParseMultipartForm）
- 存储目录由 `UPLOAD_DIR` 环境变量控制，默认 `./uploads`
- 文件名：UUID + 原始扩展名，避免重名和路径注入
- 下载：按附件 ID 查库获取 file_path，`http.ServeFile` 流式返回，设置 Content-Disposition
- 删除合同时 CASCADE 删附件记录；删除附件时同步清理磁盘文件
- 记录上传人 `uploaded_by`

## 9. 前端页面结构

### 登录页
用户名/密码登录，失败 toast 提示。

### 主布局（侧边栏 + 内容区）

| 页面 | 功能 |
|---|---|
| 仪表盘 | 统计卡片（供应商总数/合作中/活跃合同）、30天内到期合同预警、评级分布 |
| 供应商管理 | 列表（状态/分类筛选、关键字搜索、分页）；详情页 Tab：联系人、合同、订单、评估历史 |
| 合同管理 | 列表（到期天数高亮：≤30天橙色、已过期红色）；详情抽屉含附件区 |
| 采购订单 | 列表 + 状态标记 + 下单日期 |
| 绩效评估 | 列表 + 打分弹窗（五维度滑块，实时预览综合分与评级） |
| 系统管理（admin） | 用户管理（启用/停用）、评估权重配置 |

### 前端文件约定

- 页面入口 `routes/{domain}/`：仅组合，从 features 导入
- 领域逻辑 `features/{domain}/`：hooks + components + lib + index.ts barrel
- HTTP 层 `api/{domain}.ts`：类型定义 + fetch 调用
- 原子组件 `components/ui/`：StatusBadge, Field, Pagination, Badge 等
- 禁止直接 import API 函数——通过 `useApis()` 注入

## 10. 配置

环境变量（后端 `.env.development`）：

| 变量 | 默认值 | 说明 |
|---|---|---|
| PORT | 8020 | 后端端口 |
| DATABASE_URL | （必填） | PostgreSQL 连接串 |
| CORS_ORIGINS | http://localhost:5174 | 允许的前端域 |
| JWT_SECRET | sms-dev-secret | JWT 签名密钥 |
| ACCESS_TOKEN_TTL_MIN | 15 | Access Token 有效期（分钟） |
| REFRESH_TOKEN_TTL_H | 168 | Refresh Token 有效期（小时，7天） |
| UPLOAD_DIR | ./uploads | 附件存储目录 |
| SECURE_COOKIE | false | 生产环境设为 true |

前端 Vite 开发代理 `/api` → 后端 8020 端口，开发服务器监听 5174。

## 11. 初始化数据

通过 `cmd/seed/main.go` 执行：

1. 应用 `schema.sql`（幂等）
2. 插入三个角色：admin、buyer、viewer
3. 创建管理员账号：admin / admin123
4. 插入默认评估权重（五条）

命令：`make reset`（开发环境重建数据库）

## 12. 开发命令

```bash
# 在 monorepo 根目录（mytokenjoy/）执行
pnpm install          # 安装所有依赖
pnpm infra            # 启动基础设施（postgres:5510 + redis:6310）
pnpm start sms        # 启动 sms（backend:8020 + frontend:5174）
pnpm reset sms        # 重置 sms 数据库 + seed
pnpm test:sms         # 运行 sms 全部测试
pnpm lint:sms         # sms lint
pnpm build:sms        # sms 前端构建
```
