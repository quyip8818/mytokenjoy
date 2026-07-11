import type { BudgetNode, BudgetProjectView } from '@/api/types'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Progress } from '@/components/ui/progress'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { cn } from '@/lib/utils'
import { displayToPoints, formatDisplayCurrency, pointsToDisplay } from '@/lib/points'

type BudgetAllocationTableProps = {
  node: BudgetNode
  children: BudgetNode[]
  nodeProjects: BudgetProjectView[]
  overrunPolicyLabel: string
  editing: boolean
  drafts: Record<string, { budget: string }>
  onUpdateDraft: (id: string, value: string) => void
}

export function BudgetAllocationTable({
  node,
  children,
  nodeProjects,
  overrunPolicyLabel,
  editing,
  drafts,
  onUpdateDraft,
}: BudgetAllocationTableProps) {
  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <Table>
        <TableHeader>
          <TableRow className="border-border/50 hover:bg-transparent">
            <TableHead className="text-xs font-medium uppercase text-muted-foreground">
              名称
            </TableHead>
            <TableHead className="text-right text-xs font-medium uppercase text-muted-foreground">
              额度
            </TableHead>
            <TableHead className="text-right text-xs font-medium uppercase text-muted-foreground">
              已消耗
            </TableHead>
            <TableHead className="w-32 text-xs font-medium uppercase text-muted-foreground">
              进度
            </TableHead>
            <TableHead className="text-xs font-medium uppercase text-muted-foreground">
              超限策略
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {children.map((child) => {
            const childPct =
              child.budget > 0 ? Math.round((child.consumed / child.budget) * 100) : 0
            const draftBudget = drafts[child.id]?.budget ?? String(pointsToDisplay(child.budget))
            const draftValue = parseFloat(draftBudget)
            const budgetOver =
              editing && !Number.isNaN(draftValue) && displayToPoints(draftValue) > node.budget

            return (
              <TableRow key={child.id} className="even:bg-muted/40 hover:bg-muted/50">
                <TableCell className="font-medium">{child.name}</TableCell>
                <TableCell className="text-right">
                  {editing ? (
                    <Input
                      type="number"
                      min={0}
                      value={draftBudget}
                      onChange={(event) => onUpdateDraft(child.id, event.target.value)}
                      aria-label={`${child.name} 预算额度`}
                      className={cn(
                        'h-7 w-28 text-right tabular-nums',
                        budgetOver && 'border-red-500 focus-visible:ring-red-500/30',
                      )}
                    />
                  ) : (
                    <span className="tabular-nums">{formatDisplayCurrency(child.budget)}</span>
                  )}
                </TableCell>
                <TableCell className="text-right tabular-nums">
                  {formatDisplayCurrency(child.consumed)}
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <Progress value={childPct} className="h-1.5 flex-1" />
                    <span className="w-8 text-right text-xs tabular-nums text-muted-foreground">
                      {childPct}%
                    </span>
                  </div>
                </TableCell>
                <TableCell>
                  <Badge variant="outline" className="text-xs">
                    {overrunPolicyLabel}
                  </Badge>
                </TableCell>
              </TableRow>
            )
          })}
          {nodeProjects.map((project) => {
            const projectPct =
              project.budget > 0 ? Math.round((project.consumed / project.budget) * 100) : 0
            return (
              <TableRow key={project.id} className="even:bg-muted/40 hover:bg-muted/50">
                <TableCell className="font-medium text-muted-foreground">
                  {project.name}
                  <span className="ml-1.5 text-xs text-muted-foreground/60">(项目)</span>
                </TableCell>
                <TableCell className="text-right tabular-nums">
                  {formatDisplayCurrency(project.budget)}
                </TableCell>
                <TableCell className="text-right tabular-nums">
                  {formatDisplayCurrency(project.consumed)}
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <Progress value={projectPct} className="h-1.5 flex-1" />
                    <span className="w-8 text-right text-xs tabular-nums text-muted-foreground">
                      {projectPct}%
                    </span>
                  </div>
                </TableCell>
                <TableCell>
                  <Badge variant="outline" className="text-xs">
                    {overrunPolicyLabel}
                  </Badge>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
