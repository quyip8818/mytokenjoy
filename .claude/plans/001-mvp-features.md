# TokenJoy MVP Frontend Implementation Plan

## Scope

Implement remaining MVP frontend features per the PRD with mock data. **Excludes** the 3 already-implemented org pages (structure, data-source, roles).

## Features to Build

### 1. Budget Management (`/budget/*`)

- **Budget Overview** (`/budget/overview`) — tree view of org nodes with budget allocation, consumption progress bars, overrun policy badges
- **Budget Allocation** (`/budget/allocation`) — form to allocate budget to org nodes, support for Budget Groups
- **Alert Rules** (`/budget/alerts`) — threshold config (80%/90%/100%), notification target selection

### 2. API-KEY Management (`/keys/*`)

- **Provider Keys** (`/keys/provider`) — table of upstream provider keys (OpenAI, Anthropic, etc.), status, rotation, pool management
- **Platform Keys** (`/keys/platform`) — table of issued Platform Keys with quota, status, bound member/app
- **Approval Flow** (`/keys/approval`) — pending/approved/rejected requests list with approve/reject actions

### 3. Model Routing (`/models/*`)

- **Model List** (`/models/list`) — available models from all providers, enable/disable, cost info
- **Routing Rules** (`/models/routing`) — whitelist config per org node, fallback strategy, inheritance visualization

### 4. Observability (`/dashboard/*`)

- **Cost Dashboard** (`/dashboard/cost`) — total spend, department breakdown pie chart, daily trend line chart, top consumers
- **Usage Analysis** (`/dashboard/usage`) — token consumption by team, model usage distribution, quota progress, member ranking

### 5. Audit (`/audit/*`)

- **Operation Logs** (`/audit/operations`) — table of admin operations (key create/disable, budget change, permission change) with filters
- **Call Logs** (`/audit/calls`) — API call records with caller, model, tokens, latency, expandable I/O content

---

## Technical Approach

### File Structure (new files)

```
src/
├── api/
│   ├── budget.ts        # Budget domain API
│   ├── keys.ts          # API-KEY domain API
│   ├── models.ts        # Model routing API
│   ├── dashboard.ts     # Dashboard/analytics API
│   ├── audit.ts         # Audit log API
│   └── types.ts         # Extended with new interfaces
├── components/
│   ├── budget/          # Budget domain components
│   ├── keys/            # Key management components
│   ├── models/          # Model routing components
│   ├── dashboard/       # Chart widgets, stat cards
│   └── audit/           # Log table components
├── mocks/
│   ├── data.ts          # Extended with new mock data
│   └── handlers.ts      # Extended with new mock endpoints
└── routes/
    ├── budget/
    │   ├── overview.tsx
    │   ├── allocation.tsx
    │   └── alerts.tsx
    ├── keys/
    │   ├── provider.tsx
    │   ├── platform.tsx
    │   └── approval.tsx
    ├── models/
    │   ├── list.tsx
    │   └── routing.tsx
    ├── dashboard/
    │   ├── cost.tsx
    │   └── usage.tsx
    └── audit/
        ├── operations.tsx
        └── calls.tsx
```

### Patterns to Follow

- Page state via `useState` (match existing pattern, no zustand stores)
- API calls via typed API client functions (e.g., `budgetApi.getOverview()`)
- MSW handlers return mock data with `delay()` for realism
- Chinese UI text throughout
- shadcn/ui components (Card, Table, Dialog, Badge, Button, etc.)
- Recharts for charts (simpler API than ECharts, good for dashboards)
- TanStack React Table for data tables with pagination
- react-hook-form for any form inputs

### Routing & Navigation Updates

- Add new nav groups in Sidebar: 预算管理, Key 管理, 模型路由, 数据看板, 审计日志
- Add all new routes to App.tsx under AdminLayout
- Default route stays `/org/structure`

### New shadcn Components Needed

- `progress` (for budget bars)
- `tabs` (for page sub-views)
- `tooltip` (for chart hover info)
- `sheet` (for side detail panels)

---

## Implementation Order

1. **Types & API layer** — Define all new interfaces and API client functions
2. **Mock data & handlers** — Create realistic mock data and MSW handlers
3. **Sidebar & routing** — Add navigation items and route definitions
4. **Dashboard pages** — Cost + Usage (high visual impact, good demo)
5. **Budget pages** — Overview + Allocation + Alerts
6. **API-KEY pages** — Provider + Platform + Approval
7. **Model routing pages** — List + Rules
8. **Audit pages** — Operations + Calls
9. **Add new shadcn components** — progress, tabs, tooltip as needed
