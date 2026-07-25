import { useState } from 'react'
import { Plus, Pencil, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useFilteredQuery, useInjectedMutation, queryKeys } from '@/features/query'
import { useSession } from '@/features/session'
import { useSupplierOptions } from '@/features/suppliers'
import { useApis } from '@/api/use-apis'
import { MODEL_STATUS, MODEL_TYPES } from '@/config/enums'
import { Badge, Field } from '@/components/ui'
import type { AiModel } from '@/api/models'

export function ModelsPage() {
  const { user } = useSession()
  const canEdit = user?.role === 'admin' || user?.role === 'buyer'
  const apis = useApis()
  const suppliers = useSupplierOptions()

  const { data, loading, filter, search } = useFilteredQuery({
    initialFilter: { page: 1, pageSize: 12, keyword: '', supplierId: 0, modelType: '', status: '' },
    queryKeyFactory: (f) => queryKeys.models.list(f),
    fetcher: (a, f) => a.modelsApi.list(f),
  })

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<AiModel | null>(null)
  const [form, setForm] = useState({
    supplierId: 0,
    modelName: '',
    modelId: '',
    modelType: '文本',
    contextLength: '',
    inputPrice: '',
    outputPrice: '',
    discount: '',
    status: 'available',
    description: '',
  })
  const [saving, setSaving] = useState(false)

  const deleteMut = useInjectedMutation<void, number>({
    mutationFn: (a, id) => a.modelsApi.delete(id),
    invalidateKeys: [queryKeys.models.all],
    onSuccess: () => toast.success('删除成功'),
  })

  const openCreate = () => {
    setEditing(null)
    setForm({
      supplierId: 0,
      modelName: '',
      modelId: '',
      modelType: '文本',
      contextLength: '',
      inputPrice: '',
      outputPrice: '',
      discount: '',
      status: 'available',
      description: '',
    })
    setDialogOpen(true)
  }

  const openEdit = (m: AiModel) => {
    setEditing(m)
    setForm({
      supplierId: m.supplierId,
      modelName: m.modelName,
      modelId: m.modelId ?? '',
      modelType: m.modelType ?? '文本',
      contextLength: m.contextLength?.toString() ?? '',
      inputPrice: m.inputPrice?.toString() ?? '',
      outputPrice: m.outputPrice?.toString() ?? '',
      discount: m.discount?.toString() ?? '',
      status: m.status,
      description: m.description ?? '',
    })
    setDialogOpen(true)
  }

  const handleSave = async () => {
    if (!form.modelName || !form.supplierId) {
      toast.error('模型名称和供应商不能为空')
      return
    }
    setSaving(true)
    try {
      const payload = {
        supplierId: form.supplierId,
        modelName: form.modelName,
        modelId: form.modelId || undefined,
        modelType: form.modelType || undefined,
        contextLength: form.contextLength ? Number(form.contextLength) : undefined,
        inputPrice: form.inputPrice ? Number(form.inputPrice) : undefined,
        outputPrice: form.outputPrice ? Number(form.outputPrice) : undefined,
        discount: form.discount ? Number(form.discount) : undefined,
        status: form.status,
        description: form.description || undefined,
      }
      if (editing) {
        await apis.modelsApi.update(editing.id, payload)
        toast.success('更新成功')
      } else {
        await apis.modelsApi.create(payload)
        toast.success('创建成功')
      }
      setDialogOpen(false)
      search({})
    } catch (e: any) {
      toast.error(e.message)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = (m: AiModel) => {
    if (!confirm(`确定删除模型「${m.modelName}」吗？`)) return
    deleteMut.mutate(m.id)
  }

  const typeFilters = ['全部', ...MODEL_TYPES]

  return (
    <div className="space-y-4">
      {/* Hero */}
      <div className="rounded-lg border bg-gradient-to-r from-blue-50 to-purple-50 p-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold">AI 模型目录</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              覆盖 {suppliers.length} 家供应商 · 共 {data?.total ?? 0} 个模型
            </p>
          </div>
          {canEdit && (
            <button
              onClick={openCreate}
              className="inline-flex items-center gap-1.5 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"
            >
              <Plus className="h-4 w-4" /> 新建模型
            </button>
          )}
        </div>
      </div>

      <div className="space-y-3 rounded-lg border bg-white p-4">
        <div className="flex items-center gap-2">
          <span className="w-14 text-xs font-medium text-muted-foreground">类型</span>
          <div className="flex flex-wrap gap-1.5">
            {typeFilters.map((t) => (
              <button
                key={t}
                onClick={() => search({ modelType: t === '全部' ? '' : t })}
                className={`rounded-full px-3 py-1 text-xs font-medium transition ${(t === '全部' && !filter.modelType) || filter.modelType === t ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/80'}`}
              >
                {t}
              </button>
            ))}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span className="w-14 text-xs font-medium text-muted-foreground">供应商</span>
          <div className="flex flex-wrap gap-1.5">
            <button
              onClick={() => search({ supplierId: 0 })}
              className={`rounded-full px-3 py-1 text-xs font-medium transition ${!filter.supplierId ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/80'}`}
            >
              全部
            </button>
            {suppliers.map((s) => (
              <button
                key={s.id}
                onClick={() => search({ supplierId: s.id })}
                className={`rounded-full px-3 py-1 text-xs font-medium transition ${filter.supplierId === s.id ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/80'}`}
              >
                {s.name}
              </button>
            ))}
          </div>
          <input
            className="ml-auto h-8 w-44 rounded-md border px-3 text-sm"
            placeholder="搜索模型名称"
            value={filter.keyword}
            onChange={(e) => search({ keyword: e.target.value })}
          />
        </div>
      </div>

      {/* 卡片网格 */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {data?.items.map((m) => (
          <div
            key={m.id}
            className="group rounded-lg border bg-white p-4 transition-shadow hover:shadow-md"
          >
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-2">
                <span className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-100 text-xs font-bold text-blue-700">
                  {m.supplierName?.slice(0, 1) ?? '?'}
                </span>
                <div>
                  <div className="text-sm font-medium">{m.modelName}</div>
                  <div className="text-xs text-muted-foreground">{m.supplierName}</div>
                </div>
              </div>
              <Badge variant={m.status === 'available' ? 'success' : 'outline'}>
                {MODEL_STATUS[m.status]?.label ?? m.status}
              </Badge>
            </div>
            <div className="mt-3 grid grid-cols-2 gap-2 text-xs text-muted-foreground">
              <div>类型：{m.modelType ?? '-'}</div>
              <div>上下文：{m.contextLength ? `${Math.round(m.contextLength / 1000)}K` : '-'}</div>
              <div>输入价：{m.inputPrice ?? '-'}</div>
              <div>输出价：{m.outputPrice ?? '-'}</div>
            </div>
            {canEdit && (
              <div className="mt-3 flex justify-end gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                <button
                  onClick={() => openEdit(m)}
                  className="rounded p-1 text-muted-foreground hover:bg-muted"
                >
                  <Pencil className="h-3.5 w-3.5" />
                </button>
                <button
                  onClick={() => handleDelete(m)}
                  className="rounded p-1 text-muted-foreground hover:bg-red-50 hover:text-red-500"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              </div>
            )}
          </div>
        ))}
        {!loading && data?.items.length === 0 && (
          <div className="col-span-full py-16 text-center text-muted-foreground">
            暂无符合条件的模型
          </div>
        )}
      </div>

      {/* 弹窗 */}
      {dialogOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          onClick={() => setDialogOpen(false)}
        >
          <div
            className="w-full max-w-lg rounded-lg bg-white p-6 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="mb-4 text-lg font-semibold">{editing ? '编辑模型' : '新建模型'}</h2>
            <div className="space-y-3">
              <Field label="供应商" required>
                <select
                  className="input"
                  value={form.supplierId}
                  onChange={(e) => setForm({ ...form, supplierId: Number(e.target.value) })}
                  disabled={!!editing}
                >
                  <option value={0}>请选择</option>
                  {suppliers.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </select>
              </Field>
              <Field label="模型名称" required>
                <input
                  className="input"
                  value={form.modelName}
                  onChange={(e) => setForm({ ...form, modelName: e.target.value })}
                />
              </Field>
              <Field label="模型标识符 (model_id)" hint="NewAPI 使用的标准标识符，留空则不同步">
                <input
                  className="input font-mono text-sm"
                  value={form.modelId}
                  onChange={(e) => setForm({ ...form, modelId: e.target.value })}
                  placeholder="如 gpt-4o-2024-11-20"
                />
              </Field>
              <div className="grid grid-cols-2 gap-3">
                <Field label="模型类型">
                  <select
                    className="input"
                    value={form.modelType}
                    onChange={(e) => setForm({ ...form, modelType: e.target.value })}
                  >
                    {MODEL_TYPES.map((t) => (
                      <option key={t} value={t}>
                        {t}
                      </option>
                    ))}
                  </select>
                </Field>
                <Field label="上下文长度">
                  <input
                    className="input"
                    type="number"
                    value={form.contextLength}
                    onChange={(e) => setForm({ ...form, contextLength: e.target.value })}
                    placeholder="如 128000"
                  />
                </Field>
              </div>
              <div className="grid grid-cols-3 gap-3">
                <Field label="输入价">
                  <input
                    className="input"
                    type="number"
                    step="0.01"
                    value={form.inputPrice}
                    onChange={(e) => setForm({ ...form, inputPrice: e.target.value })}
                  />
                </Field>
                <Field label="输出价">
                  <input
                    className="input"
                    type="number"
                    step="0.01"
                    value={form.outputPrice}
                    onChange={(e) => setForm({ ...form, outputPrice: e.target.value })}
                  />
                </Field>
                <Field label="折扣(%)">
                  <input
                    className="input"
                    type="number"
                    value={form.discount}
                    onChange={(e) => setForm({ ...form, discount: e.target.value })}
                  />
                </Field>
              </div>
              <Field label="状态">
                <select
                  className="input"
                  value={form.status}
                  onChange={(e) => setForm({ ...form, status: e.target.value })}
                >
                  {Object.entries(MODEL_STATUS).map(([k, v]) => (
                    <option key={k} value={k}>
                      {v.label}
                    </option>
                  ))}
                </select>
              </Field>
              <Field label="描述">
                <textarea
                  className="input min-h-[60px] resize-none"
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                />
              </Field>
            </div>
            <div className="mt-6 flex justify-end gap-2">
              <button
                onClick={() => setDialogOpen(false)}
                className="rounded-md border px-4 py-2 text-sm"
              >
                取消
              </button>
              <button
                onClick={handleSave}
                disabled={saving}
                className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
              >
                {saving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
