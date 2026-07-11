import { useState, useMemo } from 'react'
import type { BudgetNode } from '@/api/types'
import { cn } from '@/lib/utils'
import { ChevronRight, Folder, FolderOpen, Search, Users } from 'lucide-react'
import { Input } from '@/components/ui/input'

interface BudgetTreePanelProps {
  tree: BudgetNode[]
  selectedId?: string
  onSelect: (nodeId: string) => void
}

function TreeNode({
  node,
  level,
  selectedId,
  onSelect,
  expandedIds,
  onToggle,
  searchKeyword,
}: {
  node: BudgetNode
  level: number
  selectedId: string | undefined
  onSelect: (nodeId: string) => void
  expandedIds: Set<string>
  onToggle: (id: string) => void
  searchKeyword: string
}) {
  const hasChildren = node.children && node.children.length > 0
  const isExpanded = expandedIds.has(node.id)
  const isSelected = selectedId === node.id

  const highlightText = (text: string) => {
    if (!searchKeyword) return text
    const lower = text.toLowerCase()
    const idx = lower.indexOf(searchKeyword.toLowerCase())
    if (idx === -1) return text
    return (
      <>
        {text.slice(0, idx)}
        <mark className="rounded-sm bg-amber-100 px-0.5 text-inherit">
          {text.slice(idx, idx + searchKeyword.length)}
        </mark>
        {text.slice(idx + searchKeyword.length)}
      </>
    )
  }

  return (
    <div>
      <div
        role="treeitem"
        tabIndex={0}
        aria-selected={isSelected}
        aria-expanded={hasChildren ? isExpanded : undefined}
        className={cn(
          'group relative flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm',
          isSelected ? 'bg-primary/8 text-primary' : 'text-foreground hover:bg-muted',
        )}
        style={{ paddingLeft: `${level * 16 + 8}px` }}
        onClick={() => onSelect(node.id)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            onSelect(node.id)
          } else if (e.key === 'ArrowRight') {
            e.preventDefault()
            if (hasChildren && !isExpanded) {
              onToggle(node.id)
            }
          } else if (e.key === 'ArrowLeft') {
            e.preventDefault()
            if (hasChildren && isExpanded) {
              onToggle(node.id)
            }
          } else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
            e.preventDefault()
            const items = (
              e.currentTarget.closest('[role="tree"]') as HTMLElement
            )?.querySelectorAll<HTMLElement>('[role="treeitem"]')
            if (!items) return
            const idx = Array.from(items).indexOf(e.currentTarget as HTMLElement)
            const next = e.key === 'ArrowDown' ? idx + 1 : idx - 1
            if (next >= 0 && next < items.length) {
              items[next].focus()
            }
          }
        }}
      >
        {/* Expand toggle */}
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

        {/* Icon */}
        {hasChildren ? (
          isExpanded ? (
            <FolderOpen className="size-4 shrink-0 text-muted-foreground" />
          ) : (
            <Folder className="size-4 shrink-0 text-muted-foreground" />
          )
        ) : (
          <Users className="size-4 shrink-0 text-muted-foreground" />
        )}

        {/* Name */}
        <span className="flex-1 truncate font-medium">{highlightText(node.name)}</span>
      </div>

      {/* Children */}
      {hasChildren && isExpanded && (
        <div>
          {node.children!.map((child) => (
            <TreeNode
              key={child.id}
              node={child}
              level={level + 1}
              selectedId={selectedId}
              expandedIds={expandedIds}
              onSelect={onSelect}
              onToggle={onToggle}
              searchKeyword={searchKeyword}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export function BudgetTreePanel({ tree, selectedId, onSelect }: BudgetTreePanelProps) {
  const [search, setSearch] = useState('')
  const [userExpanded, setUserExpanded] = useState<Set<string> | null>(null)

  // Default: only expand root-level nodes (consistent with org tree)
  const defaultExpanded = useMemo(
    () => (tree.length > 0 ? new Set(tree.map((node) => node.id)) : new Set<string>()),
    [tree],
  )
  const expanded = userExpanded ?? defaultExpanded

  const toggleExpand = (id: string) => {
    setUserExpanded((prev) => {
      let current: Set<string>
      if (prev !== null) {
        current = prev
      } else if (search) {
        // During search, seed from all-expanded so collapsing one doesn't collapse everything
        const allIds = new Set<string>()
        function collectAll(nodes: BudgetNode[]) {
          for (const n of nodes) {
            allIds.add(n.id)
            if (n.children) collectAll(n.children)
          }
        }
        collectAll(filteredTree)
        current = allIds
      } else {
        current = defaultExpanded
      }
      const next = new Set(current)
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
    const lower = search.toLowerCase()

    function filterNodes(nodes: BudgetNode[]): BudgetNode[] {
      return nodes.reduce<BudgetNode[]>((acc, n) => {
        const children = n.children ? filterNodes(n.children) : []
        if (n.name.toLowerCase().includes(lower) || children.length > 0) {
          acc.push({ ...n, children: children.length > 0 ? children : n.children })
        }
        return acc
      }, [])
    }

    return filterNodes(tree)
  }, [tree, search])

  // When searching, auto-expand all filtered nodes so matches are visible
  // but respect user's manual toggles if they've interacted
  const effectiveExpanded = useMemo(() => {
    if (!search) return expanded
    // Collect all IDs in the filtered tree
    const allIds = new Set<string>()
    function collectAll(nodes: BudgetNode[]) {
      for (const n of nodes) {
        allIds.add(n.id)
        if (n.children) collectAll(n.children)
      }
    }
    collectAll(filteredTree)
    // If user has manually toggled, respect their state
    if (userExpanded !== null) return userExpanded
    return allIds
  }, [search, filteredTree, expanded, userExpanded])

  return (
    <div className="flex w-64 shrink-0 flex-col border-r border-border">
      <div className="flex items-center gap-2 border-b border-border p-3">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="text"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value)
              // Reset manual toggles when search keyword changes so auto-expand takes over
              setUserExpanded(null)
            }}
            placeholder="搜索部门..."
            className="h-8 pl-8 text-sm"
          />
        </div>
      </div>
      <div className="flex-1 overflow-y-auto p-2" role="tree">
        {filteredTree.map((node) => (
          <TreeNode
            key={node.id}
            node={node}
            level={0}
            selectedId={selectedId}
            expandedIds={effectiveExpanded}
            onSelect={onSelect}
            onToggle={toggleExpand}
            searchKeyword={search}
          />
        ))}
      </div>
    </div>
  )
}
