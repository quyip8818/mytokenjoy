import { useState } from 'react'
import type { Role } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { PencilIcon, TrashIcon, Search, Plus, Shield, ShieldCheck } from 'lucide-react'

interface RoleListProps {
  roles: Role[]
  selectedRoleId: string | null
  onSelect: (role: Role) => void
  onAdd: () => void
  onEdit: (role: Role) => void
  onDelete: (role: Role) => void
}

export function RoleList({
  roles,
  selectedRoleId,
  onSelect,
  onAdd,
  onEdit,
  onDelete,
}: RoleListProps) {
  const [search, setSearch] = useState('')

  const filtered = roles.filter((r) =>
    r.name.toLowerCase().includes(search.toLowerCase()),
  )

  const presetRoles = filtered.filter((r) => r.type === 'preset')
  const customRoles = filtered.filter((r) => r.type === 'custom')

  return (
    <div className="w-[240px] shrink-0 border-r border-border bg-card flex flex-col">
      {/* Header */}
      <div className="p-4 space-y-3">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-foreground">角色</h3>
          <Button size="sm" className="h-7 text-xs gap-1" onClick={onAdd}>
            <Plus className="h-3.5 w-3.5" strokeWidth={1.5} />
            新建
          </Button>
        </div>
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" strokeWidth={1.5} />
          <Input
            type="text"
            placeholder="搜索角色..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-8 h-8 text-sm"
          />
        </div>
      </div>

      {/* Role groups */}
      <div className="flex-1 overflow-y-auto px-2 pb-3">
        {presetRoles.length > 0 && (
          <RoleGroup
            title="系统预设"
            roles={presetRoles}
            selectedRoleId={selectedRoleId}
            onSelect={onSelect}
            onEdit={onEdit}
            onDelete={onDelete}
          />
        )}
        {customRoles.length > 0 && (
          <RoleGroup
            title="自定义"
            roles={customRoles}
            selectedRoleId={selectedRoleId}
            onSelect={onSelect}
            onEdit={onEdit}
            onDelete={onDelete}
          />
        )}
        {filtered.length === 0 && (
          <p className="text-xs text-muted-foreground text-center mt-8">无匹配角色</p>
        )}
      </div>
    </div>
  )
}

function RoleGroup({
  title,
  roles,
  selectedRoleId,
  onSelect,
  onEdit,
  onDelete,
}: {
  title: string
  roles: Role[]
  selectedRoleId: string | null
  onSelect: (role: Role) => void
  onEdit: (role: Role) => void
  onDelete: (role: Role) => void
}) {
  return (
    <div className="mb-4">
      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide px-2 mb-1.5">
        {title}
      </p>
      <div className="space-y-0.5">
        {roles.map((role) => (
          <RoleItem
            key={role.id}
            role={role}
            selected={role.id === selectedRoleId}
            onSelect={() => onSelect(role)}
            onEdit={() => onEdit(role)}
            onDelete={() => onDelete(role)}
          />
        ))}
      </div>
    </div>
  )
}

function RoleItem({
  role,
  selected,
  onSelect,
  onEdit,
  onDelete,
}: {
  role: Role
  selected: boolean
  onSelect: () => void
  onEdit: () => void
  onDelete: () => void
}) {
  const Icon = role.type === 'preset' ? ShieldCheck : Shield

  return (
    <div
      className={cn(
        'group flex items-center gap-2.5 px-2.5 py-2 rounded-md cursor-pointer transition-colors duration-100',
        selected
          ? 'bg-muted text-foreground'
          : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground',
      )}
      onClick={onSelect}
    >
      <Icon className="h-4 w-4 shrink-0" strokeWidth={1.5} />
      <span className={cn('text-sm truncate flex-1', selected && 'font-medium')}>
        {role.name}
      </span>
      <div className="flex items-center gap-1">
        <span className="text-xs text-muted-foreground tabular-nums">{role.memberCount}</span>
        {role.type === 'custom' && (
          <div className="hidden group-hover:flex gap-0.5 ml-1">
            <button
              className="p-0.5 rounded hover:bg-background text-muted-foreground hover:text-foreground transition-colors"
              onClick={(e) => { e.stopPropagation(); onEdit() }}
            >
              <PencilIcon className="h-3 w-3" strokeWidth={1.5} />
            </button>
            <button
              className="p-0.5 rounded hover:bg-background text-muted-foreground hover:text-destructive transition-colors"
              onClick={(e) => { e.stopPropagation(); onDelete() }}
            >
              <TrashIcon className="h-3 w-3" strokeWidth={1.5} />
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
