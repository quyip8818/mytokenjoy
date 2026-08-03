import { useState, useEffect } from 'react'
import { FormDialog } from '@/components/ui/form-dialog'
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
import type { PlatformModel } from '@/api/types'

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

export interface ModelFormData {
  provider: string
  type: string
  name: string
  capabilities: string[]
  maxContext: number
  inputPrice: number
  outputPrice: number
  cacheInputPrice: number
}

interface ModelFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  mode: 'create' | 'edit'
  initialData?: PlatformModel | null
  busy?: boolean
  error?: string | null
  onSubmit: (data: ModelFormData) => void | Promise<void>
}

export function ModelFormDialog({
  open,
  onOpenChange,
  mode,
  initialData,
  busy = false,
  error,
  onSubmit,
}: ModelFormDialogProps) {
  const [provider, setProvider] = useState('')
  const [type, setType] = useState('')
  const [name, setName] = useState('')
  const [capabilities, setCapabilities] = useState<string[]>(['chat'])
  const [maxContext, setMaxContext] = useState('1000000')
  const [inputPrice, setInputPrice] = useState('')
  const [outputPrice, setOutputPrice] = useState('')
  const [cacheInputPrice, setCacheInputPrice] = useState('')

  useEffect(() => {
    if (open && mode === 'edit' && initialData) {
      setProvider(initialData.provider)
      setType(initialData.type)
      setName(initialData.name)
      setCapabilities(initialData.capabilities?.length ? initialData.capabilities : ['chat'])
      setMaxContext(String(initialData.maxContext || 1000000))
      setInputPrice(initialData.inputPrice > 0 ? String(initialData.inputPrice) : '')
      setOutputPrice(initialData.outputPrice > 0 ? String(initialData.outputPrice) : '')
      setCacheInputPrice(initialData.cacheInputPrice > 0 ? String(initialData.cacheInputPrice) : '')
    } else if (open && mode === 'create') {
      setProvider('')
      setType('')
      setName('')
      setCapabilities(['chat'])
      setMaxContext('1000000')
      setInputPrice('')
      setOutputPrice('')
      setCacheInputPrice('')
    }
  }, [open, mode, initialData])

  const canSubmit = provider && type.trim()

  const handleSubmit = () => {
    if (!canSubmit) return
    return onSubmit({
      provider,
      type: type.trim(),
      name: name.trim() || type.trim(),
      capabilities,
      maxContext: Number(maxContext) || 1000000,
      inputPrice: Number(inputPrice) || 0,
      outputPrice: Number(outputPrice) || 0,
      cacheInputPrice: Number(cacheInputPrice) || 0,
    })
  }

  const toggleCap = (cap: string) => {
    setCapabilities((prev) =>
      prev.includes(cap) ? prev.filter((c) => c !== cap) : [...prev, cap],
    )
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={mode === 'create' ? '添加模型' : '编辑模型'}
      description={mode === 'create' ? '将模型添加到平台目录' : `编辑 ${initialData?.name ?? ''}`}
      submitLabel={mode === 'create' ? '添加' : '保存'}
      submitDisabled={!canSubmit}
      busy={busy}
      error={error}
      onSubmit={handleSubmit}
      className="sm:max-w-lg"
    >
      {/* Provider */}
      <div className="space-y-1.5">
        <Label>供应商 <span className="text-destructive">*</span></Label>
        <Select value={provider} onValueChange={setProvider}>
          <SelectTrigger className="w-full">
            <SelectValue placeholder="选择供应商" />
          </SelectTrigger>
          <SelectContent>
            {PROVIDERS.map((p) => (
              <SelectItem key={p.value} value={p.value}>{p.label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Model ID */}
      <div className="space-y-1.5">
        <Label>模型标识 <span className="text-destructive">*</span></Label>
        <Input
          placeholder="如 gpt-4o, claude-sonnet-4-20250514"
          value={type}
          onChange={(e) => setType(e.target.value)}
          disabled={mode === 'edit'}
        />
      </div>

      {/* Display Name */}
      <div className="space-y-1.5">
        <Label>显示名称</Label>
        <Input
          placeholder="留空则使用模型标识"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
      </div>

      {/* Capabilities */}
      <div className="space-y-1.5">
        <Label>能力标签</Label>
        <div className="flex flex-wrap gap-1.5">
          {CAPABILITIES.map((cap) => (
            <Badge
              key={cap}
              variant={capabilities.includes(cap) ? 'default' : 'outline'}
              className="cursor-pointer select-none"
              onClick={() => toggleCap(cap)}
            >
              {cap}
            </Badge>
          ))}
        </div>
      </div>

      {/* Max Context */}
      <div className="space-y-1.5">
        <Label>上下文窗口</Label>
        <Input
          type="number"
          placeholder="1000000"
          value={maxContext}
          onChange={(e) => setMaxContext(e.target.value)}
        />
      </div>

      {/* Pricing */}
      <div className="grid grid-cols-3 gap-3">
        <div className="space-y-1.5">
          <Label>输入价格 (元/百万tokens)</Label>
          <Input
            type="number"
            placeholder="0"
            value={inputPrice}
            onChange={(e) => setInputPrice(e.target.value)}
          />
        </div>
        <div className="space-y-1.5">
          <Label>输出价格 (元/百万tokens)</Label>
          <Input
            type="number"
            placeholder="0"
            value={outputPrice}
            onChange={(e) => setOutputPrice(e.target.value)}
          />
        </div>
        <div className="space-y-1.5">
          <Label>缓存价格 (元/百万tokens)</Label>
          <Input
            type="number"
            placeholder="0"
            value={cacheInputPrice}
            onChange={(e) => setCacheInputPrice(e.target.value)}
          />
        </div>
      </div>
    </FormDialog>
  )
}
