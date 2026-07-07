import type { Department } from '@/api/types'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { ChevronRight, Folder, FolderOpen, Search, Users } from 'lucide-react'

interface PlatformKeysDeptTreeProps {
  departments: Department[]
  selectedId: string | undefined
  onSelect: (id: string) => void
  expanded: Set<string>
  onToggle: (id: string) => void
  treeSearch: string
  onTreeSearchChange: (value: string) => void
}

function TreeNode({
  node,
  depth,
  selectedId,
  onSelect,
  expanded,
  onToggle,
}: {
  node: Department
  depth: number
  selectedId: string | undefined
  onSelect: (id: string) => void
  expanded: Set<string>
  onToggle: (id: string) => void
}) {
  const hasChildren = node.children && node.children.length > 0
  const isExpanded = expanded.has(node.id)
  const isSelected = selectedId === node.id

  return (
    <>
      <div
        role="treeitem"
        tabIndex={0}
        aria-selected={isSelected}
        aria-expanded={hasChildren ? isExpanded : undefined}
        onClick={() => onSelect(node.id)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            onSelect(node.id)
          }
        }}
        className={cn(
          'group flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm',
          isSelected ? 'bg-primary/8 text-primary' : 'text-foreground hover:bg-muted',
        )}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
      >
        {hasChildren ? (
          <span
            role="button"
            tabIndex={-1}
            aria-label={isExpanded ? '收起' : '展开'}
            onClick={(e) => {
              e.stopPropagation()
              onToggle(node.id)
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.stopPropagation()
                onToggle(node.id)
              }
            }}
            className="flex size-4 shrink-0 items-center justify-center"
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
        <span className="flex-1 truncate font-medium">{node.name}</span>
      </div>

      {isExpanded &&
        hasChildren &&
        node.children!.map((child) => (
          <TreeNode
            key={child.id}
            node={child}
            depth={depth + 1}
            selectedId={selectedId}
            onSelect={onSelect}
            expanded={expanded}
            onToggle={onToggle}
          />
        ))}
    </>
  )
}

export function PlatformKeysDeptTree({
  departments,
  selectedId,
  onSelect,
  expanded,
  onToggle,
  treeSearch,
  onTreeSearchChange,
}: PlatformKeysDeptTreeProps) {
  return (
    <div className="flex w-64 shrink-0 flex-col border-r border-border">
      <div className="border-b border-border p-3">
        <div className="relative">
          <Search className="absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={treeSearch}
            onChange={(e) => onTreeSearchChange(e.target.value)}
            placeholder="搜索部门..."
            className="h-8 pl-8 text-sm"
          />
        </div>
      </div>
      <div className="flex-1 overflow-y-auto p-2" role="tree">
        {departments.map((node) => (
          <TreeNode
            key={node.id}
            node={node}
            depth={0}
            selectedId={selectedId}
            onSelect={onSelect}
            expanded={expanded}
            onToggle={onToggle}
          />
        ))}
      </div>
    </div>
  )
}
