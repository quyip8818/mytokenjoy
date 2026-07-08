import { Building2, MessageSquare, Users } from 'lucide-react'
import type { Platform } from '@/api/types'

export const PLATFORM_ICON_META: Record<
  Platform,
  { icon: typeof MessageSquare; iconClassName: string }
> = {
  feishu: { icon: MessageSquare, iconClassName: 'bg-blue-50 text-blue-600' },
  dingtalk: { icon: Building2, iconClassName: 'bg-sky-50 text-sky-600' },
  wecom: { icon: Users, iconClassName: 'bg-emerald-50 text-emerald-600' },
}
