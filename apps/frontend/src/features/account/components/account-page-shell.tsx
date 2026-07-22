import { useCallback, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { PageShell } from '@/components/layout/page-shell'
import { AvatarPicker } from '@/components/ui/avatar-picker'
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
import { useSession } from '@/features/session'
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
      <div className="mx-auto w-full max-w-xl space-y-8">
        {/* Profile Card */}
        <section className="rounded-xl border border-border bg-card shadow-sm">
          <div className="border-b border-border px-5 py-3">
            <h2 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              个人信息
            </h2>
          </div>
          <div className="divide-y divide-border">
            <ProfileEditSection
              profile={profile}
              saving={props.profileSaving}
              error={props.profileError}
              onSave={props.updateProfile}
            />
            <InfoRow
              label="手机号"
              value={profile.phone}
              action={
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 text-xs text-primary hover:text-primary"
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
                  className="h-7 text-xs text-primary hover:text-primary"
                  onClick={() => props.setEmailDialogOpen(true)}
                >
                  修改
                </Button>
              }
            />
          </div>
        </section>

        {/* Companies Card */}
        {profile.companies.length > 0 && (
          <section className="rounded-xl border border-border bg-card shadow-sm">
            <div className="border-b border-border px-5 py-3">
              <h2 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                所属企业
              </h2>
            </div>
            <div className="divide-y divide-border">
              {profile.companies.map((c) => (
                <div key={c.companyId} className="flex items-center gap-3 px-5 py-3">
                  <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-muted text-xs font-medium">
                    {c.companyName.charAt(0)}
                  </div>
                  <div className="flex-1">
                    <p className="text-sm font-medium">{c.companyName}</p>
                    <p className="text-xs text-muted-foreground">{c.role}</p>
                  </div>
                  {c.current && (
                    <span className="rounded-full bg-green-50 px-2 py-0.5 text-[10px] font-medium text-green-700">
                      当前
                    </span>
                  )}
                </div>
              ))}
            </div>
          </section>
        )}

        {/* Security Card */}
        <section className="rounded-xl border border-border bg-card shadow-sm">
          <div className="border-b border-border px-5 py-3">
            <h2 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              安全
            </h2>
          </div>
          <div className="divide-y divide-border">
            <div className="flex items-center justify-between px-5 py-4">
              <div>
                <p className="text-sm font-medium">登录密码</p>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  {profile.hasPassword ? '已设置' : '未设置，设置后可用密码登录'}
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                className="h-7"
                onClick={() => props.setPasswordDialogOpen(true)}
              >
                {profile.hasPassword ? '修改' : '设置'}
              </Button>
            </div>
            <div className="flex items-center justify-between px-5 py-4">
              <div>
                <p className="text-sm font-medium">登出所有设备</p>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  登出除当前外的所有已登录设备
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                className="h-7"
                onClick={() => props.setRevokeDialogOpen(true)}
              >
                登出
              </Button>
            </div>
            <div className="flex items-center justify-between px-5 py-4">
              <div>
                <p className="text-sm font-medium text-destructive">退出当前设备</p>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  退出登录并返回登录页面
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                className="h-7 border-destructive/30 text-destructive hover:bg-destructive/10 hover:text-destructive"
                onClick={props.logout}
              >
                退出
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

function ProfileEditSection({
  profile,
  saving,
  error,
  onSave,
}: {
  profile: { name: string; avatar: string }
  saving: boolean
  error: string | null
  onSave: (params: { name?: string; avatar?: string }) => Promise<boolean>
}) {
  const { member } = useSession()
  const [editingName, setEditingName] = useState(false)
  const [nameValue, setNameValue] = useState(profile.name)

  const handleSaveName = useCallback(async () => {
    const trimmed = nameValue.trim()
    if (trimmed === profile.name) { setEditingName(false); return }
    const ok = await onSave({ name: trimmed })
    if (ok) setEditingName(false)
  }, [nameValue, profile.name, onSave])

  const handleAvatarChange = useCallback(
    (avatar: string) => { onSave({ avatar }) },
    [onSave],
  )

  return (
    <div className="px-5 py-4 space-y-4">
      {/* Avatar + Name row */}
      <div className="flex items-center gap-4">
        <AvatarPicker value={profile.avatar} onChange={handleAvatarChange} size={56} />
        <div className="flex-1 space-y-1">
          {editingName ? (
            <div className="flex items-center gap-2">
              <Input
                value={nameValue}
                onChange={(e) => setNameValue(e.target.value)}
                className="h-8 text-sm"
                placeholder="输入姓名"
                autoFocus
                onKeyDown={(e) => { if (e.key === 'Enter') handleSaveName(); if (e.key === 'Escape') setEditingName(false) }}
              />
              <Button size="sm" variant="ghost" className="h-7 text-xs" onClick={handleSaveName} disabled={saving}>
                {saving ? '…' : '保存'}
              </Button>
              <Button size="sm" variant="ghost" className="h-7 text-xs" onClick={() => { setEditingName(false); setNameValue(profile.name) }}>
                取消
              </Button>
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium">{profile.name || '未设置姓名'}</span>
              <Button size="sm" variant="ghost" className="h-6 text-xs text-primary" onClick={() => { setNameValue(profile.name); setEditingName(true) }}>
                编辑
              </Button>
            </div>
          )}
          {member && (
            <p className="text-xs text-muted-foreground">
              昵称：{member.alias || '—'}
            </p>
          )}
        </div>
      </div>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
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
    <div className="flex items-center justify-between px-5 py-3">
      <div className="flex items-center gap-4">
        <span className="w-14 text-xs text-muted-foreground">{label}</span>
        <span className="text-sm font-medium">{value}</span>
      </div>
      {action}
    </div>
  )
}
