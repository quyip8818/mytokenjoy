# 账户设置 — 待实现功能

## 头像上传

### 交互

```
用户在「基本信息」Tab 点击头像区域
  → 弹出文件选择器（png/jpeg/webp，≤2MB）
  → 前端裁剪为正方形 256×256
  → POST /me/avatar (multipart/form-data, field: "file")
  → 返回 { "url": "..." }
  → 全局头像更新（Header、成员列表等）
```

### API

| Method | Path | 说明 | Response |
|--------|------|------|----------|
| POST | `/me/avatar` | 上传头像 | `{ "url": "https://..." }` |
| DELETE | `/me/avatar` | 删除头像（恢复默认） | 204 |

### 后端

- `users` 表加 `avatar_url TEXT DEFAULT ''`
- 收到文件 → 校验类型+大小 → resize 256×256 webp → 存储
- 存储：本地 `/uploads/avatars/{userID}.webp`（自部署）或 S3（SaaS）
- `/session` 响应的 Member 增加 `avatarUrl` 字段

### 前端

- 「基本信息」Tab 顶部加头像区域（圆形，点击触发上传）
- `<input type="file" accept="image/*">` + canvas 裁剪
- Header 优先显示 `avatarUrl`，fallback 首字母

### 依赖

- Go 图片处理：`disintegration/imaging`
- 前端裁剪：`react-easy-crop` 或原生 canvas
