import { useState } from 'react'
import { FormDialog } from '@/components/ui/form-dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface InviteDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onInvite: (value: string) => Promise<void>
}

export function InviteDialog({ open, onOpenChange, onInvite }: InviteDialogProps) {
  const [value, setValue] = useState('')
  const [sending, setSending] = useState(false)

  function handleOpenChange(v: boolean) {
    if (!v) setValue('')
    onOpenChange(v)
  }

  async function handleSubmit() {
    if (!value.trim()) return
    setSending(true)
    try {
      await onInvite(value.trim())
      setValue('')
      onOpenChange(false)
    } finally {
      setSending(false)
    }
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={handleOpenChange}
      title="邀请成员"
      submitLabel="发送邀请"
      submitDisabled={!value.trim()}
      busy={sending}
      onSubmit={handleSubmit}
      className="sm:max-w-sm"
    >
      <Label className="text-sm text-muted-foreground">输入邮箱或手机号，系统将发送激活邀请</Label>
      <Input
        type="text"
        placeholder="邮箱或手机号"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') void handleSubmit()
        }}
        autoFocus
      />
    </FormDialog>
  )
}
