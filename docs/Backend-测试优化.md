# Backend 测试优化

`apps/backend/tests/` 现状分析与渐进式瘦身方案。目标是在**不降低覆盖率、不牺牲本地零依赖体验**的前提下，减少重复样板、提升可读性与维护效率。

**相关：** [Backend.md](./Backend.md) §5 · [Backend-架构.md](./Backend-架构.md)

---

## 1. 现状

### 1.1 规模

| 区域 | 文件数 | 行数 | 说明 |
| ---- | ------ | ---- | ---- |
| `tests/` 合计 | 155 | ~14,600 | 占 backend Go 总量约 37% |
| `tests/testutil/` | 24 | ~1,700 | 项目自有测试基础设施 |
| `tests/handler/` | 25 | ~2,900 | HTTP 集成测试 |
| `tests/domain/` | 53 | ~5,800 | 领域服务测试 |

### 1.2 技术栈

- 仅使用 Go 标准库：`testing`、`httptest`、`encoding/json`
- **未引入** `testify`、`go-cmp`、`httpexpect` 等第三方断言/HTTP 库
- 已有较完整的 `testutil` 层（store、app、session、relay、saas、worker）

### 1.3 主要痛点

| 痛点 | 表现 | 典型位置 |
| ---- | ---- | -------- |
| HTTP 样板重复 | 每个用例 10–15 行 `NewRequest` + `ServeHTTP` + 状态码判断 | `tests/handler/core/handler_test.go` |
| 手写断言冗长 | `if x != y { t.Fatalf(...) }` 出现 800+ 次 | 全局 |
| 场景 setup 重复 | 相同 `CreateAlert` / `NewRouter` / service 构造写多遍 | `alert_rules_test.go`、`member_test.go` |
| 相似用例未合并 | 3 个 session 401 测试逻辑几乎相同 | `handler_test.go` |
| 结构体比较困难 | 多字段逐个 `if` 判断 | domain 测试 |

### 1.4 不是问题的部分

- `testutil` 基础设施本身（~1,700 行）——这是正确的抽象层，**不应删除**
- memory store + seed 驱动的测试策略——保证 `make test-unit` 零 PG 依赖
- postgres integration tests——唯一验证 SQL 正确性的途径
- 测试/生产比约 0.58:1——对多租户 SaaS 合理，目标不是盲目砍测试

---

## 2. 目标与非目标

### 2.1 目标

- 减少重复样板，预估整体测试代码 **14–20%**（约 2,000–3,000 行）
- 新测试默认使用统一 DSL，旧测试随改动渐进迁移
- 断言失败时输出更可读（diff、响应 body）
- 保持 `make test-unit` / `make test-integration` 行为不变

### 2.2 非目标

- 不切换到 Ginkgo/Gomega BDD 范式
- 不用 testcontainers 替代 memory store（牺牲本地开发体验）
- 不削减 postgres 集成测试或核心业务场景覆盖
- 不追求「换框架砍一半」——集成场景复杂度无法被库消除

---

## 3. 推荐技术选型

### 3.1 依赖 additions

```bash
cd apps/backend
go get github.com/stretchr/testify/require
go get github.com/google/go-cmp/cmp
```

| 库 | 用途 | 使用约定 |
| -- | ---- | -------- |
| `testify/require` | 断言，失败即 `t.FailNow()` | **只用 `require`，不用 `assert`** |
| `google/go-cmp` | 结构体/slice 深度比较 | domain、store roundtrip |
| 项目内 `testhttp.Client` | HTTP 测试 DSL | handler 测试专用 |

### 3.2 不采用的方案

| 方案 | 原因 |
| ---- | ---- |
| Ginkgo + Gomega | 仪式重、与 Go 惯用法冲突，行数常不降反升 |
| httpexpect | 不了解项目 cookie/SaaS/webhook 约定，不如薄封装 `testutil` |
| testify/suite | struct 继承式组织测试，维护成本高 |
| mockery 全量替换 | 现有 hand-rolled mock 够用，迁移 ROI 低 |

---

## 4. 核心改造：HTTP Client DSL

在 `tests/testutil/http/` 新增 fluent Client，封装现有 `NewRouter` / `NewApp` / `AdminCookie` 约定。

### 4.1 API 设计

```go
// tests/testutil/http/client.go

type Client struct {
    t      *testing.T
    router http.Handler
    cookie string
    headers map[string]string
}

// 构造
func NewClient(t *testing.T, opts ...ClientOption) *Client

// 身份
func (c *Client) AsAdmin() *Client
func (c *Client) WithCookie(cookie string) *Client
func (c *Client) WithHeader(key, value string) *Client

// 请求
func (c *Client) GET(path string) *Response
func (c *Client) POST(path string, body any) *Response
func (c *Client) PUT(path string, body any) *Response
func (c *Client) DELETE(path string) *Response

// 响应链
type Response struct { ... }

func (r *Response) AssertStatus(want int) *Response
func (r *Response) AssertContentType(want string) *Response
func (r *Response) DecodeJSON(dst any) *Response
func (r *Response) Body() string
```

### 4.2 改造前后对比

**Before（~15 行）：**

```go
func TestOrgRolesList(t *testing.T) {
    router := testhttp.NewRouter(t)
    req := httptest.NewRequest(http.MethodGet, "/api/org/roles", nil)
    req.Header.Set("Cookie", testhttp.AdminCookie(t))
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)
    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
    var roles []types.Role
    if err := json.NewDecoder(rec.Body).Decode(&roles); err != nil {
        t.Fatal(err)
    }
    if len(roles) != 6 {
        t.Fatalf("expected 6 roles, got %d", len(roles))
    }
}
```

**After（~6 行）：**

```go
func TestOrgRolesList(t *testing.T) {
    var roles []types.Role
    testhttp.NewClient(t).AsAdmin().
        GET("/api/org/roles").
        AssertStatus(http.StatusOK).
        DecodeJSON(&roles)
    require.Len(t, roles, 6)
}
```

### 4.3 SaaS / Webhook 扩展

```go
// SaaS 场景
client := testhttp.NewSaaSClient(t, mock).AsPlatformAdmin()

// Webhook
client := testhttp.NewClient(t, testhttp.WithApp(mutate)).
    WithHeader("X-Webhook-Secret", secret).
    POST("/api/internal/webhooks/newapi-log", body)
```

保留现有 `ServeAuthz` 供 authz 安全测试使用，新用例优先 Client。

---

## 5. 断言规范

### 5.1 require 替换手写 if

```go
// Bad
if err != nil {
    t.Fatal(err)
}
if rule.ID == "" {
    t.Fatal("expected non-empty ID")
}

// Good
require.NoError(t, err)
require.NotEmpty(t, rule.ID)
```

### 5.2 go-cmp 用于结构体

```go
want := []int{80, 90, 100}
if diff := cmp.Diff(want, rule.Thresholds); diff != "" {
    t.Fatalf("thresholds mismatch (-want +got):\n%s", diff)
}

// 忽略不稳定字段
cmp.Diff(want, got, cmpopts.IgnoreFields(types.AlertRule{}, "ID", "CreatedAt"))
```

### 5.3 保留现有 AssertDomainStatus

`testutil.AssertDomainStatus` 已封装 domain error 判断，继续用于业务错误码测试，内部可改为 `require` 实现。

---

## 6. Table-Driven 与 Scenario Fixture

### 6.1 合并相似 HTTP 用例

```go
func TestSessionAuth(t *testing.T) {
    tests := []struct {
        name   string
        cookie string
        want   int
    }{
        {"missing member", "", http.StatusUnauthorized},
        {"invalid token", "tokenjoy_session_member=missing", http.StatusUnauthorized},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := testhttp.NewClient(t)
            if tt.cookie != "" {
                c = c.WithCookie(tt.cookie)
            }
            c.GET("/api/session").AssertStatus(tt.want)
        })
    }
}
```

### 6.2 Domain Scenario Helper

各 bounded context 在 `helpers_test.go` 提供场景构造，避免每个测试重复 setup。

```go
// tests/domain/budget/helpers_test.go

func newBudgetService(t *testing.T) (budget.Service, store.Store) { ... }

func newAlertRule(t *testing.T, svc budget.Service, opts ...func(*types.AlertRule)) types.AlertRule {
    t.Helper()
    rule := types.AlertRule{
        NodeID:     seed.IDDept3,
        NodeName:   "后端组",
        Thresholds: []int{80, 90, 100},
        Enabled:    true,
    }
    for _, opt := range opts {
        opt(&rule)
    }
    created, err := svc.CreateAlert(testutil.Ctx(), rule)
    require.NoError(t, err)
    return created
}
```

```go
func TestDisabledAlertRuleDoesNotTrigger(t *testing.T) {
    svc, _ := newBudgetService(t)
    rule := newAlertRule(t, svc)
    updated, err := svc.UpdateAlert(testutil.Ctx(), rule.ID, types.AlertRule{Enabled: false})
    require.NoError(t, err)
    require.False(t, updated.Enabled)
}
```

### 6.3 命名约定

| 类型 | 命名 | 位置 |
| ---- | ---- | ---- |
| 服务构造 | `newXxxService(t)` | `helpers_test.go` |
| 实体工厂 | `newXxx(t, svc, opts...)` | `helpers_test.go` |
| 完整场景 | `buildXxxScenario(t, opts...)` | `testutil/relay` 风格 |
| 纯数据 | `defaultXxx()` 无 `t` 参数 | 仅当无 error 路径 |

---

## 7. 分模块优化优先级

### P0 — 高收益、低风险（先做）

| 模块 | 文件 | 当前行数 | 手段 | 预估节省 |
| ---- | ---- | -------- | ---- | -------- |
| handler/core | `handler_test.go` | ~290 | Client + table-driven | 120–150 |
| handler/gateway | `webhook_test.go` | ~276 | Client + scenario | 80–100 |
| domain/budget | `alert_rules_test.go` | ~343 | scenario helper + require | 100–130 |
| domain/org | `member_test.go` | ~337 | scenario helper + table-driven | 100–130 |
| domain/audit | `filter_test.go` 等 | ~282 | 合并相似 filter 用例 | 80–100 |

### P1 — 渐进迁移

| 模块 | 手段 |
| ---- | ---- |
| `tests/handler/authz/` | Client + `ServeAuthz` 并存 |
| `tests/handler/billing/` | SaaS Client 封装 |
| `tests/domain/keys/` | go-cmp 比较 key 状态 |
| `tests/store/postgres/` | 共享 roundtrip case 表（memory/pg 同表驱动） |

### P2 — 观望

| 模块 | 说明 |
| ---- | ---- |
| `testutil/saas` | 体积大但功能集中，暂不拆 |
| mockery 生成 | 仅在新接口 mock 复杂时考虑 |

---

## 8. 迁移步骤

### Phase 1：基础设施（1–2 天）

1. `go get` 添加 `testify/require`、`go-cmp`
2. 实现 `tests/testutil/http/client.go` + `response.go`
3. 为 Client 本身写单元测试（`tests/testutil/http/client_test.go`）
4. 更新 [Backend.md](./Backend.md) §5 测试说明

### Phase 2：示范迁移（2–3 天）

1. 改造 `tests/handler/core/handler_test.go` 作为 handler 范本
2. 改造 `tests/domain/budget/alert_rules_test.go` 作为 domain 范本
3. PR review 确认风格后固化为约定

### Phase 3：渐进铺开（持续）

- **规则：触达即迁移**——修改某测试文件时，顺手改为新风格
- **规则：新测试必须用新风格**——禁止新增手写 httptest 样板
- 每个 PR 控制迁移范围，避免 mega-refactor

### 验收标准

```bash
cd apps/backend
make test-unit
make test-integration   # 有 DATABASE_URL 时
make lint
```

全部通过；迁移前后测试用例数量不减少。

---

## 9. 目录与文件变更

```
tests/testutil/http/
├── http.go          # 现有：NewRouter、NewApp、ServeAuthz
├── client.go        # 新增：Client DSL
├── response.go      # 新增：Response 链式断言
└── client_test.go   # 新增：Client 自测

tests/domain/<域>/
└── helpers_test.go  # 强化：scenario factory

go.mod               # 新增 testify、go-cmp
```

---

## 10. 反模式清单

| 反模式 | 正确做法 |
| ------ | -------- |
| 每个测试 `testhttp.NewRouter(t)` + 10 行 httptest | `testhttp.NewClient(t).AsAdmin().GET(...)` |
| `assert.Equal` 失败后继续执行 | 使用 `require.Equal` |
| 在 `_test.go` 里构造完整 app 逻辑 | 走 `testutil.NewTestApp` / `NewClient` |
| 为减行数删除边界用例 | 合并为 table-driven，保留覆盖 |
| 引入 Ginkgo 重写全部测试 | 保持 `testing` + `t.Run` |
| 在 `internal/` 新增 `*_test.go` | 遵守 `tests/` 外挂约定 |

---

## 11. 预期收益

| 手段 | 适用面 | 预估减行 | 可读性提升 |
| ---- | ------ | -------- | ---------- |
| HTTP Client DSL | handler ~2,900 行 | 800–1,200 | 高 |
| testify/require | 全局 | 500–800 | 中 |
| table-driven | handler、audit | 400–600 | 高 |
| scenario helper | domain | 600–900 | 高 |
| go-cmp | domain、store | 200–400 | 中 |
| **合计** | ~14,600 行 | **2,000–3,000（14–20%）** | — |

---

## 12. 不可削减的测试资产

以下代码是正确投资，优化时应保留并复用：

- `testutil.NewMemoryStoreFromConfig` — memory store 零依赖路径
- `testutil/relay.BuildGatewayScenario` — relay 集成场景
- `testutil/saas.NewAPIMock` — NewAPI 边界 mock
- `testutil/worker` — ingest/outbox 异步断言
- `tests/store/postgres/*_test.go` — SQL 正确性保障
- `tests/handler/core/contract_test.go` — GET 端点契约回归

---

## 13. 检查清单

新增或修改测试时：

- [ ] 使用 `tests/` 目录，不在 `internal/` 写测试
- [ ] Handler 测试使用 `testhttp.Client`，不手写 httptest 样板
- [ ] 断言使用 `require` / `go-cmp`，避免 `if + t.Fatalf`
- [ ] 相似用例合并为 `t.Run` + table-driven
- [ ] 重复 setup 提取到 `helpers_test.go`
- [ ] `t.Helper()` 标记所有 testutil 辅助函数
- [ ] `make test-unit` 通过
- [ ] 新 GET 端点更新 `contract_test.go`
