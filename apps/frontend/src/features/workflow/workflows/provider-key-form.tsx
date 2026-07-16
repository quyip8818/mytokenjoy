import { useState } from 'react'
import { toast } from 'sonner'
import { useInjectedApis } from '@/api/use-apis'
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
import { useWorkflow } from '../hooks/use-workflow'
import { workflowErrorMessage } from '../lib/error-message'
import { PROVIDER_LABELS } from '@/lib/provider-labels'

export function ProviderKeyFormWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'provider-key-form'>) {
  const apis = useInjectedApis()
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
      await apis.providerKeyApi.create({ provider, name, key: keyValue })
      toast.success('供应商 Key 已添加')
      onSuccess?.()
      closeAll()
    } catch (err) {
      toast.error(workflowErrorMessage(err, '添加失败'))
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
              {Object.entries(PROVIDER_LABELS).map(([value, label]) => (
                <SelectItem key={value} value={value}>
                  {label}
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
