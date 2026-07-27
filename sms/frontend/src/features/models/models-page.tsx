import { useState } from 'react'
import { RefreshCw, Pencil, Trash2, Power } from 'lucide-react'
import { toast } from 'sonner'
import { useFilteredQuery, queryKeys } from '@/features/query'
import { useSession } from '@/features/session'
import { useApis } from '@/api/use-apis'
import { MODEL_STATUS, MODEL_TYPES } from '@/config/enums'
import { Badge, Field } from '@/components/ui'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import type { AiModel } from '@/api/models'

function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel,
  variant,
  onConfirm,
  onCancel,
}: {
  open: boolean
  title: string
  message: string
  confirmLabel: string
  variant?: 'danger' | 'warning'
  onConfirm: () => void
  onCancel: () => void
}) {
  if (!open) return null
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      onClick={onCancel}
    >
      <div
        className="w-full max-w-sm rounded-lg bg-white p-6 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-base font-semibold">{title}</h3>
        <p className="mt-2 text-sm text-muted-foreground">{message}</p>
        <div className="mt-5 flex justify-end gap-2">
          <button onClick={onCancel} className="rounded-md border px-4 py-2 text-sm">
            取消
          </button>
          <button
            onClick={onConfirm}
            className={`rounded-md px-4 py-2 text-sm font-medium text-white ${variant === 'danger' ? 'bg-red-600 hover:bg-red-700' : 'bg-amber-600 hover:bg-amber-700'}`}
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}

export function ModelsPage() {
  const { user } = useSession()
  const canEdit = user?.role === 'admin' || user?.role === 'buyer'
  const apis = useApis()

  const { data, loading, filter, search, refresh } = useFilteredQuery({
    initialFilter: { page: 1, pageSize: 20, keyword: '', modelType: '', status: '' },
    queryKeyFactory: (f) => queryKeys.models.list(f),
    fetcher: (a, f) => a.modelsApi.list(f),
  })

  const [syncing, setSyncing] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<AiModel | null>(null)
  const [priceForm, setPriceForm] = useState({ inputPrice: '', outputPrice: '', discount: '' })
  const [saving, setSaving] = useState(false)

  // 二次确认状态
  const [confirmAction, setConfirmAction] = useState<{
    type: 'delete' | 'toggle'
    model: AiModel
  } | null>(null)

  const handleSync = async () => {
    setSyncing(true)
    try {
      const result = await apis.newapiApi.pull()
      toast.success(`同步完成：${result.channelsSynced} 个渠道，${result.modelsCreated} 个模型`)
      refresh()
    } catch (e: any) {
      toast.error(`同步失败：${e.message}`)
    } finally {
      setSyncing(false)
    }
  }

  const openPricing = (m: AiModel) => {
    setEditing(m)
    setPriceForm({
      inputPrice: m.inputPrice?.toString() ?? '',
      outputPrice: m.outputPrice?.toString() ?? '',
      discount: m.discount?.toString() ?? '',
    })
    setDialogOpen(true)
  }

  const handleSavePricing = async () => {
    if (!editing) return
    setSaving(true)
    try {
      await apis.modelsApi.updatePricing(editing.id, {
        inputPrice: priceForm.inputPrice ? Number(priceForm.inputPrice) : undefined,
        outputPrice: priceForm.outputPrice ? Number(priceForm.outputPrice) : undefined,
        discount: priceForm.discount ? Number(priceForm.discount) : undefined,
      })
      toast.success('售价更新成功')
      setDialogOpen(false)
      refresh()
    } catch (e: any) {
      toast.error(e.message)
    } finally {
      setSaving(false)
    }
  }

  const handleConfirm = async () => {
    if (!confirmAction) return
    const { type, model } = confirmAction
    setConfirmAction(null)
    try {
      if (type === 'delete') {
        await apis.modelsApi.delete(model.id)
        toast.success(`已删除「${model.modelName}」`)
      } else {
        const newStatus = model.status === 'available' ? 'deprecated' : 'available'
        await apis.modelsApi.update(model.id, { status: newStatus })
        toast.success(
          newStatus === 'deprecated'
            ? `已禁用「${model.modelName}」`
            : `已启用「${model.modelName}」`,
        )
      }
      refresh()
    } catch (e: any) {
      toast.error(e.message)
    }
  }

  const typeFilters = ['全部', ...MODEL_TYPES]

  return (
    <PageShell>
      <PageHeader
        title="AI 模型目录"
        description={`从 NewAPI 同步 · 共 ${data?.total ?? 0} 个模型`}
        actions={
          canEdit ? (
            <button
              onClick={handleSync}
              disabled={syncing}
              className="inline-flex items-center gap-1.5 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
            >
              <RefreshCw className={`h-4 w-4 ${syncing ? 'animate-spin' : ''}`} />
              {syncing ? '同步中...' : '同步模型'}
            </button>
          ) : undefined
        }
      />

      {/* Filters */}
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
          <input
            className="ml-auto h-8 w-44 rounded-md border px-3 text-sm"
            placeholder="搜索模型名称"
            value={filter.keyword}
            onChange={(e) => search({ keyword: e.target.value })}
          />
        </div>
      </div>

      {/* Table */}
      <div className="rounded-lg border bg-white">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/30 text-xs text-muted-foreground">
              <th className="px-4 py-3 text-left font-medium">模型</th>
              <th className="px-4 py-3 text-left font-medium">渠道</th>
              <th className="px-4 py-3 text-right font-medium">成本价 (输入/输出)</th>
              <th className="px-4 py-3 text-right font-medium">售价 (输入/输出)</th>
              <th className="px-4 py-3 text-center font-medium">状态</th>
              {canEdit && <th className="px-4 py-3 text-center font-medium">操作</th>}
            </tr>
          </thead>
          <tbody>
            {data?.items.map((m) => (
              <tr
                key={m.id}
                className={`border-b last:border-0 hover:bg-muted/20 ${m.status === 'deprecated' ? 'opacity-50' : ''}`}
              >
                <td className="px-4 py-3">
                  <div className="font-medium">{m.modelName}</div>
                  {m.modelId && (
                    <div className="font-mono text-[11px] text-muted-foreground">{m.modelId}</div>
                  )}
                </td>
                <td className="px-4 py-3 text-muted-foreground">{m.channelName ?? '—'}</td>
                <td className="px-4 py-3 text-right font-mono text-xs text-muted-foreground">
                  {m.costInput != null ? `¥${m.costInput}` : '—'} /{' '}
                  {m.costOutput != null ? `¥${m.costOutput}` : '—'}
                </td>
                <td className="px-4 py-3 text-right">
                  <span className="font-mono text-xs font-medium">
                    {m.inputPrice != null ? `¥${m.inputPrice}` : '—'} /{' '}
                    {m.outputPrice != null ? `¥${m.outputPrice}` : '—'}
                  </span>
                  {m.discount != null && m.discount > 0 && (
                    <span className="ml-1 text-[10px] text-green-600">-{m.discount}%</span>
                  )}
                </td>
                <td className="px-4 py-3 text-center">
                  <Badge variant={m.status === 'available' ? 'success' : 'outline'}>
                    {MODEL_STATUS[m.status]?.label ?? m.status}
                  </Badge>
                </td>
                {canEdit && (
                  <td className="px-4 py-3 text-center">
                    <div className="inline-flex items-center gap-1">
                      <button
                        onClick={() => openPricing(m)}
                        className="rounded p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground"
                        title="编辑售价"
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </button>
                      <button
                        onClick={() => setConfirmAction({ type: 'toggle', model: m })}
                        className={`rounded p-1.5 hover:bg-muted ${m.status === 'available' ? 'text-amber-500 hover:text-amber-600' : 'text-green-500 hover:text-green-600'}`}
                        title={m.status === 'available' ? '禁用' : '启用'}
                      >
                        <Power className="h-3.5 w-3.5" />
                      </button>
                      <button
                        onClick={() => setConfirmAction({ type: 'delete', model: m })}
                        className="rounded p-1.5 text-muted-foreground hover:bg-red-50 hover:text-red-600"
                        title="删除"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  </td>
                )}
              </tr>
            ))}
            {!loading && data?.items.length === 0 && (
              <tr>
                <td colSpan={canEdit ? 6 : 5} className="py-16 text-center text-muted-foreground">
                  暂无模型数据，点击「同步模型」从 NewAPI 拉取
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* 售价编辑弹窗 */}
      {dialogOpen && editing && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          onClick={() => setDialogOpen(false)}
        >
          <div
            className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="mb-1 text-lg font-semibold">编辑售价</h2>
            <p className="mb-4 text-sm text-muted-foreground">
              {editing.modelName}
              {editing.modelId && <span className="ml-1 font-mono">({editing.modelId})</span>}
            </p>
            <div className="mb-4 rounded-md bg-muted/50 px-3 py-2 text-xs text-muted-foreground">
              成本价：输入 ¥{editing.costInput ?? '—'} / 输出 ¥{editing.costOutput ?? '—'} (每百万
              tokens)
            </div>
            <div className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <Field label="售价 · 输入 (¥/M tokens)">
                  <input
                    className="input"
                    type="number"
                    step="0.01"
                    value={priceForm.inputPrice}
                    onChange={(e) => setPriceForm({ ...priceForm, inputPrice: e.target.value })}
                    placeholder={editing.costInput?.toString() ?? '0'}
                  />
                </Field>
                <Field label="售价 · 输出 (¥/M tokens)">
                  <input
                    className="input"
                    type="number"
                    step="0.01"
                    value={priceForm.outputPrice}
                    onChange={(e) => setPriceForm({ ...priceForm, outputPrice: e.target.value })}
                    placeholder={editing.costOutput?.toString() ?? '0'}
                  />
                </Field>
              </div>
              <Field label="折扣 (%)">
                <input
                  className="input"
                  type="number"
                  value={priceForm.discount}
                  onChange={(e) => setPriceForm({ ...priceForm, discount: e.target.value })}
                  placeholder="0"
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
                onClick={handleSavePricing}
                disabled={saving}
                className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
              >
                {saving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 二次确认弹窗 */}
      <ConfirmDialog
        open={!!confirmAction}
        title={
          confirmAction?.type === 'delete'
            ? '确认删除'
            : confirmAction?.model.status === 'available'
              ? '确认禁用'
              : '确认启用'
        }
        message={
          confirmAction?.type === 'delete'
            ? `确定删除模型「${confirmAction.model.modelName}」吗？此操作不可撤销。`
            : confirmAction?.model.status === 'available'
              ? `确定禁用模型「${confirmAction?.model.modelName}」吗？禁用后用户将无法使用该模型。`
              : `确定启用模型「${confirmAction?.model.modelName}」吗？`
        }
        confirmLabel={
          confirmAction?.type === 'delete'
            ? '删除'
            : confirmAction?.model.status === 'available'
              ? '禁用'
              : '启用'
        }
        variant={confirmAction?.type === 'delete' ? 'danger' : 'warning'}
        onConfirm={handleConfirm}
        onCancel={() => setConfirmAction(null)}
      />
    </PageShell>
  )
}
