import { AlertTriangle } from 'lucide-react'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'

export function QuotaCheckWorkflow({ entry, onPop, onClose }: WorkflowComponentProps) {
  const reservedPool = (entry.payload.reservedPool as number) ?? 0
  const requested = (entry.payload.requested as number) ?? 0

  return (
    <WorkflowPanelChrome
      title="额度不足"
      showBack
      onBack={onPop}
      onClose={onClose}
      footer={<WorkflowPanelFooter primaryLabel="知道了" onPrimary={onPop} />}
    >
      <div className="flex flex-col items-center justify-center py-12 text-center space-y-4">
        <div className="flex h-14 w-14 items-center justify-center rounded-full bg-amber-50">
          <AlertTriangle className="h-7 w-7 text-amber-600" />
        </div>
        <div className="space-y-2 max-w-sm">
          <p className="font-semibold">预留池额度不足，无法通过审批</p>
          <p className="text-sm text-muted-foreground">
            申请额度 ¥{requested.toLocaleString()}，当前预留池剩余 ¥{reservedPool.toLocaleString()}
            。 请先调整预算分配或拒绝此申请。
          </p>
        </div>
      </div>
    </WorkflowPanelChrome>
  )
}
