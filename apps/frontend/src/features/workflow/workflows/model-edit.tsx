import { useState } from 'react'
import { toast } from 'sonner'
import { useInjectedApis } from '@/api/use-apis'
import { isCustomModel } from '@/features/models'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Eye, EyeOff } from 'lucide-react'
import { useWorkflow } from '../hooks/use-workflow'
import { workflowErrorMessage } from '../lib/error-message'

export function ModelEditWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'model-edit'>) {
  const apis = useInjectedApis()
  const { closeAll } = useWorkflow()
  const model = entry.payload.model
  const onSuccess = entry.payload.onSuccess as ((id?: string) => void) | undefined
  const [displayName, setDisplayName] = useState(model.name)
  const [description, setDescription] = useState(model.description)
  const [endpoint, setEndpoint] = useState(model.endpoint ?? '')
  const [apiKey, setApiKey] = useState(model.apiKey ?? '')
  const [apiKeyVisible, setApiKeyVisible] = useState(false)
  const [endpointModelName, setEndpointModelName] = useState(model.endpointModelName ?? '')
  const [maxContext, setMaxContext] = useState(String(model.maxContext || 1000000))
  const [maxTokens, setMaxTokens] = useState(model.maxTokens ? String(model.maxTokens) : '')
  const [inputPrice, setInputPrice] = useState(String(model.inputPrice))
  const [outputPrice, setOutputPrice] = useState(String(model.outputPrice))
  const [submitting, setSubmitting] = useState(false)

  const canSubmit =
    displayName.trim() && (!isCustomModel(model) || (endpoint.trim() && apiKey.trim()))

  const handleSubmit = async () => {
    if (!canSubmit) return
    setSubmitting(true)
    try {
      await apis.modelApi.update(model.modelId, {
        name: displayName.trim(),
        description: description.trim(),
        endpoint: isCustomModel(model) ? endpoint.trim() : undefined,
        apiKey: isCustomModel(model) ? apiKey.trim() : undefined,
        endpointModelName: isCustomModel(model) ? endpointModelName.trim() : undefined,
        inputPrice: Number(inputPrice),
        outputPrice: Number(outputPrice),
        maxContext: Number(maxContext),
        maxTokens: Number(maxTokens) || undefined,
        capabilities: model.capabilities ?? ['chat'],
      })
      toast.success('模型信息已更新')
      onSuccess?.(model.modelId)
      closeAll()
    } catch (err) {
      toast.error(workflowErrorMessage(err, '更新失败'))
    } finally {
      setSubmitting(false)
    }
  }

  const markDirty = () => onSetDirty(true)

  return (
    <WorkflowPanelChrome
      title="编辑自定义模型"
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={submitting ? '保存中...' : '保存'}
          onPrimary={handleSubmit}
          primaryDisabled={!canSubmit || submitting}
        />
      }
    >
      <WorkflowFormLayout>
        <div className="space-y-1.5">
          <Label>
            模型显示名称 <span className="text-destructive">*</span>
          </Label>
          <Input
            value={displayName}
            onChange={(e) => {
              setDisplayName(e.target.value)
              markDirty()
            }}
          />
        </div>
        <div className="space-y-1.5">
          <Label>描述</Label>
          <Textarea
            value={description}
            onChange={(e) => {
              setDescription(e.target.value)
              markDirty()
            }}
            rows={3}
          />
        </div>
        {isCustomModel(model) && (
          <>
            <div className="space-y-1.5">
              <Label>
                API Key <span className="text-destructive">*</span>
              </Label>
              <div className="relative">
                <Input
                  value={apiKey}
                  onChange={(e) => {
                    setApiKey(e.target.value)
                    markDirty()
                  }}
                  placeholder="在此输入您的 API Key"
                  type={apiKeyVisible ? 'text' : 'password'}
                  className="pr-9"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="absolute right-0 top-0 h-full w-8 text-muted-foreground hover:text-foreground"
                  onClick={() => setApiKeyVisible(!apiKeyVisible)}
                  aria-label={apiKeyVisible ? '隐藏 API Key' : '显示 API Key'}
                >
                  {apiKeyVisible ? <EyeOff className="size-3.5" /> : <Eye className="size-3.5" />}
                </Button>
              </div>
            </div>
            <div className="space-y-1.5">
              <Label>
                API endpoint URL <span className="text-destructive">*</span>
              </Label>
              <Input
                value={endpoint}
                onChange={(e) => {
                  setEndpoint(e.target.value)
                  markDirty()
                }}
                placeholder="https://api.example.com/v1"
              />
            </div>
            <div className="space-y-1.5">
              <Label>API endpoint中的模型名称</Label>
              <Input
                value={endpointModelName}
                onChange={(e) => {
                  setEndpointModelName(e.target.value)
                  markDirty()
                }}
                placeholder="endpoint model name"
              />
            </div>
          </>
        )}
        {/* Completion mode hidden for now
        <div className="space-y-1.5">
          <Label>Completion mode</Label>
          <Select
            value={completionMode}
            onValueChange={(v) => {
              setCompletionMode(v)
              markDirty()
            }}
          >
            <SelectTrigger className="h-9">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {COMPLETION_MODES.map((mode) => (
                <SelectItem key={mode.value} value={mode.value}>
                  {mode.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        */}
        <div className="space-y-1.5">
          <Label>
            模型上下文长度 <span className="text-destructive">*</span>
          </Label>
          <Input
            type="number"
            min={1}
            value={maxContext}
            onChange={(e) => {
              setMaxContext(e.target.value)
              markDirty()
            }}
          />
        </div>
        <div className="space-y-1.5">
          <Label>最大 token 上限</Label>
          <Input
            type="number"
            min={0}
            value={maxTokens}
            onChange={(e) => {
              setMaxTokens(e.target.value)
              markDirty()
            }}
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
                markDirty()
              }}
              disabled={!isCustomModel(model)}
            />
            {!isCustomModel(model) && (
              <p className="text-xs text-muted-foreground">内置模型价格由平台同步，不可编辑</p>
            )}
          </div>
          <div className="space-y-1.5">
            <Label>输出单价 (¥/M tokens)</Label>
            <Input
              type="number"
              value={outputPrice}
              onChange={(e) => {
                setOutputPrice(e.target.value)
                markDirty()
              }}
              disabled={!isCustomModel(model)}
            />
          </div>
        </div>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
