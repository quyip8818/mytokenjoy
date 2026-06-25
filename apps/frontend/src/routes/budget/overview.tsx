import { Link } from 'react-router'
import { PieChart } from 'lucide-react'
import { StatCard } from '@/components/ui/stat-card'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatusBadge } from '@/components/ui/status-badge'
import { listEmpty } from '@/lib/list-empty'
import { ROUTES } from '@/config/routes'
import { BudgetRow } from '@/routes/budget/components/budget-row'
import { useBudgetOverviewPage } from '@/routes/budget/hooks/use-budget-overview-page'

export default function BudgetOverviewPage() {
  const {
    tree,
    loading,
    error,
    refresh,
    summary,
    periodLabel,
    canAllocate,
    budgetCta,
    handleAllocate,
    handleMemberQuota,
  } = useBudgetOverviewPage()

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
      actions={<StatusBadge variant="info">周期：{loading ? '-' : periodLabel}</StatusBadge>}
    >
      <DataSection
        loading={loading}
        error={error}
        onRetry={refresh}
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
                onMemberQuota={handleMemberQuota}
                allocateHighlight={budgetCta.className}
                allocateCtaId={budgetCta.id}
                canAllocate={canAllocate}
              />
            ))}
          </TableBody>
        </Table>
        <p className="mt-4 text-xs text-muted-foreground">
          超限行为由全局{' '}
          <Link to={ROUTES.budgetAlerts} className="text-blue-600 hover:underline">
            超限策略
          </Link>{' '}
          统一配置。预算周期为自然月，月初已用额度清零由后端处理，Demo 不模拟月重置。
        </p>
      </DataSection>
    </PageShell>
  )
}
