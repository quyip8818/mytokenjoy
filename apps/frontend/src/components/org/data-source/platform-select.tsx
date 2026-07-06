import { useState } from 'react'
import type { Platform } from '@/api/types'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { MessageSquare, Building2, Users } from 'lucide-react'

const platforms: {
  id: Platform
  name: string
  description: string
  icon: typeof MessageSquare
}[] = [
  { id: 'feishu', name: '飞书', description: '通过飞书开放平台同步组织架构和成员信息', icon: MessageSquare },
  { id: 'dingtalk', name: '钉钉', description: '通过钉钉开放平台同步组织架构和成员信息', icon: Building2 },
  { id: 'wecom', name: '企业微信', description: '通过企业微信 API 同步组织架构和成员信息', icon: Users },
]

interface PlatformSelectProps {
  onSelect: (platform: Platform) => void
}

export function PlatformSelect({ onSelect }: PlatformSelectProps) {
  const [selected, setSelected] = useState<Platform | null>(null)

  return (
    <div className="space-y-6">
      <div className="text-center space-y-1.5">
        <h2 className="text-sm font-semibold text-foreground">选择数据源平台</h2>
        <p className="text-sm text-muted-foreground">
          选择需要对接的第三方协作平台，导入组织架构与成员数据
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {platforms.map((p) => {
          const Icon = p.icon
          const isSelected = selected === p.id
          return (
            <button
              key={p.id}
              type="button"
              onClick={() => setSelected(p.id)}
              className={cn(
                'group relative flex flex-col items-center gap-3 rounded-lg border p-5 text-center transition-all duration-150 outline-none',
                'hover:shadow-sm',
                'focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:border-primary',
                isSelected
                  ? 'border-primary shadow-sm'
                  : 'border-border bg-card'
              )}
            >
              <div className={cn(
                'flex size-10 items-center justify-center rounded-md transition-colors duration-150',
                isSelected ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'
              )}>
                <Icon className="size-5" />
              </div>
              <div>
                <p className="text-sm font-medium text-foreground">{p.name}</p>
                <p className="mt-1 text-xs text-muted-foreground leading-relaxed">{p.description}</p>
              </div>
            </button>
          )
        })}
      </div>

      <div className="flex justify-center">
        <Button disabled={!selected} onClick={() => selected && onSelect(selected)} className="px-6">
          开始配置
        </Button>
      </div>
    </div>
  )
}
