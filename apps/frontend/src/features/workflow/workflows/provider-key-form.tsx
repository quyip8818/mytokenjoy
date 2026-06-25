import { useState } from 'react'
import { toast } from 'sonner'
import { providerKeyApi } from '@/api/keys'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useWorkflow } from '../use-workflow'

const PROVIDERS = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'deepseek', label: 'DeepSeek' },
  { value: 'qwen', label: '通义千问' },
  { value: 'custom', label: '自定义' },
]

export function ProviderKeyFormWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'provider-key-form'>) {
  const { closeAll } = useWorkflow()
  const onSuccess = entry.payload.onSuccess as (() => void) | undefined
  const [provider, setProvider] = useState('openai')
  const [name, setName] = useState('')
  const [keyValue, setKeyValue] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async () => {
    if (!name || !keyValue) return
    setSubmitting(true)
    try {
      await providerKeyApi.create({ provider, name, key: keyValue })
      toast.success('供应商 Key 已添加')
      onSuccess?.()
      closeAll()
    } catch {
      toast.error('添加失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title="添加供应商 Key"
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={submitting ? '添加中...' : '添加'}
          onPrimary={handleSubmit}
          primaryDisabled={!name || !keyValue || submitting}
        />
      }
    >
      <WorkflowFormLayout>
        <div className="space-y-1.5">
          <Label>供应商</Label>
          <Select
            value={provider}
            onValueChange={(v) => {
              if (v) setProvider(v)
              onSetDirty(true)
            }}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {PROVIDERS.map((p) => (
                <SelectItem key={p.value} value={p.value}>
                  {p.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-1.5">
          <Label>名称</Label>
          <Input
            value={name}
            onChange={(e) => {
              setName(e.target.value)
              onSetDirty(true)
            }}
            placeholder="如：OpenAI 生产"
          />
        </div>
        <div className="space-y-1.5">
          <Label>API Key</Label>
          <Input
            type="password"
            value={keyValue}
            onChange={(e) => {
              setKeyValue(e.target.value)
              onSetDirty(true)
            }}
            placeholder="sk-..."
          />
        </div>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
