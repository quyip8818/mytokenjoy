import { useState, useEffect, useCallback } from 'react'
import { AlertTriangle, Send, X } from 'lucide-react'
import { toast } from '@/lib/toast'
import { Button } from '@/components/ui/button'
import { orgApi } from '@/api/org'

interface PendingActivationBannerProps {
  pendingCount: number
}

const COOLDOWN = 90 // seconds

export function PendingActivationBanner({ pendingCount }: PendingActivationBannerProps) {
  const [dismissed, setDismissed] = useState(false)
  const [remaining, setRemaining] = useState(0)
  const [sending, setSending] = useState(false)

  // ponytail: simple countdown — interval runs only while remaining > 0, cleanup on unmount
  useEffect(() => {
    if (remaining <= 0) return
    const id = setInterval(() => setRemaining((r) => Math.max(0, r - 1)), 1000)
    return () => clearInterval(id)
  }, [remaining > 0]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleSend = useCallback(async () => {
    if (remaining > 0 || sending) return
    setSending(true)
    try {
      const { sent } = await orgApi.members.batchInvite()
      toast.success(`已发送 ${sent} 封激活邀请`)
      setRemaining(COOLDOWN)
    } catch {
      toast.error('发送激活邀请失败')
    } finally {
      setSending(false)
    }
  }, [remaining, sending])

  if (pendingCount <= 0 || dismissed) return null

  const disabled = remaining > 0 || sending

  return (
    <div className="flex items-center gap-3 rounded-md border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm text-amber-800">
      <AlertTriangle className="size-4 shrink-0 text-amber-600" />
      <span className="flex-1">
        当前有 <span className="font-medium">{pendingCount}</span> 名成员尚未激活
      </span>
      <Button
        variant="ghost"
       
        className="h-7 text-xs text-amber-700 hover:bg-amber-100 disabled:opacity-50"
        disabled={disabled}
        onClick={handleSend}
      >
        <Send className="size-3.5" />
        {remaining > 0 ? `${remaining}s` : '发送激活邀请'}
      </Button>
      <Button
        variant="ghost"
        size="icon-sm"
        className="size-6 text-amber-600 hover:bg-amber-100"
        onClick={() => setDismissed(true)}
        aria-label="关闭"
      >
        <X className="size-3.5" />
      </Button>
    </div>
  )
}
