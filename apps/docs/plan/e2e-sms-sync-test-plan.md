# SMS → TokenJoy 模型同步 E2E 测试方案

## 测试目标

验证从 SMS 配置模型/渠道/价格 → OAuth2 认证 → TokenJoy 定时拉取 → NewAPI 写入 → 前端模型列表展示的完整链路。

## 测试分层

### Layer 1: API 层 E2E（纯接口验证）

验证后端 API 的正确性，不涉及浏览器。使用 Playwright 的 `request` API 直接发 HTTP 请求。

### Layer 2: 浏览器 E2E（前端展示验证）

验证 TokenJoy 前端模型列表正确展示同步过来的模型数据。

---

## Layer 1: API 层测试用例

### 1.1 SMS OAuth2 Token 签发

| # | 用例 | 验证点 |
|---|------|--------|
| 1 | 有效 client_credentials → 200 + JWT | access_token 非空, expires_in=600, scope=sync:read |
| 2 | 错误 secret → 401 | 返回 401, message 包含 invalid |
| 3 | 不存在的 client_id → 401 | 返回 401 |
| 4 | 错误 grant_type → 400 | 返回 400, message 包含 unsupported |

### 1.2 SMS Sync Catalog API

| # | 用例 | 验证点 |
|---|------|--------|
| 5 | 有效 token 请求 catalog → 200 | 返回 JSON, 包含 channels[] 和 models[] |
| 6 | 无 Authorization header → 401 | 返回 401 |
| 7 | 过期 token → 401 | 返回 401 |
| 8 | catalog 数据结构完整性 | models[] 每项有 modelId, displayName, provider, callType, inputPrice, outputPrice |
| 9 | channels[] 每项有 name, type, baseUrl, key, models, group | 字段校验 |

### 1.3 TokenJoy 同步触发后验证

| # | 用例 | 验证点 |
|---|------|--------|
| 10 | 手动触发同步 → TokenJoy models API 返回 SMS 模型 | source=sms 的模型出现在列表 |
| 11 | TokenJoy-NewAPI pricing API 包含同步的定价 | ModelRatio 和 CompletionRatio 有对应 key |
| 12 | SMS 不可达时同步不影响现有模型 | 模型列表不变，无 500 报错 |

---

## Layer 2: 浏览器 E2E 测试用例

### 2.1 TokenJoy 模型列表展示

| # | 用例 | 验证点 |
|---|------|--------|
| 13 | 登录 TokenJoy → 导航到模型列表页 | 页面正常加载，表格可见 |
| 14 | 同步的模型在列表中展示 | 模型名称、提供商、价格列正确 |
| 15 | 模型来源标记正确 | source=sms 的模型有"SMS 同步"标识（如有实现） |
| 16 | 手动创建的模型不被同步覆盖 | 手动创建 → 触发同步 → 验证手动模型仍在 |

---

## 测试环境前置条件

```
1. SMS 后端运行在 localhost:8020
2. TokenJoy 后端运行在 localhost:8010
3. TokenJoy-NewAPI 运行在 localhost:3010
4. TokenJoy 前端运行在 localhost:5173（或 4173 preview）
5. SMS 数据库中已有 oauth_clients 记录（client_id=tokenjoy-sync）
6. SMS 数据库中已有模型和渠道数据
7. TokenJoy .env 中 SMS_SYNC_ENABLED=true
```

## 测试文件结构

```
apps/frontend/e2e/
├── sms-sync-api.spec.ts         # Layer 1: API 层
└── sms-sync-models.spec.ts      # Layer 2: 浏览器层
```

## 技术实现要点

### API 层测试
- 使用 `test.use({ baseURL: 'http://localhost:8020' })` 指向 SMS
- 先调 `/api/oauth/token` 获取 token
- 带 token 调 `/api/sync/catalog` 验证响应
- 切换 baseURL 到 TokenJoy 验证同步后的模型列表

### 浏览器层测试
- 复用现有 global-setup.ts 的 admin 登录 session
- 导航到 `/models/list`
- 使用 `getByRole` / `getByText` 验证表格内容
- 等待同步完成（轮询或手动触发）

### 测试数据管理
- `beforeAll`: 通过 SMS API 创建测试模型（如 `e2e-test-model`）
- `afterAll`: 清理测试模型
- 使用唯一前缀避免与其他测试冲突

---

## 执行顺序

1. 先跑 Layer 1（API 层）— 确保接口正确
2. 再跑 Layer 2（浏览器层）— 确保 UI 展示正确
3. Layer 1 失败则 Layer 2 不执行（依赖关系）

## 预期总用例数

| 层 | 用例数 |
|----|--------|
| API 层 | 12 |
| 浏览器层 | 4 |
| **总计** | **16** |
