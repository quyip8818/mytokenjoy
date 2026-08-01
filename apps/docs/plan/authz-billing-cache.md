# AuthzSvc Billing 缓存优化

> 来源：auth-system.md 旧 §8（性能改进方向）。**现状：均未实现，仍为有效待办。**

## 当前瓶颈

`ResolveCompanyChargeRate`（`domain/billing/currency.go`）在 LRU 缓存之外，每次请求 2 次 DB 查询（`Company.GetByID` + `Billing.GetCurrency`）。

## P0：将 billing rate 纳入 LRU

billing currency 和 quota_per_unit 是公司级配置，变更频率极低。与 member authz 一起缓存（`authz/service.go` 的 `revisionCache`），以 revision 为失效键。

效果：热路径 DB 查询从 2 降至 0（revision 5s TTL 内）。

## P1：revision 查询走 Redis

当前 `revisionCache` 是进程内内存缓存（`sync.Mutex` + map，TTL 5s），多实例部署下各自维护独立缓存，不共享失效事件。改为 Redis 后，bump 时 `DEL` key 即时对所有实例生效。

## P2：`CompanyType` 缓存

`companyInfoFromContext` fallback 分支和 `CompanyResolve` 中间件仍会打 `Company().GetByID`，未缓存。低优先级，P0 完成后可顺带收掉。
