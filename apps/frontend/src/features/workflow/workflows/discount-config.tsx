import { useState } from 'react'
import { toast } from '@/lib/toast'
import { useInjectedApis } from '@/api/use-apis'
import { useInjectedQuery } from '@/features/query/use-injected-query'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { workflowErrorMessage } from '../lib/error-message'

function formatDiscount(d: number): string {
  if (d === 1) return '无优惠'
  if (d < 1) return `${Math.round(d * 100)}% (${Math.round((1 - d) * 100)}% off)`
  return `${Math.round(d * 100)}% (加价 ${Math.round((d - 1) * 100)}%)`
}

export function DiscountConfigWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'discount-config'>) {
  const apis = useInjectedApis()
  const { companyId, companyName, onSuccess } = entry.payload

  const {
    data: discounts = [],
    loading,
    refresh,
  } = useInjectedQuery({
    queryKey: ['platform', 'discounts', companyId],
    queryFn: (a) => a.platformApi.listCompanyDiscounts(companyId),
  })

  const [modelType, setModelType] = useState('*')
  const [discount, setDiscount] = useState('0.80')
  const [note, setNote] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const canSubmit = modelType.trim() && Number(discount) > 0

  const handleSubmit = async () => {
    if (!canSubmit) return
    setSubmitting(true)
    try {
      await apis.platformApi.setCompanyDiscount(companyId, {
        modelType: modelType.trim(),
        discount: Number(discount),
        note: note || undefined,
      })
      toast.success('优惠已保存')
      setNote('')
      onSetDirty(false)
      onSuccess?.()
      void refresh()
    } catch (err) {
      toast.error(workflowErrorMessage(err, '保存失败'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title={`${companyName} — 模型优惠配置`}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={submitting ? '保存中…' : '保存'}
          onPrimary={handleSubmit}
          primaryDisabled={!canSubmit || submitting}
        />
      }
    >
      <WorkflowFormLayout>
        {/* Current discounts */}
        <div>
          <h4 className="text-sm font-medium text-muted-foreground mb-2">当前优惠</h4>
          {loading ? (
            <p className="text-sm text-muted-foreground">加载中…</p>
          ) : discounts.length === 0 ? (
            <p className="text-sm text-muted-foreground">暂无优惠配置</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>模型</TableHead>
                  <TableHead>折扣</TableHead>
                  <TableHead>备注</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {discounts.map((d) => (
                  <TableRow key={d.modelType}>
                    <TableCell className="font-mono text-xs">
                      {d.modelType === '*' ? '所有模型' : d.modelType}
                    </TableCell>
                    <TableCell className="tabular-nums">{formatDiscount(d.discount)}</TableCell>
                    <TableCell className="text-muted-foreground text-xs">
                      {d.note || '-'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>

        {/* Add/update form */}
        <div className="space-y-3 border-t pt-4">
          <h4 className="text-sm font-medium text-muted-foreground">新增 / 更新优惠</h4>

          <div className="space-y-1.5">
            <Label>模型类型</Label>
            <Input
              value={modelType}
              onChange={(e) => {
                setModelType(e.target.value)
                onSetDirty(true)
              }}
              placeholder="* = 所有模型，或输入具体模型名"
            />
            <span className="text-xs text-muted-foreground block">
              输入 * 表示通配（未单独配置的模型统一适用）
            </span>
          </div>

          <div className="space-y-1.5">
            <Label>折扣系数</Label>
            <Input
              type="number"
              step="0.01"
              min="0.01"
              value={discount}
              onChange={(e) => {
                setDiscount(e.target.value)
                onSetDirty(true)
              }}
              placeholder="0.80 = 八折"
            />
            <span className="text-xs text-muted-foreground block">
              0.8 = 八折, 1.0 = 无优惠, 1.2 = 加价20%
            </span>
          </div>

          <div className="space-y-1.5">
            <Label>备注</Label>
            <Input
              value={note}
              onChange={(e) => {
                setNote(e.target.value)
                onSetDirty(true)
              }}
              placeholder="合同优惠"
            />
          </div>
        </div>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
