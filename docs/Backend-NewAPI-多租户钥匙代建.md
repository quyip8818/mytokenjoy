# NewAPI 多租户钥匙代建（方案 A · as-built）

| | |
| --- | --- |
| 状态 | **已实施（P0–P2）**；P3 生产错挂修复待维护窗 |
| 方案 | 单一运营 Admin AT + Admin 代建 Token，`user_id` = `newapi_wallet_user_id` |
| Patch | [`0002-admin-token-contract.patch`](../apps/newapi/patches/new-api/0002-admin-token-contract.patch) |
| 上游 | [`UPSTREAM_REF`](../apps/newapi/patches/new-api/UPSTREAM_REF) |

## 行为

```text
CreateCompany → CreateUser → newapi_wallet_user_id = W
Create PlatformKey → CreateToken(user_id=W) → Token.user_id = W
wallet_sync TopUp(W) / UpdateToken(by id) 同属 W
```

| 不变量 | |
| --- | --- |
| I1 | 一企一钱包用户 |
| I2 | sync 后的 Token.`user_id` == 该企业 wallet |
| I3 | 仅一把 Admin AT（env） |
| I4 | 金钱 SSOT 在 Postgres |
| I5 | secret 由 NewAPI 签发 |
| I6 | 管理面不切换企业身份 |
| I7 | 归属错 / 无响应 `id` → Fail-fast，禁止静默挂运营号 |

## 合同（长期）

- Admin：`POST /api/token/` 可读 body `user_id`，校验目标用户存在且可管；响应含 `id` / `user_id` / `key`
- 非 Admin：忽略他人 `user_id`，只能给自己建
- Admin：`GET`/`PUT`/`DELETE`/`regenerate` by id 可跨用户（与 0002 同权级）

实现可换（patch → 上游合入）；Backend 语义不变。

## 代码

| 层 | 路径 |
| --- | --- |
| Patch | `apps/newapi/patches/new-api/0003-…` + Dockerfile apply |
| Client | `integration/newapi/token.go`：`validateCreatedToken`；错归属尽力 Delete |
| Port | `adminport.TokenResult.UserID` |
| Sync | `newapisync/platformkey/create.go`：wallet id 必填；persist 失败 `deleteRemoteTokenBestEffort` |

已删除：Create 后 `findTokenByName`（列 Admin 自有 Token）主路径。

## 兼容与迁移

| 场景 | 做法 |
| --- | --- |
| 旧 NewAPI（无 0002 合同） | Create 无 `id` / 归属错 → **失败**；不降级 |
| `TokenResult.UserID` | 新字段，零值兼容 Update/Rotate 旧调用 |
| Postgres schema | **无** migration |
| 本地错挂 Token | 强制 `build-image.sh` + `pnpm docker:reset` |
| 生产错挂 | 维护窗逐条删旧重建 mapping（P3，无专用 CLI） |

## 验证

```bash
bash apps/newapi/scripts/build-image.sh
cd apps/backend && go test -tags=testhook ./tests/integration/newapi/... ./tests/domain/adminport/... -count=1
```

## 决策

| 日期 | |
| --- | --- |
| 2026-07-14 | 选 A；排除 impersonation / 「全挂运营」终态 |
| 2026-07-14 | P0–P2 落地；文档收为 as-built |
