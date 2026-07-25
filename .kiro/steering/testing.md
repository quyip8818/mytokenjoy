---
inclusion: fileMatch
fileMatchPattern: "**/*test*,**/*spec*,**/tests/**,**/Makefile"
---

# 测试规范

## 命令

```bash
# Apps (TokenJoy)
pnpm test                    # apps 全量（frontend vitest + backend go test）
pnpm test:integration        # apps 后端集成测试
pnpm test:e2e                # apps 前端 Playwright E2E
pnpm lint                    # apps lint

# SMS
pnpm test:sms                # sms 全量（frontend vitest + backend go test）
pnpm test:sms:integration    # sms 后端集成测试
pnpm test:sms:e2e            # sms 前端 Playwright E2E
pnpm lint:sms                # sms lint

# 单文件
cd apps/backend && go test -tags=testhook -run "TestXxx" ./tests/domain/xxx/ -v -count=1
cd sms/backend && go test -run "TestXxx" ./tests/domain/xxx/ -v
```

## 后端核心规则（apps/backend）

1. **改了 schema.sql 或 seed/ → 必须 bump `testTemplateVersion`**（`tests/testutil/pg/template.go`）
2. **时钟固定**：测试默认 `ClockAnchor = "2026-06-19"`，period = `"2026-06"`。不要用 `time.Now()` 断言
3. **Build tag**：必须加 `-tags=testhook`
4. **测试文件放 `apps/backend/tests/`**，镜像 internal/ 路径，外部测试包
5. **DATABASE_URL**：`postgres://tokenjoy:tokenjoy@127.0.0.1:5510/tokenjoy`

## 后端核心规则（sms/backend）

1. **测试文件放 `sms/backend/tests/`**，镜像 internal/ 路径
2. **DATABASE_URL**：`postgres://sms:sms@127.0.0.1:5510/sms`
3. 无需 build tag

## 前端核心规则

1. Vitest + @testing-library/react
2. apps 测试放 `apps/frontend/tests/`，sms 放 `sms/frontend/tests/`
3. 镜像 src/ 路径
4. API 层用 `vi.mock` mock，不要真实请求
5. E2E 用 Playwright

## 端口速查

| 服务 | apps | sms |
|------|------|-----|
| Postgres | 5510 | 5510（共用） |
| Redis | 6310 | 6310（共用） |
| Backend | 8010 | 8020 |
| Frontend | 5173 | 5174 |
| NewAPI | 3010 | 3020 |
