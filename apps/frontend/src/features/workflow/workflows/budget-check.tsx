import { useNavigate } from '@tanstack/react-router'
import { AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { formatMoney } from '@/lib/quota-display'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { useWorkflow } from '../hooks/use-workflow'

export function BudgetCheckWorkflow({ entry, onPop, onClose }: WorkflowComponentProps<'budget-check'>) {
  const navigate = useNavigate()
  const { closeAll } = useWorkflow()
  const reservedPool = entry.payload.reservedPool ?? 0
  const requested = entry.payload.requested ?? 0

  return (
    <WorkflowPanelChrome
      title="额度不足"
      showBack
      onBack={onPop}
      onClose={onClose}
      footer={<WorkflowPanelFooter primaryLabel="知道了" onPrimary={onPop} />}
    >
      <WorkflowFormLayout
        variant="wide"
        className="flex flex-col items-center justify-center py-12 text-center"
      >
        <div className="flex h-14 w-14 items-center justify-center rounded-full bg-amber-50">
          <AlertTriangle className="h-7 w-7 text-amber-600" />
        </div>
        <div className="max-w-sm space-y-2">
          <p className="font-semibold">预留池额度不足，无法通过审批</p>
          <p className="text-sm text-muted-foreground">
            申请额度 {formatMoney(requested)}，当前预留池剩余 {formatMoney(reservedPool)}。请先调整预算分配或拒绝此申请。
          </p>
          <Button
            variant="link"
            className="text-sm"
            onClick={() => {
              closeAll()
              void navigate({ to: '/budget' })
            }}
          >
            前往预算管理 →
          </Button>
        </div>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
