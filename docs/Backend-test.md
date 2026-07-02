# TokenJoy Backend 测试指南

`apps/backend` 测试规范与运行方式。设计背景见 [Backend-设计.md](./Backend-设计.md)。

---

## 1. 目录约定

**所有测试在 `apps/backend/tests/`，`internal/` 禁止 `*_test.go`。**

```
tests/
├── testutil/              # bootstrap、assert、mock、saas、gateway_scenario
├── handler/               # HTTP 契约 + 行为（含 onboarding、gateway、platform、billing）
├── domain/
│   ├── audit/、budget/、keys/、models/、org/、session/、dashboard/
│   ├── company/、billing/、platform/、renter/、usage/
├── pkg/
│   ├── budget/、common/、newapi/
├── infra/
│   ├── permission/
│   └── notification/
├── integration/datasource/feishu/
├── store/
│   ├── seed/
│   ├── memory/
│   └── postgres/          # integration tag
├── worker/
└── config/
```

| 规则         | 说明                                                                                                            |
| ------------ | --------------------------------------------------------------------------------------------------------------- |
| 包名         | `package <name>_test`，黑盒 import `internal/...`                                                               |
| Fixture      | `testutil.TestConfig()` + `testutil.NewMemoryStore(t, cfg)`                                                     |
| Handler 集成 | `testutil.NewTestApp(t)`（`-tags=testhook`；`app.NewWithStore` + `store/memory`）                               |
| SaaS         | `testutil.ApplySaaSConfig`、`saas.go` helper；见 [Backend-SaaS多租户架构.md](./Backend-SaaS多租户架构.md) §十一 |
| 确定性       | `SimulateDelay=false`；单测不走 Postgres；`config.Load()` 需 `DATABASE_URL`                                     |

---

## 2. 运行

```bash
cd apps/backend
make test-unit                     # go test -tags=testhook ./tests/...
make test-integration              # -tags=integration（需 DATABASE_URL）
go test ./tests/domain/keys/... -v # 单包
```

根目录 `pnpm test` = 前端 Vitest + `make test-unit`。Postgres 集成：`pnpm test:integration` 或 CI `backend-integration` job。

---

## 3. 分层

| 层             | 目录                                   | CI                    |
| -------------- | -------------------------------------- | --------------------- |
| L1 纯函数      | `tests/pkg/budget`、`tests/pkg/common` | `verify`              |
| L2 Domain      | `tests/domain/*`                       | `verify`              |
| L3 Handler     | `tests/handler/*`                      | `verify`              |
| L4 Integration | `tests/integration/*`                  | 手工 / 按需           |
| L5 Postgres    | `tests/store/postgres`                 | `backend-integration` |
| L6 Relay 全栈  | `apps/newapi/scripts/gate-verify.sh`   | 手工                  |

**契约测试：** `contract_test.go`（demo profile + admin cookie）；`contract_prod_test.go`（prod 无 cookie → 401）；`authz_test.go`（写操作 401/403）。

---

## 4. 编写模板

### Domain Service

```go
func newKeysService(t *testing.T) (keys.Service, store.Store) {
    t.Helper()
    cfg, st := testutil.NewMemoryStoreFromConfig(t)
    lifecycle := relay.NewTokenLifecycle(cfg, st, nil)
    return keys.NewService(cfg, st, lifecycle), st
}
```

### Handler

```go
router := testutil.NewTestRouter(t)
req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Cookie", testutil.SessionCookie(seed.IDMemberAdmin))
```

### 新 GET 端点

在 `contract_test.go` 的 `getContractCases()` 追加 `{name, path}`。prod 行为用 `testutil.WithProfile(config.ProfileProd)`。

---

## 5. 与前端边界

|      | Backend                                      | Frontend           |
| ---- | -------------------------------------------- | ------------------ |
| 工具 | Go testing                                   | Vitest             |
| Mock | `testutil/mock`                              | `createMockApis()` |
| 契约 | [Frontend-API契约.md](./Frontend-API契约.md) | 同源               |

默认 CI 不启动 backend 进程。

---

## 6. PR 自检

- [ ] 测试在 `tests/`，未在 `internal/` 新增 `*_test.go`
- [ ] `make test-unit` 通过
- [ ] 使用 `testutil.TestConfig()`，未裸调 `config.Load()`（`tests/store/seed/loader_test` 除外，须设 `DATABASE_URL`）
- [ ] 新 GET → `contract_test.go`；新写操作 → handler 或 domain 用例
- [ ] 跨域逻辑在 domain 层断言 Store 副作用
