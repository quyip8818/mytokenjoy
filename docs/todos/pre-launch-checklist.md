# 上线前待办清单

> 创建日期：2026-07-16  
> 这些项需要统一规划实施，不适合零散修复

---

## 必须完成

### 1. Per-tenant API Rate Limiting (SEC-04)

**问题：** 无任何租户级别的请求速率限制，存在 noisy neighbor 和暴力破解风险。

**方案：**
- 基于 Redis 的令牌桶，key = `rate:{companyID}`
- 对登录端点额外加 per-IP 限制
- middleware 层实现，在 CompanyResolve 之后、RequireSession 之前

**涉及文件：**
- `internal/http/middleware/rate_limit.go`（新建）
- `internal/http/router.go`（挂载）
- `internal/config/config.go`（添加限流配置）

**预估：** 2d

---

### 2. JWT 添加 iss/aud 声明 (SEC-07)

**问题：** Token 无 issuer/audience 字段，多服务共享 secret 时存在 token 误用风险。

**方案：**
- `sessiontoken.Claims` 添加 `Issuer: "tokenjoy"`, `Audience: ["tokenjoy-api"]`
- `Parse` 时验证 iss/aud
- Platform token 使用不同的 audience: `["tokenjoy-platform"]`

**涉及文件：**
- `internal/identity/sessiontoken/issuer.go`
- 所有相关测试的 token 生成

**注意：** 发布后旧 token 会立即失效（无 iss/aud），需确保部署时所有用户重新登录。项目未上线所以无影响。

**预估：** 1d

---

### 3. 部署配置检查

确保生产环境：
- [ ] `SESSION_SECRET` ≠ `PLATFORM_SESSION_SECRET`（两个独立随机值）
- [ ] `NEW_API_WEBHOOK_SECRET` 为强随机值（≥32字节）
- [ ] `DATA_SOURCE_CREDENTIAL_KEY` 通过 secrets manager 注入
- [ ] `SECURE_COOKIE=true`
- [ ] `CORS_ORIGINS` 仅包含实际前端域名
- [ ] `DB_MAX_CONNS` 根据实例规格调整（建议：实例CPU×10）

---

## 可选优化（按需）

| 项目 | 触发条件 |
|------|----------|
| RLS defense-in-depth | 多开发者协作时 |
| Webhook secret 轮换 | 需要零停机更换 secret 时 |
| Gateway precheck Redis 缓存 | API QPS > 1000 时 |
| AuthzRevision LISTEN/NOTIFY | TTL 5s 不满足需求时 |
| Per-tenant Prometheus labels | 需要按租户监控时 |
