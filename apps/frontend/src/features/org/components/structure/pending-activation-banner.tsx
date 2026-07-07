import { AlertTriangle, Send } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface PendingActivationBannerProps {
  pendingCount: number
}

export function PendingActivationBanner({ pendingCount }: PendingActivationBannerProps) {
  if (pendingCount <= 0) return null

  return (
    <div className="flex items-center gap-3 rounded-md border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm text-amber-800">
      <AlertTriangle className="size-4 shrink-0 text-amber-600" />
      <span className="flex-1">
        当前有 <span className="font-medium">{pendingCount}</span> 名成员尚未激活
      </span>
      <Button variant="ghost" size="sm" className="h-7 text-xs text-amber-700 hover:bg-amber-100">
        <Send className="size-3.5" />
        发送激活邀请
      </Button>
    </div>
  )
}
