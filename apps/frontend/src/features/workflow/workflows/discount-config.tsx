import { useMemo, useState } from 'react'
import { toast } from '@/lib/toast'
import { useInjectedApis } from '@/api/use-apis'
import { useInjectedQuery } from '@/features/query/use-injected-query'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ConfirmActionDialog, type ConfirmActionState } from '@/components/ui/confirm-action-dialog'
import { workflowErrorMessage } from '../lib/error-message'
import { platformKeys } from '@/features/platform/models/query-keys'

export function DiscountConfigWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'discount-config'>) {
  const apis = useInjectedApis()
  const { companyId, companyName, onSuccess } = entry.payload

  const { data: models = [], loading: modelsLoading } = useInjectedQuery({
    queryKey: platformKeys.models(),
    queryFn: (a) => a.platformApi.listModels(),
  })

  const { data: discounts = [], loading: discountsLoading } = useInjectedQuery({
    queryKey: ['platform', 'discounts', companyId],
    queryFn: (a) => a.platformApi.listCompanyDiscounts(companyId),
  })

  // Map modelType -> saved percentage string
  const initialPercentages = useMemo(() => {
    const map: Record<string, string> = {}
    for (const d of discounts) {
      map[d.modelType] = String(Math.round(d.discount * 100))
    }
    return map
  }, [discounts])

  // Local edits
  const [edits, setEdits] = useState<Record<string, string>>({})

  const getPercentage = (modelType: string): string => {
    if (modelType in edits) return edits[modelType]
    if (modelType in initialPercentages) return initialPercentages[modelType]
    return '100'
  }

  const handleChange = (modelType: string, value: string) => {
    setEdits((prev) => ({ ...prev, [modelType]: value }))
    onSetDirty(true)
  }

  // Reset all to 100%
  const handleReset = () => {
    const resetMap: Record<string, string> = {}
    for (const model of models) {
      resetMap[model.type] = '100'
    }
    setEdits(resetMap)
    onSetDirty(true)
  }

  const [submitting, setSubmitting] = useState(false)
  const [confirmState, setConfirmState] = useState<ConfirmActionState | null>(null)

  // Collect changed items + stale entries not in models list (e.g. old "*" wildcard)
  const changedItems = useMemo(() => {
    const items: { modelType: string; discount: number }[] = []
    const modelTypes = new Set(models.map((m) => m.type))

    // Changed model entries
    for (const model of models) {
      const current = getPercentage(model.type)
      const original = initialPercentages[model.type] ?? '100'
      if (current !== original) {
        const pct = Number(current)
        if (pct > 0) {
          items.push({ modelType: model.type, discount: pct / 100 })
        }
      }
    }

    // Stale entries in DB that are not in the current models list — auto-clean them
    for (const d of discounts) {
      if (!modelTypes.has(d.modelType)) {
        items.push({ modelType: d.modelType, discount: 1.0 })
      }
    }

    return items
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [models, edits, initialPercentages, discounts])

  const canSubmit = changedItems.length > 0

  // Find items that need confirmation (<80% or >120%)
  const extremeItems = useMemo(
    () => changedItems.filter((i) => i.discount < 0.8 || i.discount > 1.2),
    [changedItems],
  )

  const doSubmit = async () => {
    setSubmitting(true)
    try {
      await apis.platformApi.batchSetCompanyDiscounts(companyId, changedItems)
      toast.success('优惠已保存')
      onSetDirty(false)
      onSuccess?.()
    } catch (err) {
      toast.error(workflowErrorMessage(err, '保存失败'))
    } finally {
      setSubmitting(false)
    }
  }

  const handleSubmit = () => {
    if (!canSubmit) return
    if (extremeItems.length > 0) {
      const lines = extremeItems.map((i) => `${i.modelType}: ${Math.round(i.discount * 100)}%`)
      setConfirmState({
        open: true,
        title: '折扣幅度较大',
        desc: `以下模型设置超出常规范围（80%~120%）：\n${lines.join('、')}\n\n确认保存？`,
        variant: 'danger',
        confirmLabel: '确认保存',
        onConfirm: () => {
          setConfirmState(null)
          void doSubmit()
        },
      })
    } else {
      void doSubmit()
    }
  }

  const loading = modelsLoading || discountsLoading

  return (
    <>
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
          {loading ? (
            <p className="text-sm text-muted-foreground">加载中…</p>
          ) : (
            <>
              <div className="flex items-center justify-between mb-2">
                <p className="text-xs text-muted-foreground">
                  100% = 无优惠，80% = 八折，120% = 加价20%
                </p>
                <Button variant="outline" size="sm" onClick={handleReset}>
                  清理全部
                </Button>
              </div>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>模型</TableHead>
                    <TableHead className="w-28 text-right">百分比</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {models
                    .filter((m) => !m.deprecated)
                    .map((model) => {
                      const pct = Number(getPercentage(model.type))
                      const isExtreme = pct > 0 && (pct < 80 || pct > 120)
                      return (
                        <TableRow key={model.type}>
                          <TableCell className="font-mono text-xs">{model.type}</TableCell>
                          <TableCell className="text-right">
                            <div className="flex items-center justify-end gap-1">
                              <Input
                                type="number"
                                min="1"
                                step="1"
                                className={`w-20 text-right tabular-nums h-7 text-sm ${isExtreme ? 'border-destructive' : ''}`}
                                value={getPercentage(model.type)}
                                onChange={(e) => handleChange(model.type, e.target.value)}
                              />
                              <span className="text-xs text-muted-foreground">%</span>
                            </div>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                </TableBody>
              </Table>
            </>
          )}
        </WorkflowFormLayout>
      </WorkflowPanelChrome>

      <ConfirmActionDialog
        state={confirmState}
        onOpenChange={(open) => {
          if (!open) setConfirmState(null)
        }}
        onClose={() => setConfirmState(null)}
      />
    </>
  )
}
