import { useState, useEffect, useMemo } from 'react'
import { Plus, Pencil, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useFilteredQuery, useInjectedMutation, queryKeys } from '@/features/query'
import { useSession } from '@/features/session'
import { useSupplierOptions } from '@/features/suppliers'
import { useApis } from '@/api/use-apis'
import { EVAL_GRADE, DIMENSIONS } from '@/config/enums'
import { Field, Pagination } from '@/components/ui'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import type { Evaluation } from '@/api/evaluations'

export function EvaluationsPage() {
  const { user } = useSession()
  const canEdit = user?.role === 'admin' || user?.role === 'buyer'
  const apis = useApis()
  const suppliers = useSupplierOptions()

  const [weights, setWeights] = useState<Record<string, number>>({})
  useEffect(() => {
    apis.evaluationsApi.getWeights().then((ws) => {
      setWeights(Object.fromEntries(ws.map((w) => [w.dimension, w.weight])))
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const { data, loading, filter, setFilter, search } = useFilteredQuery({
    initialFilter: { page: 1, pageSize: 10, supplierId: 0, period: '' },
    queryKeyFactory: (f) => queryKeys.evaluations.list(f),
    fetcher: (a, f) => a.evaluationsApi.list(f),
  })

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<Evaluation | null>(null)
  const [form, setForm] = useState({
    supplierId: 0,
    period: '',
    quality: 80,
    performance: 80,
    price: 80,
    service: 80,
    compliance: 80,
    comment: '',
  })
  const [saving, setSaving] = useState(false)

  const deleteMut = useInjectedMutation<void, number>({
    mutationFn: (a, id) => a.evaluationsApi.delete(id),
    invalidateKeys: [queryKeys.evaluations.all],
    onSuccess: () => toast.success('删除成功'),
  })

  const previewScore = useMemo(() => {
    const total =
      form.quality * (weights.quality ?? 30) +
      form.performance * (weights.performance ?? 20) +
      form.price * (weights.price ?? 20) +
      form.service * (weights.service ?? 20) +
      form.compliance * (weights.compliance ?? 10)
    return Math.round(total) / 100
  }, [form, weights])

  const previewGrade =
    previewScore >= 90 ? 'A' : previewScore >= 80 ? 'B' : previewScore >= 60 ? 'C' : 'D'

  const openCreate = () => {
    setEditing(null)
    setForm({
      supplierId: 0,
      period: '',
      quality: 80,
      performance: 80,
      price: 80,
      service: 80,
      compliance: 80,
      comment: '',
    })
    setDialogOpen(true)
  }

  const openEdit = (row: Evaluation) => {
    setEditing(row)
    setForm({
      supplierId: row.supplierId,
      period: row.period,
      quality: row.quality,
      performance: row.performance,
      price: row.price,
      service: row.service,
      compliance: row.compliance,
      comment: row.comment ?? '',
    })
    setDialogOpen(true)
  }

  const handleSave = async () => {
    if (!form.supplierId || !form.period) {
      toast.error('供应商和评估周期不能为空')
      return
    }
    setSaving(true)
    try {
      if (editing) {
        await apis.evaluationsApi.update(editing.id, form)
        toast.success('更新成功')
      } else {
        await apis.evaluationsApi.create(form)
        toast.success('评估提交成功')
      }
      setDialogOpen(false)
      search({})
    } catch (e: any) {
      toast.error(e.message)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = (row: Evaluation) => {
    if (!confirm(`确定删除「${row.supplierName} · ${row.period}」的评估吗？`)) return
    deleteMut.mutate(row.id)
  }

  const totalPages = Math.ceil((data?.total ?? 0) / filter.pageSize)
  const dims = [
    { key: 'quality' as const, label: '质量' },
    { key: 'performance' as const, label: '性能' },
    { key: 'price' as const, label: '价格' },
    { key: 'service' as const, label: '服务' },
    { key: 'compliance' as const, label: '合规' },
  ]

  return (
    <PageShell>
      <PageHeader
        title="绩效评估"
        actions={
          canEdit ? (
            <button
              onClick={openCreate}
              className="inline-flex h-9 items-center gap-1.5 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground"
            >
              <Plus className="h-4 w-4" /> 新建评估
            </button>
          ) : undefined
        }
      />

      <div className="flex items-center gap-2">
        <select
          className="h-9 rounded-md border px-2 text-sm"
          value={filter.supplierId}
          onChange={(e) => search({ supplierId: Number(e.target.value) })}
        >
          <option value={0}>全部供应商</option>
          {suppliers.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name}
            </option>
          ))}
        </select>
        <input
          className="h-9 w-40 rounded-md border px-3 text-sm"
          placeholder="评估周期 如 2026-Q3"
          value={filter.period}
          onChange={(e) => search({ period: e.target.value })}
        />
      </div>

      <div className="rounded-lg border bg-white">
        <table className="w-full text-sm">
          <thead className="border-b bg-muted/40">
            <tr>
              <th className="px-4 py-3 text-left font-medium">供应商</th>
              <th className="px-4 py-3 text-left font-medium">周期</th>
              <th className="px-4 py-3 text-center font-medium">评级</th>
              <th className="px-4 py-3 text-right font-medium">综合分</th>
              <th className="px-4 py-3 text-right font-medium">质量</th>
              <th className="px-4 py-3 text-right font-medium">性能</th>
              <th className="px-4 py-3 text-right font-medium">价格</th>
              <th className="px-4 py-3 text-right font-medium">服务</th>
              <th className="px-4 py-3 text-right font-medium">合规</th>
              <th className="px-4 py-3 text-left font-medium">评估人</th>
              {canEdit && <th className="px-4 py-3 text-right font-medium">操作</th>}
            </tr>
          </thead>
          <tbody>
            {data?.items.map((e) => (
              <tr key={e.id} className="border-b last:border-0 hover:bg-muted/20">
                <td className="px-4 py-3">{e.supplierName ?? '-'}</td>
                <td className="px-4 py-3 text-muted-foreground">{e.period}</td>
                <td className="px-4 py-3 text-center">
                  <span
                    className={`inline-flex rounded px-2 py-0.5 text-xs font-bold text-white ${e.grade === 'A' ? 'bg-green-500' : e.grade === 'B' ? 'bg-blue-500' : e.grade === 'C' ? 'bg-yellow-500' : 'bg-red-500'}`}
                  >
                    {e.grade}
                  </span>
                </td>
                <td className="px-4 py-3 text-right font-medium">{e.totalScore}</td>
                <td className="px-4 py-3 text-right text-muted-foreground">{e.quality}</td>
                <td className="px-4 py-3 text-right text-muted-foreground">{e.performance}</td>
                <td className="px-4 py-3 text-right text-muted-foreground">{e.price}</td>
                <td className="px-4 py-3 text-right text-muted-foreground">{e.service}</td>
                <td className="px-4 py-3 text-right text-muted-foreground">{e.compliance}</td>
                <td className="px-4 py-3 text-muted-foreground">{e.evaluatorName ?? '-'}</td>
                {canEdit && (
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => openEdit(e)}
                      className="px-1.5 py-1 text-muted-foreground hover:text-primary"
                    >
                      <Pencil className="inline h-3.5 w-3.5" />
                    </button>
                    <button
                      onClick={() => handleDelete(e)}
                      className="px-1.5 py-1 text-muted-foreground hover:text-red-500"
                    >
                      <Trash2 className="inline h-3.5 w-3.5" />
                    </button>
                  </td>
                )}
              </tr>
            ))}
            {!loading && data?.items.length === 0 && (
              <tr>
                <td colSpan={11} className="px-4 py-12 text-center text-muted-foreground">
                  暂无数据
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination
        page={filter.page}
        totalPages={totalPages}
        total={data?.total ?? 0}
        onChange={(p) => setFilter((prev) => ({ ...prev, page: p }))}
      />

      {/* 打分弹窗 */}
      {dialogOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          onClick={() => setDialogOpen(false)}
        >
          <div
            className="w-full max-w-2xl rounded-lg bg-white p-6 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="mb-4 text-lg font-semibold">{editing ? '修改评估' : '新建评估'}</h2>
            <div className="flex gap-6">
              {/* 左侧表单 */}
              <div className="flex-1 space-y-3">
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
                <Field label="评估周期" required>
                  <input
                    className="input"
                    value={form.period}
                    onChange={(e) => setForm({ ...form, period: e.target.value })}
                    placeholder="如：2026-Q3"
                    disabled={!!editing}
                  />
                </Field>
                {dims.map((d) => (
                  <div key={d.key}>
                    <div className="mb-1 flex items-center justify-between">
                      <label className="text-sm font-medium">{DIMENSIONS[d.key] ?? d.label}</label>
                      <span className="text-xs text-muted-foreground">
                        权重 {weights[d.key] ?? '-'}%
                      </span>
                    </div>
                    <div className="flex items-center gap-3">
                      <input
                        type="range"
                        min={0}
                        max={100}
                        value={form[d.key]}
                        onChange={(e) => setForm({ ...form, [d.key]: Number(e.target.value) })}
                        className="flex-1"
                      />
                      <input
                        type="number"
                        min={0}
                        max={100}
                        className="w-16 rounded border px-2 py-1 text-center text-sm"
                        value={form[d.key]}
                        onChange={(e) => setForm({ ...form, [d.key]: Number(e.target.value) })}
                      />
                    </div>
                  </div>
                ))}
                <Field label="评语">
                  <textarea
                    className="input min-h-[60px] resize-none"
                    value={form.comment}
                    onChange={(e) => setForm({ ...form, comment: e.target.value })}
                    placeholder="整体评价"
                  />
                </Field>
              </div>
              {/* 右侧预览 */}
              <div className="flex w-36 flex-col items-center justify-center rounded-lg border bg-muted/30 p-4">
                <div className="text-xs text-muted-foreground">综合分预览</div>
                <div className="my-2 text-3xl font-bold">{previewScore.toFixed(1)}</div>
                <span
                  className={`rounded px-3 py-1 text-sm font-bold text-white ${previewGrade === 'A' ? 'bg-green-500' : previewGrade === 'B' ? 'bg-blue-500' : previewGrade === 'C' ? 'bg-yellow-500' : 'bg-red-500'}`}
                >
                  {EVAL_GRADE[previewGrade]?.label ?? previewGrade}
                </span>
                <div className="mt-2 text-center text-xs text-muted-foreground">
                  按当前权重自动计算
                </div>
              </div>
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
                {saving ? '提交中...' : '提交评估'}
              </button>
            </div>
          </div>
        </div>
      )}
    </PageShell>
  )
}
