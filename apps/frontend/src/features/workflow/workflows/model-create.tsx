import { useState } from 'react'
import { toast } from 'sonner'
import { useInjectedApis } from '@/api/use-apis'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useWorkflow } from '../use-workflow'
import { workflowErrorMessage } from '../lib/error-message'

export function ModelCreateWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'model-create'>) {
  const apis = useInjectedApis()
  const { closeAll } = useWorkflow()
  const onSuccess = entry.payload.onSuccess as ((id?: string | number) => void) | undefined
  const [callType, setCallType] = useState('')
  const [name, setName] = useState('')
  const [baseUrl, setBaseUrl] = useState('')
  const [inputPrice, setInputPrice] = useState('10')
  const [outputPrice, setOutputPrice] = useState('30')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async () => {
    if (!callType.trim() || !name.trim() || !baseUrl.trim()) return
    setSubmitting(true)
    try {
      const created = await apis.modelApi.create({
        type: callType.trim(),
        name: name.trim(),
        baseUrl: baseUrl.trim(),
        inputPrice: Number(inputPrice),
        outputPrice: Number(outputPrice),
      })
      toast.success('模型已添加')
      onSuccess?.(created.modelId)
      closeAll()
    } catch (err) {
      toast.error(workflowErrorMessage(err, '添加失败'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title="添加自定义模型"
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={submitting ? '保存中...' : '保存'}
          onPrimary={handleSubmit}
          primaryDisabled={!callType.trim() || !name.trim() || !baseUrl.trim() || submitting}
        />
      }
    >
      <WorkflowFormLayout>
        <div className="space-y-1.5">
          <Label>调用标识 (callType)</Label>
          <Input
            value={callType}
            onChange={(e) => {
              setCallType(e.target.value)
              onSetDirty(true)
            }}
            placeholder="my-custom-model"
          />
        </div>
        <div className="space-y-1.5">
          <Label>展示名称</Label>
          <Input
            value={name}
            onChange={(e) => {
              setName(e.target.value)
              onSetDirty(true)
            }}
            placeholder="My Custom Model"
          />
        </div>
        <div className="space-y-1.5">
          <Label>Base URL</Label>
          <Input
            value={baseUrl}
            onChange={(e) => {
              setBaseUrl(e.target.value)
              onSetDirty(true)
            }}
            placeholder="https://api.example.com/v1"
          />
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <Label>输入单价 (¥/M tokens)</Label>
            <Input
              type="number"
              value={inputPrice}
              onChange={(e) => {
                setInputPrice(e.target.value)
                onSetDirty(true)
              }}
            />
          </div>
          <div className="space-y-1.5">
            <Label>输出单价 (¥/M tokens)</Label>
            <Input
              type="number"
              value={outputPrice}
              onChange={(e) => {
                setOutputPrice(e.target.value)
                onSetDirty(true)
              }}
            />
          </div>
        </div>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
