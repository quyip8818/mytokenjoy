# AuthzSvc Billing 缓存优化

> 来源：auth-system.md 旧 §8（性能改进方向）

## 当前瓶颈

`ResolveCompanyChargeRate` 在 LRU 缓存之外，每次请求 2 次 DB 查询（`Company.GetByID` + `Billing.GetCurrency`）。

## P0：将 billing rate 纳入 LRU

billing currency 和 quota_per_unit 是公司级配置，变更频率极低。与 member authz 一起缓存，以 revision 为失效键。

效果：热路径 DB 查询从 2 降至 0（revision 5s TTL 内）。

## P1：revision 查询走 Redis

多实例共享 revision 视图，bump 时 DEL key 即时失效。

## P2：`CompanyType` 缓存

低优先级，P0 完成后可顺带收掉。
