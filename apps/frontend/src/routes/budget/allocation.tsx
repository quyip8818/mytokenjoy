import { Wallet, MoreHorizontal } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { DataSection } from '@/components/layout/data-section'
import { PageShell } from '@/components/layout/page-shell'
import { StatusBadge } from '@/components/ui/status-badge'
import { BudgetProgressCell } from '@/components/ui/budget-progress-cell'
import { listEmpty } from '@/lib/list-empty'
import { PermissionGate } from '@/components/auth/permission-gate'
import { PERMISSION } from '@/lib/permissions'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useBudgetAllocationPage } from '@/routes/budget/hooks/use-budget-allocation-page'

export default function BudgetAllocationPage() {
  const { groups, loading, error, refresh, canWrite, rowClass, handleDelete, openForm } =
    useBudgetAllocationPage()

  return (
    <PageShell
      description={
        <p className="text-sm text-muted-foreground">
          预算总览管理组织树逐级分配；本页管理独立于组织树的 Budget Group（虚拟项目组）。
        </p>
      }
      actions={
        <PermissionGate write permission={PERMISSION.BUDGET_ALLOCATE}>
          <Button size="sm" variant="brand" onClick={() => openForm()}>
            新建预算组
          </Button>
        </PermissionGate>
      }
    >
      <DataSection
        loading={loading}
        error={error}
        onRetry={refresh}
        skeletonColumns={6}
        empty={listEmpty(loading, groups, {
          icon: Wallet,
          title: '暂无预算组',
          description: '创建预算组以管理虚拟项目组的独立预算',
          actionLabel: '新建预算组',
          onAction: () => openForm(),
        })}
      >
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead>名称</TableHead>
              <TableHead className="text-right">预算 (¥)</TableHead>
              <TableHead className="text-right">已消耗 (¥)</TableHead>
              <TableHead className="w-40">进度</TableHead>
              <TableHead>关联</TableHead>
              <TableHead className="w-[120px]">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {groups.map((g) => (
              <TableRow key={g.id} className={rowClass(g.id)}>
                <TableCell className="font-medium">{g.name}</TableCell>
                <TableCell className="text-right">{g.budget.toLocaleString()}</TableCell>
                <TableCell className="text-right">{g.consumed.toLocaleString()}</TableCell>
                <TableCell className="w-40">
                  <BudgetProgressCell value={g.consumed} total={g.budget} />
                </TableCell>
                <TableCell>
                  <StatusBadge variant="info">{g.memberIds.length} 人</StatusBadge>
                  <StatusBadge variant="info" className="ml-1">
                    {g.departmentIds.length} 部门
                  </StatusBadge>
                </TableCell>
                <TableCell>
                  {canWrite ? (
                    <DropdownMenu>
                      <DropdownMenuTrigger
                        render={
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        }
                      />
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => openForm(g)}>管理</DropdownMenuItem>
                        <DropdownMenuItem
                          className="text-red-600"
                          onClick={() => handleDelete(g.id)}
                        >
                          删除
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </DataSection>
    </PageShell>
  )
}
