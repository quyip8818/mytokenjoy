# Split `support/common` into focused packages

The monolithic `internal/support/common` package (9 files, ~660 lines) has been dissolved into 5 targeted packages: `quota`, `crypto`, `simulate`, `sliceutil`, and additions to the existing `org` package. A helper (`ParseIntParam`) was relocated to `httputil` where it's actually used. The `billing.DefaultQuotaPerUnit()` wrapper function was removed in favor of direct `quota.DefaultQuotaPerUnit` constant access. The `auth` handler constructor was simplified by introducing a `httpdeps.Auth` struct. All 109 changed files compile cleanly with `go build ./...`.

Watch for: deleted test coverage for `Paginate`, `HasAny`, `Encrypt/Decrypt`, `ParseIntParam`, and `simulate.Delayer` (previously in `tests/pkg/common/`) — these unit tests were removed without replacement. (confirmed)

## High-level view

The destination packages map naturally to concern boundaries: `quota` holds billing currency constants and conversion math, `crypto` holds AES-GCM primitives and field encryption, `simulate` holds the dev-mode latency `Delayer`, `sliceutil` holds generic slice utilities (`Paginate`, `HasAny`), and the existing `org` package absorbed the routing logic and org-tree loaders that were already coupled to its types. The split is mechanically sound — every import replacement points to a real exported symbol, and the build confirms no cycles.

`PersistRoutingRules` underwent a signature change: the old version accepted a `RoutingPersistStore` interface wrapping the full store; the new version accepts `AllowlistWriter` and `OrgNodeTreeWriter` separately. Both callers were updated and compile correctly. This is the only API shape change beyond simple import path updates.

The audit filter generics (`FilterByDateRangeCreatedAt`, `FilterByEquals`, `FilterByKeyword`) were deleted entirely. They had zero callers in production code — only the now-deleted tests exercised them. This is dead-code removal, not a missing piece.

The `httpdeps.Auth` struct and `auth.NewHandler(d httpdeps.Auth)` simplification is a bonus refactor bundled into this diff. It moves token-issuer construction from `auth/mount.go` into `deps/auth.go`.

<details>
<summary>Issues (2)</summary>

1. **Deleted unit tests without replacement** — `tests/pkg/common/` (7 files, 432 lines) covering `Paginate`, `HasAny`, `Encrypt`/`Decrypt`, `ParseIntParam`, `simulate.Delayer`, and routing helpers was removed. The routing tests survive in `whitelist_enforcement_test.go`, but the others have no replacement. Add equivalent tests in the new package locations (`tests/pkg/sliceutil/`, `tests/pkg/crypto/`, etc.).
2. **Auth handler refactor bundled with mechanical rename** — The `httpdeps.Auth` struct introduction and `NewHandler` signature change are a behavioral refactor, not a rename. Safe here (build passes, logic identical), but splitting it into its own commit would make bisect/revert cleaner.

</details>

<details>
<summary>Details</summary>

## Deleted test coverage

The following tests were removed with no counterpart in the new package layout:

- `crypto_test.go` — tested `Encrypt`/`Decrypt` round-trip, `ParseKey` edge cases, `DevDefaultKey`.
- `paginate_test.go` — tested page/size bounds, empty input, over-range.
- `has_any_test.go` — tested wildcard `*`, empty required list, partial match.
- `parse_test.go` — tested `ParseIntParam` with empty/invalid/zero values.
- `simulate_test.go` — tested `Delayer.Wait` disabled path and context cancellation.
- `routing_test.go` — tested `ShrinkChildRoutingRules`, `ResolveDeptAllowedModelIDs`, etc.

Of these, only the routing tests survive (relocated to `whitelist_enforcement_test.go`). The rest exercise pure functions in `crypto`, `sliceutil`, `quota`, and `simulate` that now have zero unit-test coverage. Domain-level integration tests hit some paths indirectly but don't isolate edge cases.

## `PersistRoutingRules` signature change

Old: `PersistRoutingRules(ctx, st RoutingPersistStore, nodes, rules)` where `RoutingPersistStore` required `Models() store.ModelsRepository` and `Org() store.OrgRepository`.

New: `PersistRoutingRules(ctx, allowlist AllowlistWriter, orgNodes OrgNodeTreeWriter, nodes, rules)`.

This breaks the coupling between `support/org` and the concrete `store` package. The two callers pass `st.Models().Allowlist()` and `st.Org().Nodes()` explicitly. The change is correct and more explicit about what IO surface the function actually uses.

## Auth handler simplification

`auth/mount.go` previously constructed `registertoken.Issuer` and `invitetoken.Issuer` inline, then passed 10 arguments to `NewHandler`. Now `deps/auth.go` owns that construction in `Deps.Auth()`, and `NewHandler` accepts a single `httpdeps.Auth` struct. The logic (nil invite issuer when `INVITE_SECRET` is unconfigured) is preserved identically. This is a clean refactor but not a pure rename — it changes the construction site and public API of the handler.

## `billing.DefaultQuotaPerUnit()` removal

The function was a one-liner returning `common.DefaultQuotaPerUnit`. Its only caller (`seed/runtime/recharge.go`) now accesses `quota.DefaultQuotaPerUnit` directly. No logic change.

</details>

<details>
<summary>File map (109 files)</summary>

| Path | Change |
|------|--------|
| `internal/support/common/*` | Deleted (9 files) — contents distributed to `quota`, `crypto`, `simulate`, `sliceutil`, `org` |
| `internal/support/quota/quota.go` | New — constants, `MoneyToQuota`, `QuotaToMoney`, `ResolveBillingCurrency` |
| `internal/support/crypto/crypto.go` | New — AES-GCM encrypt/decrypt, `ParseKey`, `EncryptField`/`DecryptField` |
| `internal/support/simulate/delayer.go` | New — `Delayer` struct |
| `internal/support/sliceutil/sliceutil.go` | New — `Paginate`, `HasAny` |
| `internal/support/org/org_store.go` | New — `LoadDepartments`, `LoadBudgetTree`, `LoadRoutingRules`, `PersistRoutingRules` |
| `internal/support/org/routing.go` | New — all routing logic from `common/routing.go` + `ModelNotInDeptMessage`, `NewAPIGroupPrefix` |
| `internal/http/httputil/params.go` | New — `ParseIntParam` |
| `internal/http/deps/auth.go` | New — `Auth` struct and `Deps.Auth()` method |
| `internal/http/handler/auth/handler.go` | `NewHandler` signature changed to accept `httpdeps.Auth` |
| `internal/http/handler/auth/mount.go` | Simplified — delegates to `d.Auth()` |
| `internal/domain/billing/*` | Import `common` → `quota` |
| `internal/domain/budget/*` | Import `common` → `simulate`, `org` |
| `internal/domain/keys/*` | Import `common` → `simulate`, `org` |
| `internal/domain/models/service.go` | Import `common` → `org`, `simulate` |
| `internal/domain/org/**` | Import `common` → `crypto`, `simulate`, `org`, `sliceutil`, `quota` |
| `internal/domain/dashboard/*` | Import `common` → `org` |
| `internal/domain/usage/*` | Import `common` → `org`, `quota`, `sliceutil` |
| `internal/domain/identity/authz/resolve.go` | Import `common` → `sliceutil` |
| `internal/domain/company/service_create.go` | Import `common` → `quota` |
| `internal/store/postgres/*` | Import `common` → `crypto`, `org`, `quota` |
| `internal/integration/newapi/units.go` | Import `common` → `org` |
| `internal/integration/newapisync/platformkey/create.go` | Import `common` → `org` |
| `internal/app/compose_infra.go` | Import `common` → `simulate` |
| `seed/*` | Import `common` → `quota` |
| `tests/**` | Import `common` → corresponding new packages |
| `tests/pkg/common/*` | Deleted (7 test files) |

Full diff: `git diff HEAD` (109 files, +330 / -1459 lines)

</details>
