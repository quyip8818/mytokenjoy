import { useMemo, useState } from 'react'
import type { Department } from '@/api/types'
import { cn } from '@/lib/utils'
import { ChevronRight, Building2, Users, FolderOpen, Folder } from 'lucide-react'
import { TableSkeleton } from '@/components/ui/table-skeleton'

interface OrgTreeSidebarProps {
  departments: Department[]
  selectedDeptId: string | null
  onSelect: (deptId: string | null) => void
  loading: boolean
}

function TreeNode({
  department,
  level,
  selectedId,
  expandedIds,
  onSelect,
  onToggle,
}: {
  department: Department
  level: number
  selectedId: string | null
  expandedIds: Set<string>
  onSelect: (id: string) => void
  onToggle: (id: string) => void
}) {
  const hasChildren = department.children && department.children.length > 0
  const isSelected = selectedId === department.id
  const isExpanded = expandedIds.has(department.id)

  return (
    <div>
      <div
        role="treeitem"
        tabIndex={0}
        aria-selected={isSelected}
        aria-expanded={hasChildren ? isExpanded : undefined}
        className={cn(
          'group flex items-center gap-2 rounded-md px-2 py-1.5 text-sm cursor-pointer',
          isSelected ? 'bg-primary/8 text-primary font-medium' : 'text-foreground hover:bg-muted',
        )}
        style={{ paddingLeft: `${level * 16 + 8}px` }}
        onClick={() => onSelect(department.id)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            onSelect(department.id)
          }
        }}
      >
        {hasChildren ? (
          <span
            role="button"
            tabIndex={-1}
            className="flex size-4 shrink-0 items-center justify-center"
            onClick={(e) => {
              e.stopPropagation()
              onToggle(department.id)
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.stopPropagation()
                onToggle(department.id)
              }
            }}
          >
            <ChevronRight
              className={cn(
                'size-3.5 text-muted-foreground transition-transform duration-150',
                isExpanded && 'rotate-90',
              )}
            />
          </span>
        ) : (
          <span className="size-4" />
        )}

        {hasChildren ? (
          isExpanded ? (
            <FolderOpen className="size-4 shrink-0 text-muted-foreground" />
          ) : (
            <Folder className="size-4 shrink-0 text-muted-foreground" />
          )
        ) : (
          <Users className="size-4 shrink-0 text-muted-foreground" />
        )}

        <span className="flex-1 truncate">{department.name}</span>
      </div>

      {hasChildren && isExpanded && (
        <div>
          {department.children!.map((child) => (
            <TreeNode
              key={child.id}
              department={child}
              level={level + 1}
              selectedId={selectedId}
              expandedIds={expandedIds}
              onSelect={onSelect}
              onToggle={onToggle}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export function OrgTreeSidebar({
  departments,
  selectedDeptId,
  onSelect,
  loading,
}: OrgTreeSidebarProps) {
  const defaultExpanded = useMemo(() => new Set(departments.map((d) => d.id)), [departments])
  const [userExpanded, setUserExpanded] = useState<Set<string> | null>(null)
  const expanded = userExpanded ?? defaultExpanded

  const toggleExpand = (id: string) => {
    setUserExpanded((prev) => {
      const current = prev ?? defaultExpanded
      const next = new Set(current)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  return (
    <div className="flex w-[260px] shrink-0 flex-col border-r border-border bg-card">
      <div className="border-b border-border px-4 py-3">
        <span className="text-xs font-medium text-muted-foreground tracking-wide">
          选择查看范围
        </span>
      </div>

      <div
        role="treeitem"
        tabIndex={0}
        aria-selected={!selectedDeptId}
        className={cn(
          'flex cursor-pointer items-center gap-2 border-b border-border px-4 py-2.5 text-sm',
          !selectedDeptId
            ? 'bg-primary/8 text-primary font-medium'
            : 'text-foreground hover:bg-muted',
        )}
        onClick={() => onSelect(null)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            onSelect(null)
          }
        }}
      >
        <Building2 className="size-4 shrink-0 text-muted-foreground" />
        <span>全公司</span>
      </div>

      <div className="flex-1 overflow-y-auto p-2">
        {loading ? (
          <TableSkeleton rows={6} columns={1} />
        ) : (
          departments.map((dept) => (
            <TreeNode
              key={dept.id}
              department={dept}
              level={0}
              selectedId={selectedDeptId}
              expandedIds={expanded}
              onSelect={onSelect}
              onToggle={toggleExpand}
            />
          ))
        )}
      </div>
    </div>
  )
}
