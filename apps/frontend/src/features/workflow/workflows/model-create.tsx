import { useState } from 'react'
import { toast } from 'sonner'
import { useInjectedApis } from '@/api/use-apis'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Eye, EyeOff } from 'lucide-react'
import { useWorkflow } from '../hooks/use-workflow'
import { workflowErrorMessage } from '../lib/error-message'

export function ModelCreateWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'model-create'>) {
  const apis = useInjectedApis()
  const { closeAll } = useWorkflow()
  const onSuccess = entry.payload.onSuccess as ((id?: string | number) => void) | undefined
  const [modelName, setModelName] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [apiKeyVisible, setApiKeyVisible] = useState(false)
  const [baseUrl, setBaseUrl] = useState('')
  const [endpointModelName, setEndpointModelName] = useState('')
  const [maxContext, setMaxContext] = useState('1000000')
  const [maxTokens, setMaxTokens] = useState('')
  const [inputPrice, setInputPrice] = useState('10')
  const [outputPrice, setOutputPrice] = useState('30')
  const [submitting, setSubmitting] = useState(false)

  const canSubmit = modelName.trim() && baseUrl.trim() && apiKey.trim() && Number(maxContext) > 0

  const handleSubmit = async () => {
    if (!canSubmit) return
    setSubmitting(true)
    try {
      const created = await apis.modelApi.create({
        type: modelName.trim(),
        name: displayName.trim() || modelName.trim(),
        baseUrl: baseUrl.trim(),
        apiKey: apiKey.trim() || undefined,
        endpointModelName: endpointModelName.trim() || undefined,
        inputPrice: Number(inputPrice),
        outputPrice: Number(outputPrice),
        maxContext: Number(maxContext),
        maxTokens: Number(maxTokens) || undefined,
        capabilities: ['chat'],
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

  const markDirty = () => onSetDirty(true)

  return (
    <WorkflowPanelChrome
      title="添加自定义模型"
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
            模型名称 <span className="text-destructive">*</span>
          </Label>
          <Input
            value={modelName}
            onChange={(e) => {
              setModelName(e.target.value)
              markDirty()
            }}
            placeholder="输入模型全称"
          />
        </div>
        <div className="space-y-1.5">
          <Label>模型显示名称</Label>
          <Input
            value={displayName}
            onChange={(e) => {
              setDisplayName(e.target.value)
              markDirty()
            }}
            placeholder="模型在界面的显示名称"
          />
        </div>
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
            value={baseUrl}
            onChange={(e) => {
              setBaseUrl(e.target.value)
              markDirty()
            }}
            placeholder="Base URL, e.g. https://api.openai.com/v1"
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
            placeholder="endpoint model name, e.g. chatgpt4.0"
          />
        </div>
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
            placeholder="4096"
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
            placeholder="4096"
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
            />
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
            />
          </div>
        </div>
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
