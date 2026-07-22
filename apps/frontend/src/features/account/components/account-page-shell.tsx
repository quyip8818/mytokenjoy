import { Button } from '@/components/ui/button'
import { PageShell } from '@/components/layout/page-shell'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import type { AccountPageState } from '../hooks/use-account-page'
import { ChangePasswordDialog } from './change-password-dialog'
import { ChangeContactDialog } from './change-contact-dialog'

export function AccountPageShell(props: AccountPageState) {
  const { profile, profileLoading } = props

  if (profileLoading) {
    return (
      <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">
        加载中…
      </div>
    )
  }

  if (!profile) {
    return (
      <div className="flex h-40 items-center justify-center text-sm text-destructive">
        加载失败
      </div>
    )
  }

  return (
    <PageShell description={<h1 className="text-sm font-semibold">账户设置</h1>}>
      <div className="max-w-2xl space-y-6">
        {/* Profile */}
        <section>
          <h2 className="mb-3 text-xs font-medium text-muted-foreground">基本信息</h2>
          <div className="rounded-lg border border-border divide-y divide-border">
            <InfoRow label="姓名" value={profile.name} />
            <InfoRow
              label="手机号"
              value={profile.phone}
              action={
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-xs text-primary"
                  onClick={() => props.setPhoneDialogOpen(true)}
                >
                  修改
                </Button>
              }
            />
            <InfoRow
              label="邮箱"
              value={profile.email}
              action={
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-xs text-primary"
                  onClick={() => props.setEmailDialogOpen(true)}
                >
                  修改
                </Button>
              }
            />
          </div>
        </section>

        {/* Companies */}
        {profile.companies.length > 0 && (
          <section>
            <h2 className="mb-3 text-xs font-medium text-muted-foreground">所属企业</h2>
            <div className="rounded-lg border border-border divide-y divide-border">
              {profile.companies.map((c) => (
                <div key={c.companyId} className="flex items-center gap-2 px-4 py-3 text-sm">
                  <span className="font-medium">{c.companyName}</span>
                  <span className="text-xs text-muted-foreground">{c.role}</span>
                  {c.current && (
                    <span className="rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-medium text-primary">
                      当前
                    </span>
                  )}
                </div>
              ))}
            </div>
          </section>
        )}

        {/* Security */}
        <section>
          <h2 className="mb-3 text-xs font-medium text-muted-foreground">安全</h2>
          <div className="rounded-lg border border-border divide-y divide-border">
            <div className="flex items-center justify-between px-4 py-3">
              <div>
                <p className="text-sm font-medium">登录密码</p>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  {profile.hasPassword ? '已设置' : '未设置，设置后可用密码登录'}
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => props.setPasswordDialogOpen(true)}
              >
                {profile.hasPassword ? '修改' : '设置'}
              </Button>
            </div>
            <div className="flex items-center justify-between px-4 py-3">
              <div>
                <p className="text-sm font-medium">登出所有设备</p>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  登出除当前外的所有已登录设备
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => props.setRevokeDialogOpen(true)}
              >
                登出
              </Button>
            </div>
          </div>
        </section>
      </div>

      {/* Dialogs */}
      <ChangePasswordDialog
        open={props.passwordDialogOpen}
        onOpenChange={props.setPasswordDialogOpen}
        hasPassword={profile.hasPassword}
        error={props.passwordError}
        saving={props.passwordSaving}
        onSubmit={props.changePassword}
      />

      <ChangeContactDialog
        open={props.phoneDialogOpen}
        onOpenChange={props.setPhoneDialogOpen}
        type="phone"
        error={props.phoneError}
        saving={props.phoneSaving}
        onSubmit={props.changePhone}
      />

      <ChangeContactDialog
        open={props.emailDialogOpen}
        onOpenChange={props.setEmailDialogOpen}
        type="email"
        error={props.emailError}
        saving={props.emailSaving}
        onSubmit={props.changeEmail}
      />

      <AlertDialog open={props.revokeDialogOpen} onOpenChange={props.setRevokeDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>登出所有设备</AlertDialogTitle>
            <AlertDialogDescription>
              确认后将登出除当前设备外的所有已登录会话，此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={props.revokeSessions} disabled={props.revoking}>
              {props.revoking ? '处理中…' : '确认登出'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </PageShell>
  )
}

function InfoRow({
  label,
  value,
  action,
}: {
  label: string
  value: string
  action?: React.ReactNode
}) {
  return (
    <div className="flex items-center justify-between px-4 py-3">
      <div className="flex items-center gap-4">
        <span className="w-14 text-xs text-muted-foreground">{label}</span>
        <span className="text-sm">{value}</span>
      </div>
      {action}
    </div>
  )
}
