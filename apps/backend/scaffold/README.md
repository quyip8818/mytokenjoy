# Domain 脚手架

基于 `audit` 域裁剪的最小模板，用于快速新增 bounded context。

## 用法

在 `apps/backend/` 目录下：

```bash
make scaffold-domain DOMAIN=notification
```

`DOMAIN` 必须为小写字母开头的标识符（如 `notification`、`billing`）。

## 生成内容

| 文件                                        | 说明                          |
| ------------------------------------------- | ----------------------------- |
| `internal/domain/<DOMAIN>/service.go`       | Service interface + 空实现    |
| `internal/http/handler/<DOMAIN>/handler.go` | Handler + RegisterRoutes 骨架 |
| `tests/domain/<DOMAIN>/service_test.go`     | 领域单测                      |
| `tests/handler/<DOMAIN>_test.go`            | HTTP 契约测骨架               |

脚本会在终端打印需手动粘贴到以下文件的代码片段：

- `internal/infra/permission/keys.go` — 权限 key
- `internal/app/wiring_domain.go` 与 `registry.go` — DI 注册
- `internal/http/deps/deps.go` — 新增 `__DOMAIN_TITLE__Svc`（完整字段见 `snippets/deps_reference.go.snippet`）
- `internal/http/handler/register.go` — Handler 构造与 `/api` 路由注册

## 不自动生成（需手工）

- `internal/domain/types/` 中的 DTO（与前端契约绑定）
- `store.*Repository` 新方法（`store/memory` + Postgres 双实现）
- 前端 `api/types/` 与契约文档

## 模板目录

```
scaffold/
  domain/     service.go.tmpl, service_test.go.tmpl
  handler/    handler.go.tmpl, handler_test.go.tmpl
  snippets/   注册清单片段（含 deps_reference 完整 Deps 字段参考）
```

占位符：`__DOMAIN__`（包名）、`__DOMAIN_TITLE__`（PascalCase 类型名）。
