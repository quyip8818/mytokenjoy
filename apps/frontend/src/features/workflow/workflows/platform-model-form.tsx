import { useState } from 'react'
import { toast } from 'sonner'
import { useInjectedApis } from '@/api/use-apis'
import { useWorkflow } from '../hooks/use-workflow'
import { workflowErrorMessage } from '../lib/error-message'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const PROVIDERS = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'deepseek', label: 'DeepSeek' },
  { value: 'qwen', label: '通义千问' },
  { value: 'google', label: 'Google' },
  { value: 'azure', label: 'Azure' },
  { value: 'minimax', label: 'MiniMax' },
  { value: 'zhipu', label: '智谱' },
  { value: 'moonshot', label: 'Moonshot' },
  { value: 'custom', label: '自定义' },
] as const

const CAPABILITIES = ['chat', 'vision', 'audio', 'function_call', 'reasoning'] as const

// --- 共享表单字段 ---

interface FormFieldsProps {
  provider: string
  setProvider: (v: string) => void
  type: string
  setType?: (v: string) => void
  typeDisabled?: boolean
  name: string
  setName: (v: string) => void
  capabilities: string[]
  toggleCap: (cap: string) => void
  maxContext: string
  setMaxContext: (v: string) => void
  inputPrice: string
  setInputPrice: (v: string) => void
  outputPrice: string
  setOutputPrice: (v: string) => void
  cacheInputPrice: string
  setCacheInputPrice: (v: string) => void
  markDirty: () => void
}

/**
 * 表单居中、限宽 (max-w-2xl = 672px)，大内边距。
 * 在 75vw 宽面板中内容居中，四周留白充足，视觉舒适。
 * ponytail: 如果未来字段增多到 10+ 可考虑分 tab，当前 6 字段单列最清晰。
 */
function PlatformModelFields({
  provider,
  setProvider,
  type,
  setType,
  typeDisabled,
  name,
  setName,
  capabilities,
  toggleCap,
  maxContext,
  setMaxContext,
  inputPrice,
  setInputPrice,
  outputPrice,
  setOutputPrice,
  cacheInputPrice,
  setCacheInputPrice,
  markDirty,
}: FormFieldsProps) {
  return (
    <div className="mx-auto w-full max-w-3xl space-y-8">
      <div className="space-y-2.5">
        <Label className="text-sm font-medium">
          供应商 <span className="text-destructive">*</span>
        </Label>
        <Select
          value={provider}
          onValueChange={(v) => {
            setProvider(v)
            markDirty()
          }}
        >
          <SelectTrigger>
            <SelectValue placeholder="选择供应商" />
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

      <div className="space-y-2.5">
        <Label className="text-sm font-medium">
          模型标识 {!typeDisabled && <span className="text-destructive">*</span>}
        </Label>
        {setType && !typeDisabled ? (
          <Input
            placeholder="如 gpt-4o, claude-sonnet-4-20250514"
            value={type}
            onChange={(e) => {
              setType(e.target.value)
              markDirty()
            }}
          />
        ) : (
          <Input value={type} disabled />
        )}
        <p className="text-xs text-muted-foreground">模型的唯一标识符，创建后不可修改</p>
      </div>

      <div className="space-y-2.5">
        <Label className="text-sm font-medium">显示名称</Label>
        <Input
          placeholder="留空则使用模型标识"
          value={name}
          onChange={(e) => {
            setName(e.target.value)
            markDirty()
          }}
        />
      </div>

      <div className="space-y-2.5">
        <Label className="text-sm font-medium">能力标签</Label>
        <div className="flex flex-wrap gap-2.5">
          {CAPABILITIES.map((cap) => (
            <Badge
              key={cap}
              variant={capabilities.includes(cap) ? 'default' : 'outline'}
              className="cursor-pointer select-none px-3 py-1"
              onClick={() => toggleCap(cap)}
            >
              {cap}
            </Badge>
          ))}
        </div>
      </div>

      <div className="space-y-2.5">
        <Label className="text-sm font-medium">上下文窗口</Label>
        <Input
          type="number"
          placeholder="1000000"
          value={maxContext}
          onChange={(e) => {
            setMaxContext(e.target.value)
            markDirty()
          }}
        />
        <p className="text-xs text-muted-foreground">模型支持的最大上下文 token 数</p>
      </div>

      <div className="border-t border-border" />

      <div className="space-y-4">
        <div>
          <Label className="text-sm font-medium">定价</Label>
          <p className="mt-0.5 text-xs text-muted-foreground">单位：元 / 百万 tokens</p>
        </div>
        <div className="grid grid-cols-3 gap-5">
          <div className="space-y-2">
            <Label className="text-xs font-normal text-muted-foreground">输入价格</Label>
            <Input
              type="number"
              placeholder="0"
              value={inputPrice}
              onChange={(e) => {
                setInputPrice(e.target.value)
                markDirty()
              }}
            />
          </div>
          <div className="space-y-2">
            <Label className="text-xs font-normal text-muted-foreground">输出价格</Label>
            <Input
              type="number"
              placeholder="0"
              value={outputPrice}
              onChange={(e) => {
                setOutputPrice(e.target.value)
                markDirty()
              }}
            />
          </div>
          <div className="space-y-2">
            <Label className="text-xs font-normal text-muted-foreground">缓存价格</Label>
            <Input
              type="number"
              placeholder="0"
              value={cacheInputPrice}
              onChange={(e) => {
                setCacheInputPrice(e.target.value)
                markDirty()
              }}
            />
          </div>
        </div>
      </div>
    </div>
  )
}

// --- Create Workflow ---

export function PlatformModelCreateWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'platform-model-create'>) {
  const apis = useInjectedApis()
  const { closeAll } = useWorkflow()
  const onSuccess = entry.payload.onSuccess

  const [provider, setProvider] = useState('')
  const [type, setType] = useState('')
  const [name, setName] = useState('')
  const [capabilities, setCapabilities] = useState<string[]>(['chat'])
  const [maxContext, setMaxContext] = useState('1000000')
  const [inputPrice, setInputPrice] = useState('')
  const [outputPrice, setOutputPrice] = useState('')
  const [cacheInputPrice, setCacheInputPrice] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const canSubmit = provider && type.trim()
  const markDirty = () => onSetDirty(true)

  const toggleCap = (cap: string) => {
    setCapabilities((prev) => (prev.includes(cap) ? prev.filter((c) => c !== cap) : [...prev, cap]))
    markDirty()
  }

  const handleSubmit = async () => {
    if (!canSubmit) return
    setSubmitting(true)
    try {
      await apis.platformApi.createModel({
        type: type.trim(),
        name: name.trim() || type.trim(),
        provider,
        inputPrice: Number(inputPrice) || 0,
        outputPrice: Number(outputPrice) || 0,
        cacheInputPrice: Number(cacheInputPrice) || 0,
        capabilities,
        maxContext: Number(maxContext) || 1000000,
      })
      toast.success('模型已添加')
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
      title="添加平台模型"
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={submitting ? '添加中...' : '添加'}
          onPrimary={handleSubmit}
          primaryDisabled={!canSubmit || submitting}
        />
      }
    >
      <WorkflowFormLayout variant="full">
        <PlatformModelFields
          provider={provider}
          setProvider={setProvider}
          type={type}
          setType={setType}
          name={name}
          setName={setName}
          capabilities={capabilities}
          toggleCap={toggleCap}
          maxContext={maxContext}
          setMaxContext={setMaxContext}
          inputPrice={inputPrice}
          setInputPrice={setInputPrice}
          outputPrice={outputPrice}
          setOutputPrice={setOutputPrice}
          cacheInputPrice={cacheInputPrice}
          setCacheInputPrice={setCacheInputPrice}
          markDirty={markDirty}
        />
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}

// --- Edit Workflow ---

export function PlatformModelEditWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'platform-model-edit'>) {
  const apis = useInjectedApis()
  const { closeAll } = useWorkflow()
  const model = entry.payload.model
  const onSuccess = entry.payload.onSuccess

  const [provider, setProvider] = useState(model.provider)
  const [name, setName] = useState(model.name)
  const [capabilities, setCapabilities] = useState<string[]>(
    model.capabilities?.length ? model.capabilities : ['chat'],
  )
  const [maxContext, setMaxContext] = useState(String(model.maxContext || 1000000))
  const [inputPrice, setInputPrice] = useState(model.inputPrice > 0 ? String(model.inputPrice) : '')
  const [outputPrice, setOutputPrice] = useState(
    model.outputPrice > 0 ? String(model.outputPrice) : '',
  )
  const [cacheInputPrice, setCacheInputPrice] = useState(
    model.cacheInputPrice > 0 ? String(model.cacheInputPrice) : '',
  )
  const [submitting, setSubmitting] = useState(false)

  const canSubmit = provider && name.trim()
  const markDirty = () => onSetDirty(true)

  const toggleCap = (cap: string) => {
    setCapabilities((prev) => (prev.includes(cap) ? prev.filter((c) => c !== cap) : [...prev, cap]))
    markDirty()
  }

  const handleSubmit = async () => {
    if (!canSubmit) return
    setSubmitting(true)
    try {
      await apis.platformApi.updateModel(model.modelId, {
        name: name.trim(),
        provider,
        capabilities,
        maxContext: Number(maxContext) || 1000000,
      })
      const newInput = Number(inputPrice) || 0
      const newOutput = Number(outputPrice) || 0
      const newCache = Number(cacheInputPrice) || 0
      if (
        newInput !== model.inputPrice ||
        newOutput !== model.outputPrice ||
        newCache !== model.cacheInputPrice
      ) {
        await apis.platformApi.setPricing(model.modelId, {
          inputPrice: newInput,
          outputPrice: newOutput,
          cacheInputPrice: newCache,
        })
      }
      toast.success('模型已更新')
      onSuccess?.()
      closeAll()
    } catch (err) {
      toast.error(workflowErrorMessage(err, '更新失败'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title={`编辑 ${model.name}`}
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
      <WorkflowFormLayout variant="full">
        <PlatformModelFields
          provider={provider}
          setProvider={setProvider}
          type={model.type}
          typeDisabled
          name={name}
          setName={setName}
          capabilities={capabilities}
          toggleCap={toggleCap}
          maxContext={maxContext}
          setMaxContext={setMaxContext}
          inputPrice={inputPrice}
          setInputPrice={setInputPrice}
          outputPrice={outputPrice}
          setOutputPrice={setOutputPrice}
          cacheInputPrice={cacheInputPrice}
          setCacheInputPrice={setCacheInputPrice}
          markDirty={markDirty}
        />
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
