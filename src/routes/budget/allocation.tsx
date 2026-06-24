import { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import { Wallet, MoreHorizontal } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { budgetApi } from '@/api/budget'
import type { BudgetGroup, BudgetNode } from '@/api/types'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { EmptyState } from '@/components/ui/empty-state'
import { useRowHighlight } from '@/lib/use-row-highlight'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

export default function BudgetAllocationPage() {
  const { open } = useWorkflow()
  const { flashRow, rowClass } = useRowHighlight()
  const [groups, setGroups] = useState<BudgetGroup[]>([])
  const [tree, setTree] = useState<BudgetNode[]>([])

  const load = useCallback(async () => {
    const [g, t] = await Promise.all([budgetApi.getGroups(), budgetApi.getTree()])
    setGroups(g)
    setTree(t)
  }, [])

  useEffect(() => {
    void Promise.all([budgetApi.getGroups(), budgetApi.getTree()]).then(([g, t]) => {
      setGroups(g)
      setTree(t)
    })
  }, [])

  const handleDelete = async (id: string) => {
    await budgetApi.deleteGroup(id)
    toast.success('预算组已删除')
    void load()
  }

  const openForm = (group?: BudgetGroup) => {
    open('budget-group-form', {
      group,
      tree,
      onSuccess: (id?: string) => {
        void load()
        if (id) flashRow(id)
      },
    })
  }

  return (
    <div className="space-y-6">
      <p className="text-sm text-muted-foreground">
        预算总览管理组织树逐级分配；本页管理独立于组织树的 Budget Group（虚拟项目组）。
      </p>
      <div className="flex items-center justify-end">
        <Button
          size="sm"
          className="bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white shadow-button"
          onClick={() => openForm()}
        >
          新建预算组
        </Button>
      </div>

      <Card className="shadow-card border-border/50">
        <CardContent className="pt-5 pb-4">
          {groups.length === 0 ? (
            <EmptyState
              icon={Wallet}
              title="暂无预算组"
              description="创建预算组以管理虚拟项目组的独立预算"
              actionLabel="新建预算组"
              onAction={() => openForm()}
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow className="border-border/50 hover:bg-transparent">
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    名称
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground text-right">
                    预算 (¥)
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground text-right">
                    已消耗 (¥)
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground w-40">
                    进度
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground">
                    关联
                  </TableHead>
                  <TableHead className="text-xs font-semibold text-muted-foreground w-[120px]">
                    操作
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {groups.map((g) => {
                  const pct = Math.round((g.consumed / g.budget) * 100)
                  return (
                    <TableRow key={g.id} className={rowClass(g.id)}>
                      <TableCell className="font-medium">{g.name}</TableCell>
                      <TableCell className="text-right">{g.budget.toLocaleString()}</TableCell>
                      <TableCell className="text-right">{g.consumed.toLocaleString()}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Progress value={pct} className="flex-1 h-2" />
                          <span className="text-xs text-muted-foreground">{pct}%</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline" className="border-indigo-200 text-indigo-600">
                          {g.memberIds.length} 人
                        </Badge>
                        <Badge variant="outline" className="ml-1 border-indigo-200 text-indigo-600">
                          {g.departmentIds.length} 部门
                        </Badge>
                      </TableCell>
                      <TableCell>
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
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
