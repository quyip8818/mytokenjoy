# P1 平台初始化 - 前端开发任务

## 概述

基于 TokenJoy PRD V2.0 的 P1 阶段，拆解为可独立交付的前端开发任务。

**范围：** US-01 ~ US-05（数据源配置、组织导入、定时同步、组织架构管理、角色管理）

**前置条件：** 项目已初始化（Vite + React + TS + TailwindCSS + Base UI + TanStack Table + React Hook Form）

---

## Task 0: 工程基础设施补全

**目标：** 补齐路由、状态管理、Mock 服务、测试框架，使后续功能开发有基础可依赖。

**交付物：**

- 安装 `react-router`, `zustand`, `msw`, `vitest`, `@testing-library/react`
- 配置 Vitest（jsdom 环境）
- 搭建 MSW browser worker（开发环境自动启动）
- 创建 Admin Layout 骨架（侧边导航 + 顶部栏 + 内容区）
- 配置路由：`/org/data-source`, `/org/structure`, `/org/roles`
- 创建 API 客户端封装（`src/api/client.ts`）
- 定义所有 P1 相关的 TypeScript 类型（`src/api/types.ts`）

**验收：**

- `npm run dev` 能看到带侧边栏的空壳页面，路由跳转正常
- `npm run build` 无类型错误
- `npx vitest run --passWithNoTests` 通过

---

## Task 1: 数据源凭证配置（US-01）

**目标：** 实现数据源页面的凭证配置区域，支持平台选择、凭证填写、测试连接、搜索验证、保存。

**交付物：**

- 页面：`src/routes/org/data-source.tsx`
- 组件：`src/components/org/credential-form.tsx`
- API：`src/api/org.ts` — `dataSourceApi.getStatus`, `testConnection`, `save`, `searchMember`
- MSW handlers 覆盖上述 API
- 表单使用 React Hook Form，校验逻辑内联

**业务逻辑要点：**

- 平台三选一（飞书/钉钉/企微），选择后动态渲染对应凭证字段
- 测试连接成功后展开"测试区域"，可搜索成员验证字段映射
- 保存按钮仅在测试连接成功后可用
- 切换平台时弹确认；已有凭证修改保存时弹二次确认
- 已连接状态展示"当前数据源：XX，状态：✓ 已连接"

**验收标准：** PRD US-01 场景 1~4

---

## Task 2: 全量导入组织架构（US-02）

**目标：** 在数据源页面（已连接状态）实现"执行全量导入"功能，展示导入结果和失败重试。

**交付物：**

- 组件：`src/components/org/import-result.tsx`
- API：`dataSourceApi.import`, `retryImport`
- MSW handlers（模拟部分成功/部分失败）

**业务逻辑要点：**

- 导入为同步任务，按钮点击后进入 loading 状态
- 完成后展示：成功人数/部门数、失败详情表格（姓名、工号、失败原因）
- 失败项支持单条重试和"全部重试失败"
- 全部成功时提示"是否跳转到组织架构页面查看？"

**验收标准：** PRD US-02 场景 1~4

---

## Task 3: 定时同步策略（US-03）

**目标：** 在数据源页面实现同步策略配置区域和同步记录展示。

**交付物：**

- 组件：`src/components/org/sync-config.tsx`, `src/components/org/sync-log-table.tsx`
- API：`syncApi.getConfig`, `saveConfig`, `triggerSync`, `getLogs`
- MSW handlers

**业务逻辑要点：**

- 同步开关（开启/关闭）
- 配置项：每日开始时间、频率（6/12/24h）、删除保护阈值（成员/部门）、通知方式
- "立即同步一次"按钮
- 同步记录表格：时间、类型（定时/手动）、结果、详情链接
- 表单用 React Hook Form

**验收标准：** PRD US-03 场景 1~5

---

## Task 4: 组织架构管理（US-04）

**目标：** 实现组织架构页面，包含左侧部门树 + 右侧成员表格，支持完整的部门和成员 CRUD。

**交付物：**

- 页面：`src/routes/org/structure.tsx`
- 组件：
  - `src/components/org/department-tree.tsx`（部门树 + 搜索 + 添加/编辑/删除）
  - `src/components/org/member-table.tsx`（TanStack Table，支持多选、分页）
  - `src/components/org/member-form.tsx`（添加/编辑成员表单）
- API：`departmentApi.*`, `memberApi.*`
- MSW handlers
- 共享组件：`src/components/shared/confirm-dialog.tsx`

**业务逻辑要点：**

- 部门树：多级嵌套，可搜索，右键/按钮操作（添加子部门、改名、删除）
- 删除非空部门时阻止并提示
- 成员表格：支持筛选（仅直属/全部）、批量操作（转移部门、启用、停用、删除）
- 未激活成员顶部提示条 + 批量发送激活邀请
- 邀请成员功能（邮箱/手机号）
- 停用成员联动：Platform Key 失效提示

**验收标准：** PRD US-04 场景 1~6

---

## Task 5: 角色权限管理（US-05）

**目标：** 实现角色管理页面，支持角色 CRUD、权限分配、角色成员管理。

**交付物：**

- 页面：`src/routes/org/roles.tsx`
- 组件：
  - `src/components/org/role-list.tsx`（左侧角色列表，分"系统预设"和"自定义"两组）
  - `src/components/org/role-form.tsx`（创建/编辑角色 + 权限点勾选）
  - `src/components/org/role-member-table.tsx`（角色成员列表 + 添加/移除）
- API：`roleApi.*`
- MSW handlers

**业务逻辑要点：**

- 预设角色（5个）不可删除/修改权限，仅可查看和管理成员
- 自定义角色支持完整 CRUD
- 角色成员管理：添加成员（搜索选择）、移除成员
- 普通成员角色为保底，不可移除（前端阻止 + 提示）
- 删除有成员的角色需二次确认，提示影响人数

**验收标准：** PRD US-05 场景 1~5

---

## 任务依赖关系

```
Task 0 (基础设施)
  ├── Task 1 (数据源凭证) ── Task 2 (全量导入) ── Task 3 (定时同步)
  ├── Task 4 (组织架构)
  └── Task 5 (角色管理)
```

Task 1/2/3 是串行的（共享数据源页面，逻辑递进）。Task 4 和 Task 5 可与 Task 1~3 并行开发。

---

## 共享 UI 组件（按需在各 Task 中创建）

| 组件                   | 基于           | 首次需要于 |
| ---------------------- | -------------- | ---------- |
| Button                 | Base UI        | Task 0     |
| Input                  | Base UI        | Task 1     |
| Select                 | Base UI        | Task 1     |
| Dialog / ConfirmDialog | Base UI        | Task 1     |
| Switch                 | Base UI        | Task 3     |
| Toast / Notification   | 自建           | Task 1     |
| Table (封装)           | TanStack Table | Task 2     |
| Tree                   | 自建           | Task 4     |
