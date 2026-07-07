import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'

interface InviteDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onInvite: (value: string) => Promise<void>
}

export function InviteDialog({ open, onOpenChange, onInvite }: InviteDialogProps) {
  const [value, setValue] = useState('')
  const [sending, setSending] = useState(false)

  const handleSubmit = async () => {
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
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>邀请成员</DialogTitle>
        </DialogHeader>
        <div className="space-y-2">
          <Label className="text-sm text-muted-foreground">
            输入邮箱或手机号，系统将发送激活邀请
          </Label>
          <Input
            type="text"
            placeholder="邮箱或手机号"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleSubmit()
            }}
            autoFocus
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button onClick={handleSubmit} disabled={!value.trim() || sending}>
            {sending ? '发送中...' : '发送邀请'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
