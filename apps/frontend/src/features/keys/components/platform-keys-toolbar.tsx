import { Plus, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { PermissionGate } from '@/features/session'
import { PERMISSION } from '@/lib/permissions'
import { cn } from '@/lib/utils'
import type { PlatformKeyTab } from '@/features/keys'

const TAB_HINTS: Record<PlatformKeyTab, { label: string; desc: string; example: string }> = {
  member: {
    label: '成员 Key',
    desc: '绑定个人，消耗成员个人额度',
    example: '例如：给张三分配一把日常开发用的 Key',
  },
  project: {
    label: '项目 Key',
    desc: '绑定项目，可分配给某个服务使用，支持设置有效期',
    example: '例如：给「智能客服」项目的后端服务分配一把共用 Key',
  },
  project_member: {
    label: '项目成员 Key',
    desc: '绑定项目中的某个成员，同时受项目总额度与该成员子额度限制',
    example: '例如：张三在「智能客服」项目中独立使用，互不影响',
  },
}

const TAB_ORDER: PlatformKeyTab[] = ['member', 'project', 'project_member']

interface PlatformKeysToolbarProps {
  activeTab: PlatformKeyTab
  onTabChange: (tab: PlatformKeyTab) => void
  search: string
  onSearchChange: (value: string) => void
  onCreateKey: () => void
}

export function PlatformKeysToolbar({
  activeTab,
  onTabChange,
  search,
  onSearchChange,
  onCreateKey,
}: PlatformKeysToolbarProps) {
  return (
    <div className="flex items-center justify-between border-b border-border px-5 py-3">
      <div className="flex items-center gap-1">
        {TAB_ORDER.map((tab) => {
          const hint = TAB_HINTS[tab]
          return (
            <Tooltip key={tab}>
              <TooltipTrigger asChild>
                <button
                  type="button"
                  onClick={() => onTabChange(tab)}
                  className={cn(
                    'rounded-md px-3 py-1.5 text-sm font-medium transition-colors duration-100',
                    activeTab === tab
                      ? 'bg-muted text-foreground'
                      : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground',
                  )}
                >
                  {hint.label}
                </button>
              </TooltipTrigger>
              <TooltipContent className="flex flex-col items-start">
                <p>{hint.desc}</p>
                <p>{hint.example}</p>
              </TooltipContent>
            </Tooltip>
          )
        })}
      </div>

      <div className="flex items-center gap-3">
        <div className="relative">
          <Search className="absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="搜索 Key..."
            className="h-8 w-52 pl-8 text-sm"
          />
        </div>
        <PermissionGate write permission={PERMISSION.KEYS_ADMIN}>
          {activeTab === 'project' ? (
            <Button variant="brand" className="h-8 gap-1.5" onClick={onCreateKey}>
              <Plus className="size-3.5" />
              签发 Key
            </Button>
          ) : (
            <Tooltip>
              <TooltipTrigger asChild>
                <span tabIndex={0}>
                  <Button variant="brand" className="h-8 gap-1.5" disabled>
                    <Plus className="size-3.5" />
                    签发 Key
                  </Button>
                </span>
              </TooltipTrigger>
              <TooltipContent>
                {activeTab === 'member'
                  ? '成员 Key 请让对应成员在「我的 Key」中创建'
                  : '项目成员 Key 请让对应成员在「我的 Key」中创建'}
              </TooltipContent>
            </Tooltip>
          )}
        </PermissionGate>
      </div>
    </div>
  )
}
