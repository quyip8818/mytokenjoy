# Backend 测试失败记录

日期：2026-07-18
运行命令：`make test-unit`
Frontend：全部通过（47 files / 156 tests）

## 失败汇总

共 46 个测试失败，分布在 17 个 package 中。

## 失败测试列表

| # | 测试名 | 文件 | 错误摘要 |
|---|--------|------|----------|
| 1 | TestAddRoleMemberNotFoundMember | role_test.go | uuid parse "non-existent-member-id" 失败 |
| 2 | TestApprovalBudgetCheckNotFound | contract_prod_test.go | 422 invalid approval id |
| 3 | TestAuthzWriteEndpoints | authz_test.go | 500 Internal server error |
| 4 | TestBatchInviteByIDs | contract_test.go | 400 memberId is required |
| 5 | TestBootstrapDemoWalletUserCreatesWallet | service_test.go | billing wallet_sync 未入队 |
| 6 | TestBudgetApprovalResolve | approvals_test.go | 422 invalid approval id |
| 7 | TestBudgetNodeUpdateOversell | contract_test.go | 400 invalid departmentId |
| 8 | TestBudgetSummaryIncludesSnapshotConsumed | service_test.go | consumed 值不匹配 |
| 9 | TestDashboardDefaultApp | dashboard_test.go | 500 Internal server error |
| 10 | TestDepartmentDeleteLeafHTTP | org_department_test.go | create 422 |
| 11 | TestDepartmentUpdateHTTP | org_department_test.go | 400 invalid id |
| 12 | TestGatewayAcceptsNewAPIStyleToken | gateway_test.go | 403 model not allowed |
| 13 | TestGatewayAllowsDevModelInLocal | gateway_test.go | 403 model not allowed |
| 14 | TestGatewayBudgetCheckCombinedKeyBlock | budget_check_test.go | Redis GET 缺失 / model not allowed |
| 15 | TestGatewayBudgetCheckDisabledSkipsGet | budget_check_test.go | model not allowed |
| 16 | TestGatewayBudgetCheckMissAllows | budget_check_test.go | model not allowed |
| 17 | TestGatewayCheckOrder_AllowsWhenDeptBudgetZero | call_chain_test.go | 403 dept budget |
| 18 | TestGatewayCheckOrder_SuccessfulProxy | call_chain_test.go | 403 dept budget |
| 19 | TestGatewayProxiesDespiteExhaustedDepartmentBudget | gateway_test.go | 403 |
| 20 | TestGatewayProxiesDespiteZeroDeptBudget | gateway_test.go | 403 |
| 21 | TestGatewayProxiesFullV1Path | gateway_test.go | 403 |
| 22 | TestGatewayProxiesValidRequest | gateway_test.go | 403 |
| 23 | TestGatewaySingleStoreCall | gateway_test.go | 403 |
| 24 | TestGetContractEndpoints | contract_test.go | 400/500 multiple |
| 25 | TestGetContractWithAdminCookie | contract_test.go | 400 invalid departmentId |
| 26 | TestImportCreatesDepartmentsAndMembers | sync_test.go | 部门重命名失败 |
| 27 | TestImportProvisionsBudgetAndRouting | sync_test.go | 路由/mapping 失败 |
| 28 | TestIngestNotifiesOnOverdraftExpansion | — | 通知 overrun 未触发 |
| 29 | TestKeysHTTPEndpoints | keys_handler_test.go | 500 |
| 30 | TestLoadPrecheckContextAllowlistTypes | gateway_precheck_test.go | 白名单不含 claude-sonnet-4-6 |
| 31 | TestMutatingContractEndpoints | mutating_contract_test.go | 500 |
| 32 | TestPlatformRechargeEnqueuesWalletSyncWhenConfigured | service_test.go | wallet_sync 未入队 |
| 33 | TestPrecheckAllowsNullCombinedKeyRemain | precheck_test.go | model not allowed |
| 34 | TestPrecheckPassesRegardlessOfDeptConsumed | precheck_test.go | model not allowed |
| 35 | TestPrecheckPassesWhenNewAPIUnavailable | precheck_test.go | model not allowed |
| 36 | TestRoutingUpdateHTTP | models_test.go | 路由更新失败 |
| 37 | TestSuspendedCompanyBlocksWrites | platform_test.go | panic: malformed HTTP version (URL 格式错误) |
| 38 | TestSyncCreateEnqueuesOutbox | newapi_sync_test.go | outbox entry 缺失 |
| 39 | TestSyncRenamesBudgetAndRouting | sync_test.go | 部门重命名无效 |
| 40 | TestSyncSoftDeletesBelowThreshold | sync_test.go | authz revision 未增 |
| 41 | TestSyncThresholdBlocksDeletion | sync_test.go | authz revision |
| 42 | TestSyncTriggerWithAPIKey | sync_test.go | authz revision |
| 43 | TestTrySyncCreateEnsuresGroupBeforeCreateToken | newapi_sync_test.go | 同步失败 |
| 44 | TestUpdatePlatformKeyProjectMemberBudget | keys_test.go | budget 更新失败 |
| 45 | TestUsageSeriesMinuteFromLedger | usage_test.go | 数据不匹配 |
| 46 | TestUsageSeriesMinuteSuccessMetaHTTP | usage_test.go | 500 |

## 错误模式分类

### 模式 A：Gateway "model not allowed"（~15 个测试）

Gateway precheck 拒绝请求，返回 "model not allowed"。涉及 model whitelist / allowlist 检查逻辑。
可能原因：allowlist 查询未正确加载模型或 gateway precheck 逻辑变更后测试数据未更新。

### 模式 B：测试传入非 UUID 短字符串（~10 个测试）

`invalid departmentId`、`invalid approval id`、`memberId is required` — 测试 fixture 传入 "dept-5"、"approval-1" 等非 UUID 值。
handler 层增加了 UUID 格式校验但测试未同步更新。

### 模式 C：500 Internal Server Error（~5 个测试）

dashboard、contract endpoints 返回 500，可能是 advisory lock 或其他 SQL 问题。

### 模式 D：Org Sync authz revision 未增（~4 个测试）

sync 操作后 authz_revision 未自增。可能是 sync 流程中缺少 bump revision 调用。

### 模式 E：Billing wallet_sync 未入队（2 个测试）

PlatformRecharge 后找不到 wallet_sync job。可能是 river enqueuer 在测试中未正确配置。

### 模式 F：panic malformed HTTP version（1 个测试）

`TestSuspendedCompanyBlocksWrites` — URL 中 company ID 格式化错误导致 httptest.NewRequest panic。

### 模式 G：allowlist 数据不匹配（1 个测试）

`TestLoadPrecheckContextAllowlistTypes` 期望 "claude-sonnet-4-6" 但实际只有 seed 数据里的模型。

## 失败 Package 列表

| # | Package | 失败数 |
|---|---------|--------|
| 1 | tests/domain/gateway | 多 |
| 2 | tests/domain/keys | 多 |
| 3 | tests/domain/newapisync | 2 |
| 4 | tests/domain/newapisync/provision | 1 |
| 5 | tests/domain/org | 4 |
| 6 | tests/domain/usage | 2 |
| 7 | tests/domain/billing | 1 |
| 8 | tests/handler/authz | 1 |
| 9 | tests/handler/budget | 2 |
| 10 | tests/handler/core | 2 |
| 11 | tests/handler/dashboard | 1 |
| 12 | tests/handler/gateway | 多 |
| 13 | tests/handler/keys | 1 |
| 14 | tests/handler/models | 1 |
| 15 | tests/handler/org | 2 |
| 16 | tests/handler/platform | 1 |
| 17 | tests/store/postgres | 1 |
