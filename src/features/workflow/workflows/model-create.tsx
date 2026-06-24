import { useState } from 'react'
import { toast } from 'sonner'
import { modelApi } from '@/api/models'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useWorkflow } from '../use-workflow'

export function ModelCreateWorkflow({ entry, onClose, onSetDirty }: WorkflowComponentProps) {
  const { closeAll } = useWorkflow()
  const onSuccess = entry.payload.onSuccess as ((id?: string) => void) | undefined
  const [name, setName] = useState('')
  const [baseUrl, setBaseUrl] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [inputPrice, setInputPrice] = useState('10')
  const [outputPrice, setOutputPrice] = useState('30')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async () => {
    if (!name.trim() || !baseUrl.trim()) return
    setSubmitting(true)
    try {
      const created = await modelApi.create({
        name: name.trim(),
        displayName: name.trim(),
        baseUrl: baseUrl.trim(),
        apiKey,
        inputPrice: Number(inputPrice),
        outputPrice: Number(outputPrice),
      })
      toast.success('模型已添加')
      onSuccess?.(created.id)
      closeAll()
    } catch {
      toast.error('添加失败')
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
          primaryDisabled={!name.trim() || !baseUrl.trim() || submitting}
        />
      }
    >
      <div className="max-w-md space-y-5">
        <div className="space-y-1.5">
          <Label>模型名称</Label>
          <Input
            value={name}
            onChange={(e) => {
              setName(e.target.value)
              onSetDirty(true)
            }}
            placeholder="my-custom-model"
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
        <div className="space-y-1.5">
          <Label>API Key</Label>
          <Input
            type="password"
            value={apiKey}
            onChange={(e) => {
              setApiKey(e.target.value)
              onSetDirty(true)
            }}
            placeholder="sk-..."
          />
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <Label>输入价格 (¥/1M)</Label>
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
            <Label>输出价格 (¥/1M)</Label>
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
      </div>
    </WorkflowPanelChrome>
  )
}
