# SMS — AI 模型供应商管理系统

## 概述

SMS 是一套内部供应商管理系统，用于管理 AI 大模型供应商（OpenAI、Anthropic、DeepSeek 等）及其采购生命周期。制品不出门，仅内部使用。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.23, chi/v5, pgx/v5, JWT |
| 前端 | React 19, Vite 8, TanStack Query 5, Zustand, Radix UI, Tailwind 4 |
| 数据库 | PostgreSQL（共享实例，3 库：sms, sms_newapi, sms_logs） |
| 部署 | monorepo 内，端口 8020(后端) / 5174(前端) / 3020(NewAPI-SMS) |

## 目录结构

```
sms/
├── backend/          # Go 后端
│   ├── cmd/
│   │   ├── server/   # HTTP 服务入口
│   │   └── seed/     # 数据种子脚本
│   ├── internal/
│   │   ├── app/      # 依赖注入组装
│   │   ├── config/   # 环境配置
│   │   ├── domain/   # 业务逻辑（10 个子域）
│   │   ├── http/     # 传输层（handler, middleware, router）
│   │   ├── integration/  # 外部集成（NewAPI）
│   │   └── store/    # 数据访问层（raw SQL + pgx）
│   ├── tests/        # 单元测试（镜像 internal/）
│   ├── schema.sql    # 数据库 DDL
│   └── Makefile
├── frontend/         # React 前端
│   ├── src/
│   │   ├── api/      # API 客户端
│   │   ├── components/   # 布局 + 原子 UI
│   │   ├── features/     # 业务特性模块
│   │   └── routes/       # 页面路由
│   ├── tests/        # vitest 单元测试
│   └── e2e/          # Playwright E2E
├── docs/             # 设计文档
│   ├── ai-model-supplier-management-design.md
│   └── plan/newapi-integration.md
└── newapi/           # NewAPI-SMS 本地引导脚本
    └── scripts/bootstrap-local.sh
```

## 业务域（internal/domain/）

| 域 | 职责 |
|----|------|
| `auth` | 登录、刷新、登出（双 token：JWT access 15min + session refresh 7天） |
| `supplier` | 供应商 + 联系人管理 |
| `contract` | 合同管理（草稿→生效→到期→终止） |
| `order` | 采购订单（状态机：待审批→已批准→已交付→已完成/取消） |
| `model` | AI 模型目录（关联供应商，含定价） |
| `evaluation` | 供应商绩效评估（5 维度加权评分） |
| `dashboard` | 数据聚合看板 |
| `newapisync` | 模型定价同步到 NewAPI |
| `user` | 用户管理（admin/buyer/viewer 三角色） |
| `types` | 共享内核（实体、枚举、视图 DTO） |

## 核心实体与生命周期

```
供应商: potential → active → frozen → blacklisted
合同:   draft → active → expired → terminated
订单:   pending → approved → delivered → completed (任意步骤可 cancel)
模型:   available / deprecated
评估:   A(≥90) / B(≥80) / C(≥60) / D(<60)
```

## 评估评分系统

5 个维度，每个维度 0-100 分，按可配置权重加权：
- 质量 (quality)
- 性能 (performance)
- 价格 (price)
- 服务 (service)
- 合规 (compliance)

公式：`总分 = Σ(维度分 × 权重) / 100`

## NewAPI 集成

- SMS 是模型定价的单一数据源（SOT）
- 同步策略：read-modify-write merge（不覆盖未管理的模型）
- 价格转换：`modelRatio = inputPrice / 2`, `completionRatio = outputPrice / inputPrice`
- 认证：从 `sms_newapi` 数据库读取 admin PAT，401 时自动刷新
- 触发时机：模型创建/更新（异步）或手动全量同步

## 认证模型

- 双 token 机制：
  - Access Token: JWT, 15 分钟 TTL
  - Refresh Token: 服务端 session 存储, httpOnly cookie, 7 天 TTL
- 角色：admin（全权限）, buyer（业务 CRUD）, viewer（只读）

## 配置（环境变量）

| 变量 | 默认值 | 说明 |
|------|--------|------|
| PORT | 8080 | 后端端口 |
| DATABASE_URL | (必填) | PostgreSQL DSN |
| CORS_ORIGINS | http://localhost:5173 | 允许的前端源 |
| JWT_SECRET | sms-dev-secret | JWT 签名密钥 |
| ACCESS_TOKEN_TTL_MIN | 15 | Access token 过期时间 |
| REFRESH_TOKEN_TTL_H | 168 | Refresh token 过期时间（7天） |
| UPLOAD_DIR | ./uploads | 合同附件存储 |
| NEWAPI_BASE_URL | (空=禁用) | NewAPI 地址 |
| NEWAPI_DATABASE_URL | (自动) | NewAPI 数据库连接 |

## 前端特性模块

| 模块 | 对应页面 |
|------|----------|
| dashboard | 数据看板 |
| suppliers | 供应商管理 + 详情 |
| contracts | 合同管理 |
| orders | 采购订单 |
| models | 模型目录 |
| evaluations | 供应商评估 |
| users | 用户管理 |
| session | 认证会话 |
| query | TanStack Query 工具 |

## 前端路由

```
/auth/login         — 登录
/dashboard          — 数据看板
/suppliers          — 供应商列表
/suppliers/:id      — 供应商详情
/contracts          — 合同管理
/orders             — 采购订单
/models             — 模型目录
/evaluations        — 供应商评估
/newapi             — NewAPI 同步
/system/users       — 用户管理
/system/weights     — 评估权重配置
```

## 开发命令

```bash
# 启动
pnpm start:sms          # 后端 + 前端

# 后端
cd sms/backend
make dev                # air 热重载
make seed               # 初始化数据
make test               # 运行测试
make lint               # 代码检查

# 前端
pnpm -F @sms/frontend dev
pnpm -F @sms/frontend test

# 测试
pnpm test:sms           # 全量
pnpm test:sms:e2e       # E2E
```

## 数据库

- 连接：`postgres://sms:sms@127.0.0.1:5510/sms`
- 无需 build tag（与 apps 后端不同）
- Schema 定义：`sms/backend/schema.sql`
- 11 张核心表：users, sessions, suppliers, supplier_contacts, ai_models, contracts, contract_attachments, purchase_orders, evaluations, evaluation_weights, newapi_sync_logs
