import { useState } from 'react'
import type { Platform } from '@/api/types'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { CheckIcon, MessageSquare, Building2, Users } from 'lucide-react'

const platforms: {
  id: Platform
  name: string
  description: string
  icon: typeof MessageSquare
  iconClassName: string
  available: boolean
}[] = [
  {
    id: 'feishu',
    name: '飞书',
    description: '通过飞书开放平台同步组织架构和成员信息',
    icon: MessageSquare,
    iconClassName: 'bg-blue-50 text-blue-600',
    available: true,
  },
  {
    id: 'dingtalk',
    name: '钉钉',
    description: '通过钉钉开放平台同步组织架构和成员信息',
    icon: Building2,
    iconClassName: 'bg-sky-50 text-sky-600',
    available: true,
  },
  {
    id: 'wecom',
    name: '企业微信',
    description: '通过企业微信 API 同步组织架构和成员信息',
    icon: Users,
    iconClassName: 'bg-emerald-50 text-emerald-600',
    available: true,
  },
]

interface PlatformSelectProps {
  onSelect: (platform: Platform) => void
}

export function PlatformSelect({ onSelect }: PlatformSelectProps) {
  const [selected, setSelected] = useState<Platform | null>(null)

  return (
    <div className="space-y-8">
      <div className="text-center space-y-1.5">
        <h2 className="text-base font-semibold text-foreground">选择数据源平台</h2>
        <p className="text-sm text-muted-foreground">
          选择需要对接的第三方协作平台，导入组织架构与成员数据
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4" role="radiogroup" aria-label="数据源平台">
        {platforms.map((p) => {
          const Icon = p.icon
          const isSelected = selected === p.id
          return (
            <button
              key={p.id}
              type="button"
              role="radio"
              aria-checked={isSelected}
              disabled={!p.available}
              onClick={() => setSelected(p.id)}
              className={cn(
                'group relative flex flex-col items-center gap-3 rounded-xl border-2 bg-card p-6 text-center transition-all duration-150 outline-none',
                'focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:border-primary',
                p.available
                  ? 'cursor-pointer hover:border-primary/40 hover:shadow-sm'
                  : 'cursor-not-allowed opacity-60',
                isSelected ? 'border-primary shadow-sm' : 'border-border',
              )}
            >
              {isSelected && (
                <span className="absolute right-3 top-3 flex size-5 items-center justify-center rounded-full bg-primary text-primary-foreground">
                  <CheckIcon className="size-3" />
                </span>
              )}
              {!p.available && (
                <Badge
                  variant="secondary"
                  className="absolute right-3 top-3 text-[10px] text-muted-foreground"
                >
                  即将支持
                </Badge>
              )}
              <div
                className={cn(
                  'flex size-12 items-center justify-center rounded-lg transition-colors duration-150',
                  p.available ? p.iconClassName : 'bg-muted text-muted-foreground',
                )}
              >
                <Icon className="size-6" />
              </div>
              <div>
                <p className="text-sm font-semibold text-foreground">{p.name}</p>
                <p className="mt-1 text-xs text-muted-foreground leading-relaxed">
                  {p.description}
                </p>
              </div>
            </button>
          )
        })}
      </div>

      <div className="flex flex-col items-center gap-2">
        <Button
          size="lg"
          disabled={!selected}
          onClick={() => selected && onSelect(selected)}
          className="px-8"
        >
          开始配置
        </Button>
        {!selected && <p className="text-xs text-muted-foreground">请先选择一个平台</p>}
      </div>
    </div>
  )
}
