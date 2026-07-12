import { Plus, Search } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import { cn } from '@/lib/utils'
import type { PlatformKeyTab } from '@/features/keys'

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
        <button
          type="button"
          onClick={() => onTabChange('member')}
          className={cn(
            'rounded-md px-3 py-1.5 text-sm font-medium transition-colors duration-100',
            activeTab === 'member'
              ? 'bg-muted text-foreground'
              : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground',
          )}
        >
          成员 Key
        </button>
        <button
          type="button"
          onClick={() => onTabChange('project')}
          className={cn(
            'rounded-md px-3 py-1.5 text-sm font-medium transition-colors duration-100',
            activeTab === 'project'
              ? 'bg-muted text-foreground'
              : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground',
          )}
        >
          项目 Key
        </button>
        <button
          type="button"
          onClick={() => onTabChange('project_member')}
          className={cn(
            'rounded-md px-3 py-1.5 text-sm font-medium transition-colors duration-100',
            activeTab === 'project_member'
              ? 'bg-muted text-foreground'
              : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground',
          )}
        >
          项目成员 Key
        </button>
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
          <Button size="sm" variant="brand" className="h-8 gap-1.5" onClick={onCreateKey}>
            <Plus className="size-3.5" />
            签发 Key
          </Button>
        </PermissionGate>
      </div>
    </div>
  )
}
