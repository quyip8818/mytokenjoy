import { useState, useMemo } from 'react'
import type { Department } from '@/api/types'
import { cn } from '@/lib/utils'
import { ChevronRight, Building2, Users, Search } from 'lucide-react'
import { Input } from '@/components/ui/input'

interface RoutingTreePanelProps {
  departments: Department[]
  selectedId?: string
  onSelect: (nodeId: string) => void
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
  selectedId: string | undefined
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
          'group flex cursor-pointer items-center gap-1.5 rounded-md px-2 py-1.5 text-sm transition-colors',
          isSelected
            ? 'bg-primary/8 font-medium text-primary'
            : 'text-foreground hover:bg-muted/70',
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
          <button
            type="button"
            className="flex size-4 shrink-0 items-center justify-center rounded transition-colors hover:bg-muted"
            onClick={(e) => {
              e.stopPropagation()
              onToggle(department.id)
            }}
            aria-label={isExpanded ? '收起' : '展开'}
          >
            <ChevronRight
              className={cn(
                'size-3 text-muted-foreground transition-transform',
                isExpanded && 'rotate-90',
              )}
            />
          </button>
        ) : (
          <span className="size-4 shrink-0" />
        )}
        {hasChildren ? (
          <Building2 className="size-3.5 shrink-0 text-muted-foreground" />
        ) : (
          <Users className="size-3.5 shrink-0 text-muted-foreground" />
        )}
        <span className="truncate">{department.name}</span>
      </div>
      {hasChildren && isExpanded && (
        <div role="group">
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

export function RoutingTreePanel({ departments, selectedId, onSelect }: RoutingTreePanelProps) {
  const [expandedIds, setExpandedIds] = useState<Set<string>>(() => {
    // Expand root nodes by default
    return new Set(departments.map((d) => d.id))
  })
  const [search, setSearch] = useState('')

  const filteredDepartments = useMemo(() => {
    if (!search.trim()) return departments
    const keyword = search.trim().toLowerCase()
    function filterTree(nodes: Department[]): Department[] {
      return nodes
        .map((node) => {
          const childMatches = node.children ? filterTree(node.children) : []
          if (node.name.toLowerCase().includes(keyword) || childMatches.length > 0) {
            return { ...node, children: childMatches.length > 0 ? childMatches : node.children }
          }
          return null
        })
        .filter(Boolean) as Department[]
    }
    return filterTree(departments)
  }, [departments, search])

  const handleToggle = (id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  return (
    <div className="flex w-56 shrink-0 flex-col overflow-hidden border-r-0">
      <div className="border-b border-border p-3">
        <div className="relative">
          <Search className="absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="搜索团队..."
            className="h-8 pl-8 text-sm"
          />
        </div>
      </div>
      <div className="flex-1 overflow-y-auto p-2" role="tree">
        {filteredDepartments.map((dept) => (
          <TreeNode
            key={dept.id}
            department={dept}
            level={0}
            selectedId={selectedId}
            expandedIds={expandedIds}
            onSelect={onSelect}
            onToggle={handleToggle}
          />
        ))}
      </div>
    </div>
  )
}
