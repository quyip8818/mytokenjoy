import { useState } from 'react'
import type { BudgetNode, MemberBudgetQuota, UpdateMemberBudgetInput } from '@/api/types'
import { Button } from '@/components/ui/button'
import { formatDisplayCurrency } from '@/lib/points'
import { Users, Wallet, Settings2 } from 'lucide-react'
import { BudgetMemberBudgetDialog } from './budget-member-budget-dialog'

interface BudgetEditMemberBudgetProps {
  node: BudgetNode
  onUpdated: () => void
  getMemberBudgets: (departmentId: string) => Promise<MemberBudgetQuota[]>
  updateMemberBudget: (
    memberId: string,
    data: UpdateMemberBudgetInput,
  ) => Promise<MemberBudgetQuota>
}

export function BudgetEditMemberBudget({
  node,
  onUpdated,
  getMemberBudgets,
  updateMemberBudget,
}: BudgetEditMemberBudgetProps) {
  const [budgetDialogOpen, setBudgetDialogOpen] = useState(false)

  return (
    <div className="rounded-lg border border-border p-4">
      <div className="mb-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Users className="size-4 text-muted-foreground" />
          <h4 className="text-sm font-semibold text-foreground">成员额度</h4>
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 gap-1.5 text-xs text-muted-foreground"
          onClick={() => setBudgetDialogOpen(true)}
        >
          <Settings2 className="size-3.5" />
          配置成员额度
        </Button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="flex items-center gap-3 rounded-md bg-muted/50 px-3 py-2.5">
          <Wallet className="size-4 shrink-0 text-muted-foreground" />
          <div className="min-w-0 flex-1">
            <p className="text-xs text-muted-foreground">平均额度/人</p>
            <p className="text-sm font-medium tabular-nums text-muted-foreground">—</p>
          </div>
        </div>

        <div className="flex items-center gap-3 rounded-md bg-muted/50 px-3 py-2.5">
          <Users className="size-4 shrink-0 text-muted-foreground" />
          <div className="min-w-0 flex-1">
            <p className="text-xs text-muted-foreground">预留池余额</p>
            <p className="text-sm font-medium tabular-nums">
              {formatDisplayCurrency(node.budget - (node.children?.reduce((s, c) => s + c.budget, 0) ?? 0))}
            </p>
          </div>
        </div>
      </div>

      <BudgetMemberBudgetDialog
        open={budgetDialogOpen}
        onOpenChange={setBudgetDialogOpen}
        departmentId={node.id}
        getMemberBudgets={getMemberBudgets}
        updateMemberBudget={updateMemberBudget}
      />
    </div>
  )
}
