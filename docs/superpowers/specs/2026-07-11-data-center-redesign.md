# 数据中心页面重设计 — 设计规格

## 概述

重新设计「成本看板」和「用量分析」两个独立页面，引入组织架构树侧边栏，让管理者能按团队维度快速定位费用去向。

**范围：** 纯前端重设计，后端 API 不变。  
**目标用户：** 拥有 `DASHBOARD_COST` / `DASHBOARD_USAGE` 权限的管理者。

---

## 1. 布局结构

### 1.1 共享模式

两个页面采用相同的 master-detail 布局：

```
┌─────────────────────────────────────────────────────┐
│  App Shell (existing nav sidebar)                   │
├──────────┬──────────────────────────────────────────┤
│ Org Tree │  Content Area                            │
│ (260px)  │  ┌─────────────────────────────────────┐ │
│          │  │ Header: title + period filter        │ │
│ 全公司   │  ├─────────────────────────────────────┤ │
│  工程部  │  │ KPI Stats Row (4 cards)             │ │
│   后端组 │  ├─────────────────────────────────────┤ │
│   前端组 │  │ Charts Row (5:3 grid)               │ │
│   AI 组  │  ├─────────────────────────────────────┤ │
│  产品部  │  │ Table 1: Sub-dept comparison        │ │
│  设计部  │  ├─────────────────────────────────────┤ │
│  运营部  │  │ Table 2: Detail table               │ │
│          │  └─────────────────────────────────────┘ │
└──────────┴──────────────────────────────────────────┘
```

### 1.2 组织架构树侧边栏

- 固定宽度 260px，不可折叠
- 数据源：复用现有 org structure API（`/api/org/tree`）
- 树节点按用户权限裁剪（只展示有权限查看的部门）
- 根节点为「全公司」，选中时展示全局汇总
- 选中父节点时：右侧展示该节点汇总 + 直接子节点对比
- 选中叶子节点时：右侧展示该部门汇总 + 成员明细
- 当前选中部门存入 URL search param `?dept=<departmentId>`
- 默认选中根节点（无 `dept` param 时）

### 1.3 路由

保持现有路由不变：
- `/dashboard/cost` → 成本看板
- `/dashboard/usage` → 用量分析

两个页面各自独立管理 `?dept=` 状态，不共享。

---

## 2. 成本看板 (`/dashboard/cost`)

### 2.1 页面头部

- 面包屑：显示当前选中节点路径（如「全公司 > 工程部 > 后端组」）
- 标题：「成本看板」
- 时间筛选器：沿用现有 period 切换（本月 / 上月 / 近7天 / 自定义）

### 2.2 KPI 指标行

4 张 StatCard，横向等分：

| 指标 | 数据来源 | 环比 |
|------|----------|------|
| 总费用 (CNY) | `getCostSummary().totalCost` | `totalCostMom` |
| 人均费用 | `getCostSummary().avgCostPerMember` | `avgCostPerMemberMom` |
| 总请求数 | `getCostSummary().totalRequests` | `totalRequestsMom` |
| 单请求费用 | `getCostSummary().avgCostPerRequest` | `avgCostPerRequestMom` |

当选中特定部门时，API 增加 `departmentId` 参数（需前端在请求中传递，后端已支持通过部门范围过滤）。

### 2.3 图表行

5:3 网格，两张图表：

**左：费用趋势折线图**
- 数据源：`getDailyCosts(period, granularity, dept)`
- X 轴为日期，Y 轴为费用
- Tooltip 展示：日期 + 费用 + 请求数
- 沿用现有 `CostTrendChart` 组件

**右：子部门费用占比环形图**
- 数据源：`getDepartmentCosts(period, parentId=selectedDept)`
- 展示选中节点直接子部门的费用分布
- 当选中叶子节点时，此图替换为「成员费用占比」（使用 `getDepartmentMemberCosts`）
- 沿用现有 `CostDistributionChart`，调整数据源

### 2.4 子部门费用对比表

- 数据源：`getDepartmentCosts(period, parentId=selectedDept)`
- 列：排名 | 部门名 | 费用 | 占比 | 人均费用（需前端计算：费用 / 该部门成员数）| 环比趋势
- 点击行可在左侧树中选中该部门（等效于树导航）
- 当选中叶子节点时，切换为「成员费用明细表」：成员名 | 费用 | 请求数 | Token 数

### 2.5 Top 消费者表

- 数据源：`getTopConsumers(period, limit=5, dept)`
- 列：排名 | 成员 | 部门 | 费用 | 请求数 | 主力模型
- 「主力模型」字段：当前 API 不返回此字段，前端暂不展示或标记为 TODO
- 展示选中范围内费用 Top 5 的成员

---

## 3. 用量分析 (`/dashboard/usage`)

### 3.1 页面头部

同成本看板结构：面包屑 + 标题 + 时间筛选器。

### 3.2 KPI 指标行

4 张 StatCard：

| 指标 | 数据来源 | 说明 |
|------|----------|------|
| 总调用量 | `getModelUsage()` 求和 requests | 环比 |
| 活跃成员 | `getTeamUsage()` 求和 memberCount / 组织总人数 | 活跃率 |
| 使用模型数 | `getModelUsage().length` | 计数 |
| 预算使用率 | `getTeamUsage()` 求和 consumed / quota | 剩余金额 |

### 3.3 图表行

5:3 网格：

**左：调用量趋势折线图**
- 数据源：`getUsageSeries(granularity=day, start, end, groupBy=none, dept)`
- X 轴日期，Y 轴调用次数
- Tooltip：日期 + 调用量 + 费用
- 沿用现有 `UsageSeriesChart`

**右：模型费用分布横向柱状图**
- 数据源：`getModelUsage(period, dept)`
- 按费用降序排列
- 沿用现有 `UsageModelChart`

### 3.4 子部门用量与配额表

- 数据源：`getTeamUsage(period, dept)`
- 列：部门 | 配额 | 已用 | 使用率（进度条）| 成员数 | 主力模型 | 日均调用
- 进度条颜色：≥90% 红色，≥70% 橙色，<70% 绿色
- 日均调用 = 已用请求数 / 当前天数（前端计算）
- 点击行同样可导航到该部门

### 3.5 模型使用明细表

- 数据源：`getModelUsage(period, dept)`
- 列：模型 | 供应商 | 调用次数 | 费用 | 费用占比 | Token 消耗
- Token 格式化为 `xM`（沿用现有 `formatTokenCount`）

---

## 4. 组件架构

### 4.1 新增共享组件

```
features/dashboard/components/
  org-tree-sidebar.tsx        — 组织架构树侧边栏
  dashboard-page-layout.tsx   — 左右分栏布局壳（sidebar + content slot）
  stat-card-row.tsx           — KPI 指标行容器
```

### 4.2 新增 Hook

```
features/dashboard/hooks/
  use-dept-selection.ts       — 管理 ?dept= URL param 状态
  use-org-tree.ts             — 获取组织树数据 + 权限过滤
```

### 4.3 改造现有组件

| 组件 | 改动 |
|------|------|
| `cost-dashboard-page-shell.tsx` | 改用 `DashboardPageLayout` 包裹，接收 `selectedDeptId` |
| `usage-dashboard-page-shell.tsx` | 同上 |
| `cost-drill-table.tsx` | 简化为纯排名表（不再需要钻取按钮，钻取由树导航替代）|
| `cost-distribution-chart.tsx` | 改为环形图样式 |
| `use-cost-dashboard-page.ts` | 移除 `DrillState`，改为接收 `deptId` prop |
| `use-usage-dashboard-page.ts` | 增加 `deptId` 参数传递给所有 API 调用 |

### 4.4 组织树数据源

复用现有 `departmentApi.getTree()`（`GET /api/org/departments/tree`），返回 `Department[]` 树结构。该接口已在组织架构页面使用，数据结构已验证。

---

## 5. 数据流

```
URL ?dept=xxx
     │
     ▼
use-dept-selection.ts  ──→  org-tree-sidebar (高亮 active 节点)
     │
     ▼
use-cost-dashboard-page.ts / use-usage-dashboard-page.ts
     │
     ├─→ getCostSummary({ period, dept })
     ├─→ getDailyCosts({ period, granularity, dept })
     ├─→ getDepartmentCosts({ period, parentId: dept })
     └─→ getTopConsumers({ period, dept })
```

点击树节点 → 更新 URL `?dept=newId` → hook 响应变化 → 重新请求数据。

---

## 6. 交互细节

### 6.1 树节点点击

- 更新 URL search param（保持其他 param 如 period 不变）
- 如果节点有子节点，自动展开该节点
- 右侧内容区所有数据刷新（带 loading skeleton）

### 6.2 Loading 状态

- 首次加载：整个右侧展示 skeleton
- 切换部门：仅数据区域 skeleton，保持 header 不变
- 组织树独立加载，不阻塞内容区

### 6.3 空状态

- 选中部门无数据时，展示空状态插图 + 「该部门暂无使用数据」
- 无权限部门不出现在树中（后端已按权限过滤）

### 6.4 响应式

- 最小宽度 1024px（含 app 导航 sidebar + org tree + content）
- 如果视口不足，org tree 可考虑叠加在 content 上方（但 MVP 不处理）

---

## 7. 不在范围内

- 后端 API 改动
- 导出功能
- 自定义图表维度切换
- 移动端适配
- 成本看板「主力模型」字段（当前 API 不支持）

---

## 8. 技术依赖

- 现有 recharts 图表库
- 现有 shadcn/ui 组件
- 现有 org structure API 数据
- react-router v7 search params
