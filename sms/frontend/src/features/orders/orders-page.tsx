import { useState } from 'react'
import { Plus, Pencil, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useFilteredQuery, useInjectedMutation, queryKeys } from '@/features/query'
import { useSession } from '@/features/session'
import { useSupplierOptions } from '@/features/suppliers'
import { useApis } from '@/api/use-apis'
import { ORDER_STATUS } from '@/config/enums'
import { StatusBadge, Field, Pagination } from '@/components/ui'
import type { PurchaseOrder } from '@/api/orders'

export function OrdersPage() {
  const { user } = useSession()
  const canEdit = user?.role === 'admin' || user?.role === 'buyer'
  const apis = useApis()
  const suppliers = useSupplierOptions()

  const { data, loading, filter, setFilter, search } = useFilteredQuery({
    initialFilter: { page: 1, pageSize: 10, keyword: '', supplierId: 0, status: '' },
    queryKeyFactory: (f) => queryKeys.orders.list(f),
    fetcher: (a, f) => a.ordersApi.list(f),
  })

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<PurchaseOrder | null>(null)
  const [form, setForm] = useState({
    orderNo: '',
    supplierId: 0,
    contractId: undefined as number | undefined,
    totalAmount: '',
    orderDate: '',
    status: 'pending',
    description: '',
  })
  const [saving, setSaving] = useState(false)

  const deleteMut = useInjectedMutation<void, number>({
    mutationFn: (a, id) => a.ordersApi.delete(id),
    invalidateKeys: [queryKeys.orders.all],
    onSuccess: () => toast.success('删除成功'),
  })

  const openCreate = () => {
    setEditing(null)
    setForm({
      orderNo: '',
      supplierId: 0,
      contractId: undefined,
      totalAmount: '',
      orderDate: '',
      status: 'pending',
      description: '',
    })
    setDialogOpen(true)
  }

  const openEdit = (row: PurchaseOrder) => {
    setEditing(row)
    setForm({
      orderNo: row.orderNo,
      supplierId: row.supplierId,
      contractId: row.contractId ?? undefined,
      totalAmount: row.totalAmount?.toString() ?? '',
      orderDate: row.orderDate ?? '',
      status: row.status,
      description: row.description ?? '',
    })
    setDialogOpen(true)
  }

  const handleSave = async () => {
    if (!form.orderNo || !form.supplierId) {
      toast.error('订单号和供应商不能为空')
      return
    }
    setSaving(true)
    try {
      const payload = {
        ...form,
        totalAmount: form.totalAmount ? Number(form.totalAmount) : undefined,
        contractId: form.contractId || undefined,
        orderDate: form.orderDate || undefined,
        description: form.description || undefined,
      }
      if (editing) {
        await apis.ordersApi.update(editing.id, payload)
        toast.success('更新成功')
      } else {
        await apis.ordersApi.create(payload)
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

  const handleDelete = (row: PurchaseOrder) => {
    if (!confirm(`确定删除订单「${row.orderNo}」吗？`)) return
    deleteMut.mutate(row.id)
  }

  const totalPages = Math.ceil((data?.total ?? 0) / filter.pageSize)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <input
            className="h-9 w-44 rounded-md border px-3 text-sm"
            placeholder="订单编号"
            value={filter.keyword}
            onChange={(e) => search({ keyword: e.target.value })}
          />
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
          <select
            className="h-9 rounded-md border px-2 text-sm"
            value={filter.status}
            onChange={(e) => search({ status: e.target.value })}
          >
            <option value="">全部状态</option>
            {Object.entries(ORDER_STATUS).map(([k, v]) => (
              <option key={k} value={k}>
                {v.label}
              </option>
            ))}
          </select>
        </div>
        {canEdit && (
          <button
            onClick={openCreate}
            className="inline-flex h-9 items-center gap-1.5 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            <Plus className="h-4 w-4" /> 新建订单
          </button>
        )}
      </div>

      <div className="rounded-lg border bg-white">
        <table className="w-full text-sm">
          <thead className="border-b bg-muted/40">
            <tr>
              <th className="px-4 py-3 text-left font-medium">订单编号</th>
              <th className="px-4 py-3 text-left font-medium">供应商</th>
              <th className="px-4 py-3 text-left font-medium">关联合同</th>
              <th className="px-4 py-3 text-right font-medium">金额</th>
              <th className="px-4 py-3 text-left font-medium">下单日期</th>
              <th className="px-4 py-3 text-left font-medium">状态</th>
              <th className="px-4 py-3 text-left font-medium">创建人</th>
              {canEdit && <th className="px-4 py-3 text-right font-medium">操作</th>}
            </tr>
          </thead>
          <tbody>
            {data?.items.map((o) => (
              <tr key={o.id} className="border-b last:border-0 hover:bg-muted/20">
                <td className="px-4 py-3 text-muted-foreground">{o.orderNo}</td>
                <td className="px-4 py-3">{o.supplierName ?? '-'}</td>
                <td className="px-4 py-3 text-muted-foreground">{o.contractNo ?? '-'}</td>
                <td className="px-4 py-3 text-right">
                  {o.totalAmount != null
                    ? Number(o.totalAmount).toLocaleString('zh-CN', { minimumFractionDigits: 2 })
                    : '-'}
                </td>
                <td className="px-4 py-3 text-muted-foreground">{o.orderDate ?? '-'}</td>
                <td className="px-4 py-3">
                  <StatusBadge status={o.status} map={ORDER_STATUS} />
                </td>
                <td className="px-4 py-3 text-muted-foreground">{o.creatorName ?? '-'}</td>
                {canEdit && (
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => openEdit(o)}
                      className="px-1.5 py-1 text-muted-foreground hover:text-primary"
                    >
                      <Pencil className="inline h-3.5 w-3.5" />
                    </button>
                    <button
                      onClick={() => handleDelete(o)}
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
                <td colSpan={8} className="px-4 py-12 text-center text-muted-foreground">
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

      {dialogOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          onClick={() => setDialogOpen(false)}
        >
          <div
            className="w-full max-w-lg rounded-lg bg-white p-6 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="mb-4 text-lg font-semibold">{editing ? '编辑订单' : '新建订单'}</h2>
            <div className="space-y-3">
              <Field label="订单编号" required>
                <input
                  className="input"
                  value={form.orderNo}
                  onChange={(e) => setForm({ ...form, orderNo: e.target.value })}
                  disabled={!!editing}
                  placeholder="如：PO-2026-0001"
                />
              </Field>
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
              <Field label="采购金额">
                <input
                  className="input"
                  type="number"
                  step="0.01"
                  value={form.totalAmount}
                  onChange={(e) => setForm({ ...form, totalAmount: e.target.value })}
                />
              </Field>
              <Field label="下单日期">
                <input
                  className="input"
                  type="date"
                  value={form.orderDate}
                  onChange={(e) => setForm({ ...form, orderDate: e.target.value })}
                />
              </Field>
              <Field label="状态">
                <select
                  className="input"
                  value={form.status}
                  onChange={(e) => setForm({ ...form, status: e.target.value })}
                >
                  {Object.entries(ORDER_STATUS).map(([k, v]) => (
                    <option key={k} value={k}>
                      {v.label}
                    </option>
                  ))}
                </select>
              </Field>
              <Field label="说明">
                <textarea
                  className="input min-h-[60px] resize-none"
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  placeholder="订单说明"
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
