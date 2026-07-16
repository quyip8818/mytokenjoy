# 多租户（公司）架构审计报告

> 审计日期：2026-07-16  
> 修复日期：2026-07-16  
> 范围：apps/backend + apps/frontend 多租户隔离

---

## 1. 架构概览

共享数据库 + 逻辑隔离模型。`company_id` 行级区分，复合主键 `(company_id, id)`。

```
HTTP → CompanyResolve(JWT→company_id→context)
     → RequireSession(验证+加载权限)
     → Domain(company.CompanyID(ctx))
     → Store(store.CompanyID(ctx) → WHERE company_id=$1)
```

认证体系：成员JWT / 平台管理员JWT / Platform Key(sk-xxx) / Webhook Secret。

---

## 2. 已修复项

| 编号 | 问题 | 修复 |
|------|------|------|
| SEC-02 | CompanyResolve 无条件 fallback 到 LocalCompanyID | SaaS 模式下无 JWT 则不注入 company context |
| SEC-03 / PERF-03 | ResolveFromMember O(C×M) 全表扫描 | `FindMemberCompanyID` 单条 SQL |
| PERF-02 | LRU cache touch O(n) | `container/list` + map，O(1) |
| PERF-01 | Authz 每请求查 GetAuthzRevision | revisionCache TTL 5s |
| PERF-04 | pgxpool 默认连接池 | `DB_MAX_CONNS`/`DB_MIN_CONNS` 环境变量配置 |
| — | store.CompanyID(ctx)=0 无防御 | warn log + `CompanyIDOrZero` 分离 |
| — | 前端 20+ 处重复 try/catch toast | `withErrorToast` 工具函数 |
| — | useBudgetPage 60 行展开 | spread 合并 |

---

## 3. 上线前必须完成

| 编号 | 问题 | 优先级 | 工作量 |
|------|------|--------|--------|
| SEC-04 | Per-tenant API rate limiting | P1 | 2d |
| SEC-07 | JWT 添加 iss/aud 声明 | P2 | 1d |

详见 `docs/todos/pre-launch-checklist.md`

---

## 4. 上线后按需优化

| 编号 | 问题 | 备注 |
|------|------|------|
| SEC-01 | RLS defense-in-depth | 有流量和多开发者后再加 |
| PERF-05 | WalletService 多实例缓存一致性 | 30s TTL 可接受，监控即可 |
| PERF-06 | Gateway precheck Redis 缓存 | 流量上来后再评估 |
| PERF-07 | CompanyContext 短 TTL 缓存 | 主键查询，暂无瓶颈 |
| — | Per-tenant Prometheus metrics | 运维需要时再加 |

---

## 5. 设计优点（保持不变）

- 复合主键天然避免 ID 冲突
- 统一 `ctxcompany` 传播
- `authz_revision` 版本化缓存失效
- Gateway: key_hash → company_id 单 SQL 校验
- Company 暂停 = 只读模式
- Advisory lock 按 company 隔离
- Webhook constant-time compare
- SEC-05 已确认：所有 state-changing 端点都是 POST/PUT/DELETE/PATCH
- SEC-06 已确认：PlatformSessionSecret 独立配置
