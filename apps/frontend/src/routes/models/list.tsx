import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { ModelForm } from '@/components/models/model-form'
import { modelApi } from '@/api/models'
import type { ModelInfo, ProviderType } from '@/api/types'

const providerLabels: Record<ProviderType, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  deepseek: 'DeepSeek',
  qwen: '通义千问',
  zhipu: '智谱 AI',
  baichuan: '百川智能',
  minimax: 'MiniMax',
  moonshot: 'Moonshot',
  stepfun: '阶跃星辰',
  baidu: '文心一言',
  hunyuan: '腾讯混元',
  doubao: '豆包',
  sensetime: '商汤日日新',
  custom: '自定义',
}

type View = 'list' | 'form'

function ModelTable({
  models,
  onToggle,
  onEdit,
  onDelete,
  showActions = true,
}: {
  models: ModelInfo[]
  onToggle: (model: ModelInfo) => void
  onEdit: (model: ModelInfo) => void
  onDelete: (model: ModelInfo) => void
  showActions?: boolean
}) {
  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">模型名称</TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">模型 ID</TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">模型类型</TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">模型供应商</TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">模型描述</TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">可见范围</TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">模型部署地址</TableHead>
          <TableHead className="text-xs font-medium uppercase text-muted-foreground">启用</TableHead>
          {showActions && <TableHead className="text-xs font-medium uppercase text-muted-foreground">操作</TableHead>}
        </TableRow>
      </TableHeader>
      <TableBody>
        {models.map(model => (
          <TableRow key={model.id} className="even:bg-muted/40 hover:bg-muted/50">
            <TableCell className="font-medium text-sm">{model.displayName}</TableCell>
            <TableCell>
              <span className="text-xs font-mono text-muted-foreground">{model.name}</span>
            </TableCell>
            <TableCell>
              {model.type === 'builtin'
                ? <Badge variant="secondary">内置模型</Badge>
                : <Badge variant="outline">自定义</Badge>
              }
            </TableCell>
            <TableCell className="text-sm">{model.type === 'custom' ? providerLabels[model.provider] : '—'}</TableCell>
            <TableCell>
              <span className="line-clamp-1 max-w-48 block text-sm text-muted-foreground">
                {model.description || '—'}
              </span>
            </TableCell>
            <TableCell className="text-sm">
              {model.visibility === 'all' ? '全部用户' : '部分成员'}
            </TableCell>
            <TableCell>
              {model.type === 'builtin'
                ? <span className="text-sm text-muted-foreground">—</span>
                : <span className="max-w-36 truncate block text-xs text-muted-foreground">
                    {model.endpoint || '—'}
                  </span>
              }
            </TableCell>
            <TableCell>
              <Switch checked={model.enabled} onCheckedChange={() => onToggle(model)} />
            </TableCell>
            {showActions && (
            <TableCell>
              <div className="flex items-center gap-1">
                {model.type === 'custom' && (
                  <Button variant="ghost" size="sm" onClick={() => onEdit(model)}>
                    编辑
                  </Button>
                )}
                {model.type === 'custom' && (
                  <AlertDialog>
                    <AlertDialogTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-destructive hover:text-destructive"
                      >
                        删除
                      </Button>
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>删除模型</AlertDialogTitle>
                        <AlertDialogDescription>
                          确认删除模型「{model.displayName}」？此操作不可撤销。
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>取消</AlertDialogCancel>
                        <AlertDialogAction
                          variant="destructive"
                          onClick={() => onDelete(model)}
                        >
                          删除
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                )}
              </div>
            </TableCell>
            )}
          </TableRow>
        ))}
        {models.length === 0 && (
          <TableRow>
            <TableCell colSpan={9} className="py-12 text-center text-sm text-muted-foreground">
              暂无数据
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  )
}

export default function ModelListPage() {
  const [models, setModels] = useState<ModelInfo[]>([])
  const [view, setView] = useState<View>('list')
  const [editingModel, setEditingModel] = useState<ModelInfo | null>(null)

  const loadModels = () => {
    modelApi.list().then(setModels).catch(console.error)
  }

  useEffect(() => {
    loadModels()
  }, [])

  const handleToggle = async (model: ModelInfo) => {
    await modelApi.toggle(model.id, !model.enabled)
    setModels(prev => prev.map(m => m.id === model.id ? { ...m, enabled: !m.enabled } : m))
  }

  const handleEdit = (model: ModelInfo) => {
    setEditingModel(model)
    setView('form')
  }

  const handleDelete = async (model: ModelInfo) => {
    try {
      await modelApi.delete(model.id)
      setModels(prev => prev.filter(m => m.id !== model.id))
    } catch (err) {
      console.error(err)
    }
  }

  const handleSaved = () => {
    loadModels()
    setView('list')
    setEditingModel(null)
  }

  const allCount = models.length
  const customCount = models.filter(m => m.type === 'custom').length
  const builtinCount = models.filter(m => m.type === 'builtin').length

  if (view === 'form') {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <button
            onClick={() => { setView('list'); setEditingModel(null) }}
            className="hover:text-foreground transition-colors"
          >
            模型管理
          </button>
          <span>/</span>
          <span className="text-foreground">
            {editingModel ? '编辑模型' : '新增自定义模型'}
          </span>
        </div>

        <div className="rounded-lg border border-border shadow-xs bg-card p-6 space-y-8">
          <ModelForm
            model={editingModel}
            onCancel={() => { setView('list'); setEditingModel(null) }}
            onSaved={handleSaved}
          />
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-border shadow-xs">
        <Tabs defaultValue="all">
          <div className="flex items-center justify-between px-5 pt-4 pb-0 border-b border-border">
            <TabsList variant="line">
              <TabsTrigger value="all">全部模型 ({allCount})</TabsTrigger>
              <TabsTrigger value="custom">自定义模型 ({customCount})</TabsTrigger>
              <TabsTrigger value="builtin">内置模型 ({builtinCount})</TabsTrigger>
            </TabsList>
            <div className="pb-3">
              <Button size="sm" onClick={() => { setEditingModel(null); setView('form') }}>
                + 新增自定义模型
              </Button>
            </div>
          </div>

          <TabsContent value="all">
            <ModelTable
              models={models}
              onToggle={handleToggle}
              onEdit={handleEdit}
              onDelete={handleDelete}
            />
          </TabsContent>
          <TabsContent value="custom">
            <ModelTable
              models={models.filter(m => m.type === 'custom')}
              onToggle={handleToggle}
              onEdit={handleEdit}
              onDelete={handleDelete}
              showActions={false}
            />
          </TabsContent>
          <TabsContent value="builtin">
            <ModelTable
              models={models.filter(m => m.type === 'builtin')}
              onToggle={handleToggle}
              onEdit={handleEdit}
              onDelete={handleDelete}
            />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
