import { useState, useEffect } from 'react'
import { Eye, EyeOff } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { DepartmentTreeSelect } from '@/components/ui/department-tree-select'
import { modelApi } from '@/api/models'
import { budgetApi } from '@/api/budget'
import type { ModelInfo, ProviderType, AuthMethod, Visibility, BudgetNode } from '@/api/types'
import { X } from 'lucide-react'

interface ModelFormProps {
  model: ModelInfo | null
  onCancel: () => void
  onSaved: () => void
}

function findNodeName(nodes: BudgetNode[], id: string): string | undefined {
  for (const n of nodes) {
    if (n.id === id) return n.name
    if (n.children) {
      const found = findNodeName(n.children, id)
      if (found) return found
    }
  }
  return undefined
}

export function ModelForm({ model, onCancel, onSaved }: ModelFormProps) {
  const isBuiltin = model?.type === 'builtin'

  const [form, setForm] = useState({
    displayName: model?.displayName ?? '',
    name: model?.name ?? '',
    provider: (model?.provider ?? 'openai') as ProviderType,
    description: model?.description ?? '',
    authMethod: (model?.authMethod ?? 'api_key') as AuthMethod,
    apiKey: model?.apiKey ?? '',
    endpoint: model?.endpoint ?? '',
    proxyUrl: model?.proxyUrl ?? '',
    supportsImage: model?.capabilities?.includes('vision') ? 'true' : 'false',
    maxContext: model?.maxContext ?? 128000,
    maxOutput: model?.maxOutput ?? 4096,
    visibility: (model?.visibility ?? 'all') as Visibility,
  })

  const [visibleDepts, setVisibleDepts] = useState<{ id: string; name: string }[]>(
    model?.visibleDepartmentIds?.map(id => ({ id, name: id })) ?? []
  )
  const [tree, setTree] = useState<BudgetNode[]>([])
  const [showKey, setShowKey] = useState(false)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    budgetApi.getTree().then((data) => {
      setTree(data)
      // Resolve department names from tree
      if (model?.visibleDepartmentIds?.length) {
        const resolved = model.visibleDepartmentIds.map(id => {
          const name = findNodeName(data, id)
          return { id, name: name ?? id }
        })
        setVisibleDepts(resolved)
      }
    })
  }, [])

  const setField = <K extends keyof typeof form>(key: K, value: typeof form[K]) => {
    setForm(prev => ({ ...prev, [key]: value }))
  }

  const handleSubmit = async () => {
    if (!form.displayName) return
    if (!isBuiltin && !form.name) return

    setSaving(true)
    try {
      const capabilities: string[] = ['chat']
      if (form.supportsImage === 'true') capabilities.push('vision')

      const payload: Partial<ModelInfo> = {
        displayName: form.displayName,
        capabilities,
        maxContext: Number(form.maxContext),
        maxOutput: Number(form.maxOutput),
        visibility: form.visibility,
        visibleDepartmentIds: form.visibility === 'partial' ? visibleDepts.map(d => d.id) : [],
      }

      if (!isBuiltin) {
        Object.assign(payload, {
          name: form.name,
          provider: form.provider,
          description: form.description,
          authMethod: form.authMethod,
          apiKey: form.apiKey,
          endpoint: form.endpoint,
          proxyUrl: form.proxyUrl,
          type: 'custom' as const,
          inputPrice: model?.inputPrice ?? 0,
          outputPrice: model?.outputPrice ?? 0,
          enabled: model?.enabled ?? true,
        })
      }

      if (model) {
        await modelApi.update(model.id, payload)
      } else {
        await modelApi.create({
          ...(payload as Omit<ModelInfo, 'id'>),
          name: form.name,
          provider: form.provider,
          description: form.description,
          authMethod: form.authMethod,
          apiKey: form.apiKey,
          endpoint: form.endpoint,
          proxyUrl: form.proxyUrl,
          type: 'custom',
          inputPrice: 0,
          outputPrice: 0,
          enabled: true,
          visibility: form.visibility,
        })
      }
      onSaved()
    } catch (err) {
      console.error(err)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-8">
      {/* Section 1: 基本信息 — custom only */}
      {!isBuiltin && (
        <div className="flex gap-4">
          <div className="w-1 shrink-0 rounded-full bg-primary" />
          <div className="flex-1 space-y-4">
            <h4 className="text-sm font-semibold">基本信息</h4>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>模型名称</Label>
                <Input
                  value={form.displayName}
                  onChange={e => setField('displayName', e.target.value)}
                  maxLength={20}
                  placeholder="输入模型展示名称"
                />
              </div>
              <div className="space-y-1.5">
                <Label>模型 ID</Label>
                <Input
                  value={form.name}
                  onChange={e => setField('name', e.target.value)}
                  maxLength={200}
                  placeholder="用于请求模型入参"
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <Label>模型提供商</Label>
              <Select value={form.provider} onValueChange={v => setField('provider', v as ProviderType)}>
                <SelectTrigger className="w-48">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="openai">OpenAI</SelectItem>
                  <SelectItem value="anthropic">Anthropic</SelectItem>
                  <SelectItem value="deepseek">DeepSeek</SelectItem>
                  <SelectItem value="qwen">通义千问</SelectItem>
                  <SelectItem value="zhipu">智谱 AI</SelectItem>
                  <SelectItem value="baichuan">百川智能</SelectItem>
                  <SelectItem value="minimax">MiniMax</SelectItem>
                  <SelectItem value="moonshot">Moonshot</SelectItem>
                  <SelectItem value="stepfun">阶跃星辰</SelectItem>
                  <SelectItem value="baidu">文心一言</SelectItem>
                  <SelectItem value="hunyuan">腾讯混元</SelectItem>
                  <SelectItem value="doubao">豆包</SelectItem>
                  <SelectItem value="sensetime">商汤日日新</SelectItem>
                  <SelectItem value="custom">自定义</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1.5">
              <Label>模型描述</Label>
              <textarea
                className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary resize-none"
                rows={3}
                value={form.description}
                onChange={e => setField('description', e.target.value)}
                maxLength={500}
                placeholder="描述模型的特点和适用场景"
              />
              <p className="text-xs text-muted-foreground text-right">{form.description.length}/500</p>
            </div>
          </div>
        </div>
      )}

      {/* Section 2: 接入配置 — custom only */}
      {!isBuiltin && (
        <div className="flex gap-4">
          <div className="w-1 shrink-0 rounded-full bg-primary" />
          <div className="flex-1 space-y-4">
            <h4 className="text-sm font-semibold">接入配置</h4>

            <div className="space-y-1.5">
              <Label>鉴权方式</Label>
              <RadioGroup
                value={form.authMethod}
                onValueChange={v => setField('authMethod', v as AuthMethod)}
                className="flex gap-6"
              >
                <label className="flex items-center gap-2 cursor-pointer">
                  <RadioGroupItem value="api_key" />
                  <span className="text-sm">API Key</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                  <RadioGroupItem value="ak_sk" />
                  <span className="text-sm">AK/SK</span>
                </label>
              </RadioGroup>
            </div>

            <div className="space-y-1.5">
              <Label>API_KEY</Label>
              <div className="relative">
                <Input
                  type={showKey ? 'text' : 'password'}
                  value={form.apiKey}
                  onChange={e => setField('apiKey', e.target.value)}
                  className="pr-10"
                  placeholder="sk-..."
                />
                <button
                  type="button"
                  aria-label={showKey ? '隐藏密钥' : '显示密钥'}
                  onClick={() => setShowKey(s => !s)}
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showKey
                    ? <EyeOff className="h-4 w-4" strokeWidth={1.5} />
                    : <Eye className="h-4 w-4" strokeWidth={1.5} />
                  }
                </button>
              </div>
            </div>

            <div className="space-y-1.5">
              <Label>模型部署地址</Label>
              <Input
                value={form.endpoint}
                onChange={e => setField('endpoint', e.target.value)}
                placeholder="https://api.example.com/v1"
              />
              <p className="text-xs text-muted-foreground">
                如果没有合适的模型，建议去模型服务商选购或者部署模型，获取地址和Key
              </p>
            </div>

            <div className="space-y-1.5">
              <Label>网络代理地址</Label>
              <Input
                value={form.proxyUrl}
                onChange={e => setField('proxyUrl', e.target.value)}
                placeholder="输入完整的代理地址，为空则不使用代理"
              />
            </div>
          </div>
        </div>
      )}

      {/* Section 3: 模型能力配置 */}
      <div className="flex gap-4">
        <div className="w-1 shrink-0 rounded-full bg-primary" />
        <div className="flex-1 space-y-4">
          <h4 className="text-sm font-semibold">模型能力配置</h4>

          <div className="space-y-1.5">
            <Label>是否支持图片</Label>
            <RadioGroup
              value={form.supportsImage}
              onValueChange={v => setField('supportsImage', v)}
              className="flex gap-6"
            >
              <label className="flex items-center gap-2 cursor-pointer">
                <RadioGroupItem value="true" />
                <span className="text-sm">支持</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <RadioGroupItem value="false" />
                <span className="text-sm">不支持</span>
              </label>
            </RadioGroup>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>模型最大输入长度</Label>
              <Input
                type="number"
                value={form.maxContext}
                onChange={e => setField('maxContext', Number(e.target.value))}
                placeholder="范围: 64000-1000000"
              />
            </div>
            <div className="space-y-1.5">
              <Label>模型最大输出长度</Label>
              <Input
                type="number"
                value={form.maxOutput}
                onChange={e => setField('maxOutput', Number(e.target.value))}
                placeholder="范围: 512-1000000"
              />
            </div>
          </div>
        </div>
      </div>

      {/* Section 4: 权限配置 */}
      <div className="flex gap-4">
        <div className="w-1 shrink-0 rounded-full bg-primary" />
        <div className="flex-1 space-y-4">
          <h4 className="text-sm font-semibold">权限配置</h4>

          <div className="space-y-1.5">
            <Label>可见范围</Label>
            <RadioGroup
              value={form.visibility}
              onValueChange={v => setField('visibility', v as Visibility)}
              className="flex gap-6"
            >
              <label className="flex items-center gap-2 cursor-pointer">
                <RadioGroupItem value="all" />
                <span className="text-sm">所有成员</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <RadioGroupItem value="partial" />
                <span className="text-sm">部分成员</span>
              </label>
            </RadioGroup>
          </div>

          {form.visibility === 'partial' && (
            <div className="space-y-2">
              <DepartmentTreeSelect
                tree={tree}
                value=""
                onChange={(id, name) => {
                  if (!visibleDepts.some(d => d.id === id)) {
                    setVisibleDepts([...visibleDepts, { id, name }])
                  }
                }}
                placeholder="选择可见团队…"
              />
              {visibleDepts.length > 0 && (
                <div className="flex flex-wrap gap-1.5">
                  {visibleDepts.map(dept => (
                    <Badge key={dept.id} variant="outline" className="gap-1">
                      {dept.name}
                      <button type="button" onClick={() => setVisibleDepts(visibleDepts.filter(d => d.id !== dept.id))} aria-label={`移除 ${dept.name}`}>
                        <X className="size-3" />
                      </button>
                    </Badge>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Footer */}
      <div className="flex justify-end gap-3 pt-4 border-t border-border">
        <Button variant="ghost" onClick={onCancel}>取消</Button>
        <Button onClick={handleSubmit} disabled={saving}>
          {saving ? '保存中...' : '保存'}
        </Button>
      </div>
    </div>
  )
}
