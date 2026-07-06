import { useState } from 'react'
import { toast } from 'sonner'
import { useApis } from '@/api/use-apis'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WORKFLOW_TABLE_HEAD_CLASS } from '../constants'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { useWorkflow } from '../use-workflow'

interface PreviewPayload {
  nodeId: string
  nodeName: string
  before: { budget: number; reservedPool: number }
  after: { budget: number; reservedPool: number }
  onSuccess?: () => void
}

export function BudgetImpactPreviewWorkflow({
  entry,
  onPop,
  onClose,
}: WorkflowComponentProps<'budget-impact-preview'>) {
  const apis = useApis()
  const { closeAll } = useWorkflow()
  const payload = entry.payload as unknown as PreviewPayload
  const [submitting, setSubmitting] = useState(false)

  const handleConfirm = async () => {
    setSubmitting(true)
    try {
      await apis.budgetApi.updateDepartment(payload.nodeId, {
        budget: payload.after.budget,
        reservedPool: payload.after.reservedPool,
      })
      toast.success('预算已更新')
      payload.onSuccess?.()
      closeAll()
    } catch {
      toast.error('保存失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title="影响范围预览"
      showBack
      onBack={onPop}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onPop}
          primaryLabel={submitting ? '保存中...' : '确认保存'}
          onPrimary={handleConfirm}
          primaryDisabled={submitting}
        />
      }
    >
      <WorkflowFormLayout variant="wide" className="space-y-4">
        <p className="text-sm text-muted-foreground">节点：{payload.nodeName}</p>
        <table className="w-full text-sm border border-border/50 rounded-lg overflow-hidden">
          <thead className={WORKFLOW_TABLE_HEAD_CLASS}>
            <tr>
              <th className="text-left px-3 py-2 font-medium">字段</th>
              <th className="text-right px-3 py-2 font-medium">原值</th>
              <th className="text-right px-3 py-2 font-medium">新值</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-t border-border/50">
              <td className="px-3 py-2">月度预算</td>
              <td className="px-3 py-2 text-right">¥{payload.before.budget.toLocaleString()}</td>
              <td className="px-3 py-2 text-right font-medium">
                ¥{payload.after.budget.toLocaleString()}
              </td>
            </tr>
            <tr className="border-t border-border/50">
              <td className="px-3 py-2">预留池</td>
              <td className="px-3 py-2 text-right">
                ¥{payload.before.reservedPool.toLocaleString()}
              </td>
              <td className="px-3 py-2 text-right font-medium">
                ¥{payload.after.reservedPool.toLocaleString()}
              </td>
            </tr>
          </tbody>
        </table>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
