# 邮件模板

Resend 模板配置参考。模板 ID 为 kebab-case 别名，在 `channel_email.go` 的 `resolveTemplateID` 中映射。

## company-invite

用途：成员邀请（company_invite / member_invite 共用）

变量：
- `{{inviteUrl}}` — 邀请接受链接
- `{{companyName}}` — 公司名称

模板内容：

```
标题：邀请您加入 {{companyName}}

正文：

您收到一封邀请

邀请您加入 {{companyName}}（由 <a href="https://www.tokenjoy.me/">TokenJoy</a> 提供服务）进行团队协作。

[接受邀请]({{inviteUrl}})

此邀请 7 天内有效。如果您不认识发送者，请忽略此邮件。

此邮件由 <a href="https://www.tokenjoy.me/">TokenJoy</a> 系统自动发送，请勿回复。
```

注意：不需要"或手动输入邀请码"部分。

## verification-code

用途：验证码（登录/注册/重置密码）

变量：
- `{{code}}` — 验证码

## budget-alert

用途：预算告警通知

## overrun-blocked

用途：超支拦截通知

## sync-threshold-exceeded

用途：同步阈值超限通知
