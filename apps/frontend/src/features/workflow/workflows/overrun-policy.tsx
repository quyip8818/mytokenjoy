import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { budgetApi } from '@/api/budget'
import type { OverrunPolicyConfig } from '@/api/types'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Checkbox } from '@/components/ui/checkbox'
import { Button } from '@/components/ui/button'
import { useWorkflow } from '../use-workflow'

export function OverrunPolicyWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'overrun-policy'>) {
  const { closeAll } = useWorkflow()
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined
  const [config, setConfig] = useState<OverrunPolicyConfig | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    budgetApi.getOverrunPolicy().then(setConfig)
  }, [])

  const updateThreshold = (index: number, value: string) => {
    if (!config) return
    const thresholds = [...config.thresholds]
    thresholds[index] = Number(value)
    setConfig({ ...config, thresholds })
    onSetDirty(true)
  }

  const addThreshold = () => {
    if (!config) return
    setConfig({ ...config, thresholds: [...config.thresholds, 85] })
    onSetDirty(true)
  }

  const removeThreshold = (index: number) => {
    if (!config || config.thresholds.length <= 1) return
    setConfig({ ...config, thresholds: config.thresholds.filter((_, i) => i !== index) })
    onSetDirty(true)
  }

  const handleSave = async () => {
    if (!config) return
    setSubmitting(true)
    try {
      await budgetApi.updateOverrunPolicy(config)
      toast.success('策略已保存')
      onSuccess?.()
      closeAll()
    } catch {
      toast.error('保存失败')
    } finally {
      setSubmitting(false)
    }
  }

  if (!config) {
    return (
      <WorkflowPanelChrome title="全局超限策略" onClose={onClose}>
        <p className="text-muted-foreground text-sm">加载中...</p>
      </WorkflowPanelChrome>
    )
  }

  return (
    <WorkflowPanelChrome
      title="全局超限策略"
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={submitting ? '保存中...' : '保存'}
          onPrimary={handleSave}
          primaryDisabled={submitting}
        />
      }
    >
      <WorkflowFormLayout variant="wide">
        <div className="space-y-3">
          <Label>预警阈值 (%)</Label>
          {config.thresholds.map((t, i) => (
            <div key={i} className="flex items-center gap-2">
              <Input
                type="number"
                min={1}
                max={100}
                value={t}
                onChange={(e) => updateThreshold(i, e.target.value)}
                className="w-24"
              />
              {config.thresholds.length > 1 && (
                <Button type="button" variant="ghost" size="sm" onClick={() => removeThreshold(i)}>
                  删除
                </Button>
              )}
            </div>
          ))}
          <Button type="button" variant="outline" size="sm" onClick={addThreshold}>
            添加阈值
          </Button>
        </div>
        <div className="space-y-3">
          <Label>通知渠道</Label>
          <label className="flex items-center gap-2 text-sm">
            <Checkbox
              checked={config.notifyEmail}
              onCheckedChange={(c) => {
                setConfig({ ...config, notifyEmail: !!c })
                onSetDirty(true)
              }}
            />
            邮箱
          </label>
          <label className="flex items-center gap-2 text-sm">
            <Checkbox
              checked={config.notifyPhone}
              onCheckedChange={(c) => {
                setConfig({ ...config, notifyPhone: !!c })
                onSetDirty(true)
              }}
            />
            手机
          </label>
          <label className="flex items-center gap-2 text-sm">
            <Checkbox
              checked={config.notifyIm}
              onCheckedChange={(c) => {
                setConfig({ ...config, notifyIm: !!c })
                onSetDirty(true)
              }}
            />
            IM
          </label>
        </div>
        <div className="space-y-1.5">
          <Label>超限阻断文案</Label>
          <Textarea
            value={config.blockMessage}
            onChange={(e) => {
              setConfig({ ...config, blockMessage: e.target.value })
              onSetDirty(true)
            }}
            rows={3}
          />
        </div>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
