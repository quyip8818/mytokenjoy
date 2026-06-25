import { useState } from 'react'
import type { Role } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { PencilIcon, TrashIcon } from 'lucide-react'

interface RoleListProps {
  roles: Role[]
  selectedRoleId: string | null
  onSelect: (role: Role) => void
  onAdd: () => void
  onEdit: (role: Role) => void
  onDelete: (role: Role) => void
  readOnly?: boolean
}

export function RoleList({
  roles,
  selectedRoleId,
  onSelect,
  onAdd,
  onEdit,
  onDelete,
  readOnly = false,
}: RoleListProps) {
  const [search, setSearch] = useState('')

  const filtered = roles.filter((r) => r.name.toLowerCase().includes(search.toLowerCase()))

  const presetRoles = filtered.filter((r) => r.type === 'preset')
  const customRoles = filtered.filter((r) => r.type === 'custom')

  return (
    <div className="flex h-full w-[220px] shrink-0 flex-col rounded-lg border border-border/50 bg-card shadow-card">
      {/* Header */}
      <div className="p-4 border-b border-border">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-semibold text-foreground">角色</h3>
          {!readOnly && (
            <Button size="xs" onClick={onAdd}>
              添加角色
            </Button>
          )}
        </div>
        <Input
          type="text"
          placeholder="搜索角色..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      {/* Role groups */}
      <div className="flex-1 overflow-y-auto p-2">
        {presetRoles.length > 0 && (
          <RoleGroup
            title="系统预设"
            roles={presetRoles}
            selectedRoleId={selectedRoleId}
            onSelect={onSelect}
            onEdit={onEdit}
            onDelete={onDelete}
            readOnly={readOnly}
          />
        )}
        {customRoles.length > 0 && (
          <>
            {presetRoles.length > 0 && <Separator className="my-2" />}
            <RoleGroup
              title="自定义"
              roles={customRoles}
              selectedRoleId={selectedRoleId}
              onSelect={onSelect}
              onEdit={onEdit}
              onDelete={onDelete}
              readOnly={readOnly}
            />
          </>
        )}
        {filtered.length === 0 && (
          <p className="text-xs text-muted-foreground text-center mt-4">无匹配角色</p>
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
  readOnly = false,
}: {
  title: string
  roles: Role[]
  selectedRoleId: string | null
  onSelect: (role: Role) => void
  onEdit: (role: Role) => void
  onDelete: (role: Role) => void
  readOnly?: boolean
}) {
  return (
    <div className="mb-3">
      <p className="text-xs font-medium text-muted-foreground uppercase px-2 mb-1">{title}</p>
      {roles.map((role) => (
        <RoleItem
          key={role.id}
          role={role}
          selected={role.id === selectedRoleId}
          onSelect={() => onSelect(role)}
          onEdit={() => onEdit(role)}
          onDelete={() => onDelete(role)}
          readOnly={readOnly}
        />
      ))}
    </div>
  )
}

function RoleItem({
  role,
  selected,
  onSelect,
  onEdit,
  onDelete,
  readOnly = false,
}: {
  role: Role
  selected: boolean
  onSelect: () => void
  onEdit: () => void
  onDelete: () => void
  readOnly?: boolean
}) {
  return (
    <div
      className={`group flex items-center justify-between px-2 py-1.5 rounded-md cursor-pointer text-sm ${
        selected ? 'bg-accent text-accent-foreground' : 'text-foreground hover:bg-muted'
      }`}
      onClick={onSelect}
    >
      <span className="truncate">{role.name}</span>
      <div className="flex items-center gap-1">
        <Badge variant="secondary" className="text-[10px] px-1.5">
          {role.memberCount}
        </Badge>
        {role.type === 'custom' && !readOnly && (
          <div className="hidden group-hover:flex gap-0.5">
            <Button
              variant="ghost"
              size="icon-xs"
              onClick={(e) => {
                e.stopPropagation()
                onEdit()
              }}
            >
              <PencilIcon />
            </Button>
            <Button
              variant="ghost"
              size="icon-xs"
              onClick={(e) => {
                e.stopPropagation()
                onDelete()
              }}
            >
              <TrashIcon />
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
