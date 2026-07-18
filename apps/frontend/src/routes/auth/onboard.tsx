import { Building2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useOnboardPage } from '@/features/auth'

export default function OnboardPage() {
  const {
    mode,
    invites,
    error,
    accepting,
    creating,
    companyName,
    setCompanyName,
    memberName,
    setMemberName,
    handleAcceptInvite,
    handleCreateCompany,
  } = useOnboardPage()

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8">
      <div className="flex w-full max-w-md flex-col gap-6">
        <h1 className="text-center text-lg font-semibold">
          {mode === 'choose' ? '选择如何继续' : '欢迎使用 TokenJoy，选择如何开始'}
        </h1>

        {/* Name input for invite acceptance */}
        {invites.length > 0 ? (
          <div className="space-y-2">
            <Label htmlFor="member-name">您的姓名</Label>
            <Input
              id="member-name"
              placeholder="输入姓名"
              value={memberName}
              onChange={(e) => setMemberName(e.target.value)}
            />
          </div>
        ) : null}

        {/* Pending invites */}
        {invites.length > 0 ? (
          <div className="flex flex-col gap-2">
            {invites.map((invite) => (
              <div
                key={invite.inviteCode}
                className="flex items-center justify-between rounded-lg border px-4 py-3"
              >
                <div className="flex items-center gap-3">
                  <Building2 className="h-5 w-5 text-muted-foreground" />
                  <div className="flex flex-col">
                    <span className="font-medium">{invite.companyName}</span>
                    <span className="text-xs text-muted-foreground">{invite.role}</span>
                  </div>
                </div>
                <Button
                  size="sm"
                  disabled={accepting !== null}
                  onClick={() => handleAcceptInvite(invite.inviteCode)}
                >
                  {accepting === invite.inviteCode ? '接受中…' : '接受邀请'}
                </Button>
              </div>
            ))}
          </div>
        ) : null}

        {/* Divider when there are invites */}
        {invites.length > 0 ? (
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <div className="h-px flex-1 bg-border" />
            <span>或</span>
            <div className="h-px flex-1 bg-border" />
          </div>
        ) : null}

        {/* Create company */}
        <div className="rounded-lg border p-4">
          <div className="mb-3 flex items-center gap-2">
            <Building2 className="h-5 w-5 text-muted-foreground" />
            <span className="font-medium">创建公司</span>
          </div>
          <form onSubmit={handleCreateCompany} className="flex gap-2">
            <Input
              placeholder="公司名称"
              value={companyName}
              onChange={(e) => setCompanyName(e.target.value)}
              required
            />
            <Button type="submit" disabled={creating || !companyName.trim()}>
              {creating ? '创建中…' : '创建'}
            </Button>
          </form>
        </div>

        {/* Placeholder options */}
        {mode === 'onboard' ? (
          <>
            <div className="flex items-center justify-between rounded-lg border px-4 py-3 opacity-50">
              <span className="font-medium">🎯 免费试用</span>
              <span className="text-xs text-muted-foreground">即将推出</span>
            </div>
            <div className="flex items-center justify-between rounded-lg border px-4 py-3 opacity-50">
              <span className="font-medium">👀 Demo 演示</span>
              <span className="text-xs text-muted-foreground">即将推出</span>
            </div>
          </>
        ) : null}

        {error ? <p className="text-sm text-destructive">{error}</p> : null}
      </div>
    </div>
  )
}
