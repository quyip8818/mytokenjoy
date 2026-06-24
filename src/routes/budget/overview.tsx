import { Link } from 'react-router'
import { useState } from 'react'
import { PieChart } from 'lucide-react'
import { StatCard } from '@/components/ui/stat-card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { BudgetProgressCell } from '@/components/ui/budget-progress-cell'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatusBadge } from '@/components/ui/status-badge'
import { budgetApi } from '@/api/budget'
import type { BudgetNode } from '@/api/types'
import { useAsyncResource } from '@/hooks/use-async-resource'
import { useWorkflowRefresh } from '@/hooks/use-workflow-refresh'
import { useDemoCta } from '@/features/demo'
import { computeUnallocated, findBudgetNode } from '@/lib/budget'
import { listEmpty } from '@/lib/list-empty'
import { cn } from '@/lib/utils'

function BudgetRow({
  node,
  depth,
  tree,
  onAllocate,
  allocateHighlight,
  allocateCtaId,
}: {
  node: BudgetNode
  depth: number
  tree: BudgetNode[]
  onAllocate: (node: BudgetNode, parent: BudgetNode | null) => void
  allocateHighlight?: string
  allocateCtaId?: string
}) {
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
                className="mr-2 w-4 text-xs text-indigo-400 hover:text-indigo-300"
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
          <Button
            id={depth === 0 ? allocateCtaId : undefined}
            variant="ghost"
            size="sm"
            className={cn(depth === 0 ? allocateHighlight : undefined)}
            onClick={() => onAllocate(node, parent)}
          >
            分配
          </Button>
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
          />
        ))}
    </>
  )
}

export default function BudgetOverviewPage() {
  const budgetCta = useDemoCta('BUDGET')
  const { data: tree = [], loading, refresh } = useAsyncResource(() => budgetApi.getTree(), [])
  const { openWithRefresh } = useWorkflowRefresh(refresh)

  const root = tree[0]
  const summary = root
    ? {
        budget: root.budget,
        consumed: root.consumed,
        unallocated: computeUnallocated(root),
      }
    : { budget: 0, consumed: 0, unallocated: 0 }

  const handleAllocate = (node: BudgetNode, parent: BudgetNode | null) => {
    openWithRefresh('budget-node-edit', { node, parent })
  }

  return (
    <PageShell
      leading={
        <div className="grid max-w-2xl grid-cols-3 gap-4">
          <StatCard label="总预算" value={loading ? '-' : `¥${summary.budget.toLocaleString()}`} />
          <StatCard label="已用" value={loading ? '-' : `¥${summary.consumed.toLocaleString()}`} />
          <StatCard
            label="未分配"
            value={loading ? '-' : `¥${summary.unallocated.toLocaleString()}`}
            accent
          />
        </div>
      }
      actions={<StatusBadge variant="info">周期：2026 年 6 月</StatusBadge>}
    >
      <DataSection
        loading={loading}
        skeletonColumns={7}
        empty={listEmpty(loading, tree, {
          icon: PieChart,
          title: '暂无预算数据',
          description: '请先导入组织后再分配预算',
        })}
      >
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead>节点</TableHead>
              <TableHead className="text-right">预算</TableHead>
              <TableHead className="text-right">已消耗</TableHead>
              <TableHead className="text-right">预留池</TableHead>
              <TableHead className="text-right">未分配</TableHead>
              <TableHead className="w-40">进度</TableHead>
              <TableHead className="w-[120px]">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {tree.map((node) => (
              <BudgetRow
                key={node.id}
                node={node}
                depth={0}
                tree={tree}
                onAllocate={handleAllocate}
                allocateHighlight={budgetCta.className}
                allocateCtaId={budgetCta.id}
              />
            ))}
          </TableBody>
        </Table>
        <p className="mt-4 text-xs text-muted-foreground">
          超限行为由全局{' '}
          <Link to="/budget/alerts" className="text-indigo-600 hover:underline">
            超限策略
          </Link>{' '}
          统一配置。预算周期为自然月，月初已用额度清零由后端处理，Demo 不模拟月重置。
        </p>
      </DataSection>
    </PageShell>
  )
}
