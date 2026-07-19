# dev-mock-llm

本地 OpenAI 兼容上游，供 **`test-model`** 全链路 ingest 测试使用。

## 它在测试里做什么

真实 LLM 不可控（tokens / 延迟 / 费用）。本服务**唯一职责**：读请求 body 里的 `dev_usage`，把相同数字写进响应 `usage`。  
NewAPI 据此结算、写 logs、发 webhook；后续 Ingest / ledger / 投影与线上一致。

Gateway 对 `test-model` 仅在 `DEPLOY_ENV=local` 时 Gateway 放行（见 `gateway_service.go`）。

## 启动

`pnpm start` 已内置本服务；单独调试：

```bash
pnpm start:dev-mock   # http://127.0.0.1:8765
```

## 请求示例

```json
{
  "model": "test-model",
  "messages": [{ "role": "user", "content": "ping" }],
  "dev_usage": { "prompt_tokens": 12000000, "completion_tokens": 8000000 }
}
```

## NewAPI channel

Docker 内 NewAPI 指向宿主机 mock：

- 模型：`test-model`
- Base URL：`http://host.docker.internal:8765`（经 `baseurl.Origin` / `verify_http_origin` 规范化）
- **必须**开启 channel `pass_through_body_enabled`：默认 relay 会丢掉非 DTO 字段，`dev_usage` 到不了本服务

```bash
apps/newapi/scripts/setup-dev-mock-channel.sh
```

完整测试流程见 [本地模式-模拟消耗Popup.md](../../docs/manual-testing/本地模式-模拟消耗Popup.md)。
