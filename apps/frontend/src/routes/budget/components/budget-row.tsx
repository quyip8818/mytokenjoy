import { useState } from 'react'
import type { BudgetNode } from '@/api/types'
import { BudgetProgressCell } from '@/components/ui/budget-progress-cell'
import { Button } from '@/components/ui/button'
import { TableCell, TableRow } from '@/components/ui/table'
import { computeUnallocated, findBudgetNode } from '@/lib/budget'
import { cn } from '@/lib/utils'

export interface BudgetRowProps {
  node: BudgetNode
  depth: number
  tree: BudgetNode[]
  onAllocate: (node: BudgetNode, parent: BudgetNode | null) => void
  allocateHighlight?: string
  allocateCtaId?: string
  canAllocate?: boolean
}

export function BudgetRow({
  node,
  depth,
  tree,
  onAllocate,
  allocateHighlight,
  allocateCtaId,
  canAllocate = true,
}: BudgetRowProps) {
  const [expanded, setExpanded] = useState(true)
  const hasChildren = node.children && node.children.length > 0
  const parent = node.parentId ? findBudgetNode(tree, node.parentId) : null
  const unallocated = computeUnallocated(node)

  return (
    <>
      <TableRow>
        <TableCell>
          <div className="flex items-center" style={{ paddingLeft: `${depth * 20}px` }}>
            {hasChildren && (
              <button
                type="button"
                onClick={() => setExpanded(!expanded)}
                className="mr-2 w-4 text-xs text-blue-400 hover:text-blue-300"
              >
                {expanded ? '▾' : '▸'}
              </button>
            )}
            {!hasChildren && <span className="mr-2 w-4" />}
            <span className="font-medium">{node.name}</span>
          </div>
        </TableCell>
        <TableCell className="text-right">¥{node.budget.toLocaleString()}</TableCell>
        <TableCell className="text-right">¥{node.consumed.toLocaleString()}</TableCell>
        <TableCell className="text-right">¥{(node.reservedPool ?? 0).toLocaleString()}</TableCell>
        <TableCell className="text-right">¥{unallocated.toLocaleString()}</TableCell>
        <TableCell className="w-40">
          <BudgetProgressCell value={node.consumed} total={node.budget} />
        </TableCell>
        <TableCell className="w-[120px]">
          {canAllocate ? (
            <Button
              id={depth === 0 ? allocateCtaId : undefined}
              variant="ghost"
              size="sm"
              className={cn(depth === 0 ? allocateHighlight : undefined)}
              onClick={() => onAllocate(node, parent)}
            >
              分配
            </Button>
          ) : null}
        </TableCell>
      </TableRow>
      {expanded &&
        node.children?.map((child) => (
          <BudgetRow
            key={child.id}
            node={child}
            depth={depth + 1}
            tree={tree}
            onAllocate={onAllocate}
            canAllocate={canAllocate}
          />
        ))}
    </>
  )
}
