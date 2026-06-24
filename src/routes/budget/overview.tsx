import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router'
import { PieChart } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Button } from '@/components/ui/button'
import { budgetApi } from '@/api/budget'
import type { BudgetNode } from '@/api/types'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { useDemoCta } from '@/features/demo'
import { EmptyState } from '@/components/ui/empty-state'
import { computeUnallocated, findBudgetNode } from '@/lib/budget'
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
  const pct = node.budget > 0 ? Math.round((node.consumed / node.budget) * 100) : 0
  const hasChildren = node.children && node.children.length > 0
  const parent = node.parentId ? findBudgetNode(tree, node.parentId) : null
  const unallocated = computeUnallocated(node)

  return (
    <>
      <TableRow className="border-border/40 hover:bg-indigo-50/30">
        <TableCell>
          <div className="flex items-center" style={{ paddingLeft: `${depth * 20}px` }}>
            {hasChildren && (
              <button
                type="button"
                onClick={() => setExpanded(!expanded)}
                className="mr-2 text-indigo-400 hover:text-indigo-300 text-xs w-4"
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
          <div className="flex items-center gap-2">
            <Progress value={pct} className="flex-1 h-2" />
            <span
              className={`text-xs ${pct >= 90 ? 'text-red-500' : pct >= 70 ? 'text-amber-500' : 'text-muted-foreground'}`}
            >
              {pct}%
            </span>
          </div>
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
  const { open } = useWorkflow()
  const budgetCta = useDemoCta('BUDGET')
  const [tree, setTree] = useState<BudgetNode[]>([])

  const load = useCallback(async () => {
    const t = await budgetApi.getTree()
    setTree(t)
  }, [])

  useEffect(() => {
    void budgetApi.getTree().then(setTree)
  }, [])

  const root = tree[0]
  const summary = root
    ? {
        budget: root.budget,
        consumed: root.consumed,
        unallocated: computeUnallocated(root),
      }
    : { budget: 0, consumed: 0, unallocated: 0 }

  const handleAllocate = (node: BudgetNode, parent: BudgetNode | null) => {
    open('budget-node-edit', { node, parent, onSuccess: load })
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="grid grid-cols-3 gap-4 flex-1 max-w-2xl">
          <Card className="shadow-card border-border/50">
            <CardContent className="pt-4 pb-3">
              <p className="text-xs text-muted-foreground">总预算</p>
              <p className="text-lg font-semibold">¥{summary.budget.toLocaleString()}</p>
            </CardContent>
          </Card>
          <Card className="shadow-card border-border/50">
            <CardContent className="pt-4 pb-3">
              <p className="text-xs text-muted-foreground">已用</p>
              <p className="text-lg font-semibold">¥{summary.consumed.toLocaleString()}</p>
            </CardContent>
          </Card>
          <Card className="shadow-card border-border/50">
            <CardContent className="pt-4 pb-3">
              <p className="text-xs text-muted-foreground">未分配</p>
              <p className="text-lg font-semibold">¥{summary.unallocated.toLocaleString()}</p>
            </CardContent>
          </Card>
        </div>
        <Badge variant="outline" className="border-indigo-200 text-indigo-600 text-xs">
          周期：2026 年 6 月
        </Badge>
      </div>

      <Card className="shadow-card border-border/50">
        <CardContent className="pt-5 pb-4">
          <Table>
            <TableHeader>
              <TableRow className="border-border/50 hover:bg-transparent">
                <TableHead className="text-xs font-semibold text-muted-foreground">节点</TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">
                  预算
                </TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">
                  已消耗
                </TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">
                  预留池
                </TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground text-right">
                  未分配
                </TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground w-40">
                  进度
                </TableHead>
                <TableHead className="text-xs font-semibold text-muted-foreground w-[120px]">
                  操作
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tree.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="p-0 border-0">
                    <EmptyState
                      icon={PieChart}
                      title="暂无预算数据"
                      description="请先导入组织后再分配预算"
                    />
                  </TableCell>
                </TableRow>
              ) : (
                tree.map((node) => (
                  <BudgetRow
                    key={node.id}
                    node={node}
                    depth={0}
                    tree={tree}
                    onAllocate={handleAllocate}
                    allocateHighlight={budgetCta.className}
                    allocateCtaId={budgetCta.id}
                  />
                ))
              )}
            </TableBody>
          </Table>
          <p className="text-xs text-muted-foreground mt-4">
            超限行为由全局{' '}
            <Link to="/budget/alerts" className="text-indigo-600 hover:underline">
              超限策略
            </Link>{' '}
            统一配置。预算周期为自然月，月初已用额度清零由后端处理，Demo 不模拟月重置。
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
