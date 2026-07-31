import { useState, useEffect, useMemo } from 'react'
import { Plus, Pencil, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useFilteredQuery, useInjectedMutation, queryKeys } from '@/features/query'
import { useSession } from '@/features/session'
import { useSupplierOptions } from '@/features/suppliers'
import { useApis } from '@/api/use-apis'
import { EVAL_GRADE, DIMENSIONS } from '@/config/enums'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { ConfirmActionDialog, type ConfirmActionState } from '@/components/ui/confirm-action-dialog'
import { Pagination } from '@/components/ui/pagination'
import { NativeSelect } from '@/components/ui/native-select'
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
    initialFilter: { page: 1, pageSize: 10, supplierId: '', period: '' },
    queryKeyFactory: (f) => queryKeys.evaluations.list(f),
    fetcher: (a, f) => a.evaluationsApi.list(f),
  })

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<Evaluation | null>(null)
  const [form, setForm] = useState({
    supplierId: '',
    period: '',
    quality: 80,
    performance: 80,
    price: 80,
    service: 80,
    compliance: 80,
    comment: '',
  })
  const [saving, setSaving] = useState(false)
  const [confirmState, setConfirmState] = useState<ConfirmActionState | null>(null)

  const deleteMut = useInjectedMutation<void, string>({
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
      supplierId: '',
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
    setConfirmState({
      open: true,
      title: '确认删除',
      desc: `确定删除「${row.supplierName} · ${row.period}」的评估吗？`,
      variant: 'danger',
      confirmLabel: '删除',
      onConfirm: () => {
        deleteMut.mutate(row.id)
        setConfirmState(null)
      },
    })
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
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" /> 新建评估
            </Button>
          ) : undefined
        }
      />

      <div className="flex items-center gap-2">
        <NativeSelect
          className="h-9 w-auto"
          value={filter.supplierId}
          onChange={(e) => search({ supplierId: e.target.value })}
        >
          <option value="">全部供应商</option>
          {suppliers.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name}
            </option>
          ))}
        </NativeSelect>
        <Input
          className="h-9 w-40"
          placeholder="评估周期 如 2026-Q3"
          value={filter.period}
          onChange={(e) => search({ period: e.target.value })}
        />
      </div>

      <div className="rounded-lg border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>供应商</TableHead>
              <TableHead>周期</TableHead>
              <TableHead className="text-center">评级</TableHead>
              <TableHead className="text-right">综合分</TableHead>
              <TableHead className="text-right">质量</TableHead>
              <TableHead className="text-right">性能</TableHead>
              <TableHead className="text-right">价格</TableHead>
              <TableHead className="text-right">服务</TableHead>
              <TableHead className="text-right">合规</TableHead>
              <TableHead>评估人</TableHead>
              {canEdit && <TableHead className="text-right">操作</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {data?.items.map((e) => (
              <TableRow key={e.id}>
                <TableCell>{e.supplierName ?? '-'}</TableCell>
                <TableCell className="text-muted-foreground">{e.period}</TableCell>
                <TableCell className="text-center">
                  <span
                    className={`inline-flex rounded px-2 py-0.5 text-xs font-bold text-white ${e.grade === 'A' ? 'bg-green-500' : e.grade === 'B' ? 'bg-blue-500' : e.grade === 'C' ? 'bg-yellow-500' : 'bg-red-500'}`}
                  >
                    {e.grade}
                  </span>
                </TableCell>
                <TableCell className="text-right font-medium">{e.totalScore}</TableCell>
                <TableCell className="text-right text-muted-foreground">{e.quality}</TableCell>
                <TableCell className="text-right text-muted-foreground">{e.performance}</TableCell>
                <TableCell className="text-right text-muted-foreground">{e.price}</TableCell>
                <TableCell className="text-right text-muted-foreground">{e.service}</TableCell>
                <TableCell className="text-right text-muted-foreground">{e.compliance}</TableCell>
                <TableCell className="text-muted-foreground">{e.evaluatorName ?? '-'}</TableCell>
                {canEdit && (
                  <TableCell className="text-right">
                    <Button variant="ghost" size="icon-sm" onClick={() => openEdit(e)}>
                      <Pencil className="h-3.5 w-3.5" />
                    </Button>
                    <Button variant="ghost" size="icon-sm" onClick={() => handleDelete(e)}>
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </TableCell>
                )}
              </TableRow>
            ))}
            {!loading && data?.items.length === 0 && (
              <TableRow>
                <TableCell colSpan={11} className="py-12 text-center text-muted-foreground">
                  暂无数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <Pagination
        page={filter.page}
        totalPages={totalPages}
        total={data?.total ?? 0}
        onChange={(p) => setFilter((prev) => ({ ...prev, page: p }))}
      />

      {/* 打分弹窗 */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editing ? '修改评估' : '新建评估'}</DialogTitle>
          </DialogHeader>
          <div className="flex gap-6">
            {/* 左侧表单 */}
            <div className="flex-1 space-y-3">
              <div>
                <Label>
                  供应商 <span className="text-red-500">*</span>
                </Label>
                <NativeSelect
                  className="mt-1"
                  value={form.supplierId}
                  onChange={(e) => setForm({ ...form, supplierId: e.target.value })}
                  disabled={!!editing}
                >
                  <option value="">请选择</option>
                  {suppliers.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </NativeSelect>
              </div>
              <div>
                <Label>
                  评估周期 <span className="text-red-500">*</span>
                </Label>
                <Input
                  className="mt-1"
                  value={form.period}
                  onChange={(e) => setForm({ ...form, period: e.target.value })}
                  placeholder="如：2026-Q3"
                  disabled={!!editing}
                />
              </div>
              {dims.map((d) => (
                <div key={d.key}>
                  <div className="mb-1 flex items-center justify-between">
                    <Label>{DIMENSIONS[d.key] ?? d.label}</Label>
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
              <div>
                <Label>评语</Label>
                <Textarea
                  className="mt-1 min-h-[60px] resize-none"
                  value={form.comment}
                  onChange={(e) => setForm({ ...form, comment: e.target.value })}
                  placeholder="整体评价"
                />
              </div>
            </div>
            {/* 右侧预览 */}
            <Card className="w-36">
              <CardContent className="flex h-full flex-col items-center justify-center p-4">
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
              </CardContent>
            </Card>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={handleSave} disabled={saving}>
              {saving ? '提交中...' : '提交评估'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmActionDialog
        state={confirmState}
        onOpenChange={(open) => !open && setConfirmState(null)}
        onClose={() => setConfirmState(null)}
      />
    </PageShell>
  )
}
