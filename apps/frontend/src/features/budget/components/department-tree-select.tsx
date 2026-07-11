import { useState, useMemo } from 'react'
import type { BudgetNode } from '@/api/types'
import { cn } from '@/lib/utils'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Input } from '@/components/ui/input'
import { ChevronDown, ChevronRight, Folder, FolderOpen, Search, Users } from 'lucide-react'

interface DepartmentTreeSelectProps {
  tree: BudgetNode[]
  value: string
  onChange: (id: string, name: string) => void
  placeholder?: string
}

function findNodeName(nodes: BudgetNode[], id: string): string | undefined {
  for (const n of nodes) {
    if (n.id === id) return n.name
    if (n.children) {
      const found = findNodeName(n.children, id)
      if (found) return found
    }
  }
  return undefined
}

function TreeItem({
  node,
  depth,
  selectedId,
  expanded,
  onToggle,
  onSelect,
}: {
  node: BudgetNode
  depth: number
  selectedId: string
  expanded: Set<string>
  onToggle: (id: string) => void
  onSelect: (node: BudgetNode) => void
}) {
  const hasChildren = node.children && node.children.length > 0
  const isExpanded = expanded.has(node.id)
  const isSelected = node.id === selectedId

  return (
    <>
      <div
        role="option"
        tabIndex={0}
        aria-selected={isSelected}
        className={cn(
          'flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm',
          isSelected ? 'bg-primary/8 text-primary' : 'text-foreground hover:bg-muted',
        )}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={() => onSelect(node)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            onSelect(node)
          } else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
            e.preventDefault()
            const items = (e.currentTarget.closest('[role="listbox"]') as HTMLElement)?.querySelectorAll<HTMLElement>('[role="option"]')
            if (!items) return
            const idx = Array.from(items).indexOf(e.currentTarget as HTMLElement)
            const next = e.key === 'ArrowDown' ? idx + 1 : idx - 1
            if (next >= 0 && next < items.length) {
              items[next].focus()
            }
          }
        }}
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
          <TreeItem
            key={child.id}
            node={child}
            depth={depth + 1}
            selectedId={selectedId}
            expanded={expanded}
            onToggle={onToggle}
            onSelect={onSelect}
          />
        ))}
    </>
  )
}

export function DepartmentTreeSelect({
  tree,
  value,
  onChange,
  placeholder = '选择团队…',
}: DepartmentTreeSelectProps) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [expanded, setExpanded] = useState<Set<string>>(() => {
    const ids = new Set<string>()
    function collect(nodes: BudgetNode[]) {
      for (const n of nodes) {
        if (n.children?.length) {
          ids.add(n.id)
          collect(n.children)
        }
      }
    }
    collect(tree)
    return ids
  })

  const toggleExpand = (id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const filteredTree = useMemo(() => {
    if (!search) return tree
    function filter(nodes: BudgetNode[]): BudgetNode[] {
      return nodes.reduce<BudgetNode[]>((acc, n) => {
        const children = n.children ? filter(n.children) : []
        if (n.name.toLowerCase().includes(search.toLowerCase()) || children.length > 0) {
          acc.push({ ...n, children: children.length > 0 ? children : n.children })
        }
        return acc
      }, [])
    }
    return filter(tree)
  }, [tree, search])

  const selectedName = value ? findNodeName(tree, value) : undefined

  const handleSelect = (node: BudgetNode) => {
    onChange(node.id, node.name)
    setOpen(false)
    setSearch('')
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className={cn(
            'flex h-8 w-full items-center justify-between rounded-md border border-border bg-transparent px-3 text-sm',
            'hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
            !selectedName && 'text-muted-foreground',
          )}
        >
          <span className="truncate">{selectedName ?? placeholder}</span>
          <ChevronDown className="size-3.5 shrink-0 text-muted-foreground" />
        </button>
      </PopoverTrigger>
      <PopoverContent
        className="w-[--radix-popover-trigger-width] p-0"
        align="start"
        sideOffset={4}
      >
        {/* Search */}
        <div className="border-b border-border p-2">
          <div className="relative">
            <Search className="absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="搜索部门..."
              className="h-7 pl-7 text-sm"
            />
          </div>
        </div>
        {/* Tree */}
        <div className="max-h-64 overflow-y-auto p-1" role="listbox">
          {filteredTree.length === 0 ? (
            <p className="py-4 text-center text-xs text-muted-foreground">无匹配结果</p>
          ) : (
            filteredTree.map((node) => (
              <TreeItem
                key={node.id}
                node={node}
                depth={0}
                selectedId={value}
                expanded={expanded}
                onToggle={toggleExpand}
                onSelect={handleSelect}
              />
            ))
          )}
        </div>
      </PopoverContent>
    </Popover>
  )
}
