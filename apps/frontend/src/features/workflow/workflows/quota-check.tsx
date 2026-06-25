import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowAlertPanel } from '../components/workflow-alert-panel'

export function QuotaCheckWorkflow({
  entry,
  onPop,
  onClose,
}: WorkflowComponentProps<'quota-check'>) {
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
      <WorkflowAlertPanel
        title="预留池额度不足，无法通过审批"
        description={`申请额度 ¥${requested.toLocaleString()}，当前预留池剩余 ¥${reservedPool.toLocaleString()}。请先调整预算分配或拒绝此申请。`}
      />
    </WorkflowPanelChrome>
  )
}
