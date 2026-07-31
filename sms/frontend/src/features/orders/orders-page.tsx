import { useState } from 'react'
import { Plus, Pencil, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useFilteredQuery, useInjectedMutation, queryKeys } from '@/features/query'
import { useSession } from '@/features/session'
import { useSupplierOptions } from '@/features/suppliers'
import { useApis } from '@/api/use-apis'
import { ORDER_STATUS } from '@/config/enums'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { StatusBadge } from '@/components/ui/badge'
import { Pagination } from '@/components/ui/pagination'
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
import { NativeSelect } from '@/components/ui/native-select'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import type { PurchaseOrder } from '@/api/orders'

export function OrdersPage() {
  const { user } = useSession()
  const canEdit = user?.role === 'admin' || user?.role === 'buyer'
  const apis = useApis()
  const suppliers = useSupplierOptions()

  const { data, loading, filter, setFilter, search } = useFilteredQuery({
    initialFilter: { page: 1, pageSize: 10, keyword: '', supplierId: '', status: '' },
    queryKeyFactory: (f) => queryKeys.orders.list(f),
    fetcher: (a, f) => a.ordersApi.list(f),
  })

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<PurchaseOrder | null>(null)
  const [form, setForm] = useState({
    orderNo: '',
    supplierId: '',
    contractId: undefined as string | undefined,
    totalAmount: '',
    orderDate: '',
    status: 'pending',
    description: '',
  })
  const [saving, setSaving] = useState(false)
  const [confirmState, setConfirmState] = useState<ConfirmActionState | null>(null)

  const deleteMut = useInjectedMutation<void, string>({
    mutationFn: (a, id) => a.ordersApi.delete(id),
    invalidateKeys: [queryKeys.orders.all],
    onSuccess: () => toast.success('删除成功'),
  })

  const openCreate = () => {
    setEditing(null)
    setForm({
      orderNo: '',
      supplierId: '',
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
    setConfirmState({
      open: true,
      title: '确认删除',
      desc: `确定删除订单「${row.orderNo}」吗？`,
      variant: 'danger',
      confirmLabel: '删除',
      onConfirm: () => {
        deleteMut.mutate(row.id)
        setConfirmState(null)
      },
    })
  }

  const totalPages = Math.ceil((data?.total ?? 0) / filter.pageSize)

  return (
    <PageShell>
      <PageHeader
        title="采购订单"
        actions={
          canEdit ? (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" /> 新建订单
            </Button>
          ) : undefined
        }
      />

      <div className="flex items-center gap-2">
        <Input
          className="h-9 w-44"
          placeholder="订单编号"
          value={filter.keyword}
          onChange={(e) => search({ keyword: e.target.value })}
        />
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
        <NativeSelect
          className="h-9 w-auto"
          value={filter.status}
          onChange={(e) => search({ status: e.target.value })}
        >
          <option value="">全部状态</option>
          {Object.entries(ORDER_STATUS).map(([k, v]) => (
            <option key={k} value={k}>
              {v.label}
            </option>
          ))}
        </NativeSelect>
      </div>

      <div className="rounded-lg border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>订单编号</TableHead>
              <TableHead>供应商</TableHead>
              <TableHead>关联合同</TableHead>
              <TableHead className="text-right">金额</TableHead>
              <TableHead>下单日期</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>创建人</TableHead>
              {canEdit && <TableHead className="text-right">操作</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {data?.items.map((o) => (
              <TableRow key={o.id}>
                <TableCell className="text-muted-foreground">{o.orderNo}</TableCell>
                <TableCell>{o.supplierName ?? '-'}</TableCell>
                <TableCell className="text-muted-foreground">{o.contractNo ?? '-'}</TableCell>
                <TableCell className="text-right">
                  {o.totalAmount != null
                    ? Number(o.totalAmount).toLocaleString('zh-CN', { minimumFractionDigits: 2 })
                    : '-'}
                </TableCell>
                <TableCell className="text-muted-foreground">{o.orderDate ?? '-'}</TableCell>
                <TableCell>
                  <StatusBadge status={o.status} map={ORDER_STATUS} />
                </TableCell>
                <TableCell className="text-muted-foreground">{o.creatorName ?? '-'}</TableCell>
                {canEdit && (
                  <TableCell className="text-right">
                    <Button variant="ghost" size="icon-sm" onClick={() => openEdit(o)}>
                      <Pencil className="h-3.5 w-3.5" />
                    </Button>
                    <Button variant="ghost" size="icon-sm" onClick={() => handleDelete(o)}>
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </TableCell>
                )}
              </TableRow>
            ))}
            {!loading && data?.items.length === 0 && (
              <TableRow>
                <TableCell colSpan={8} className="py-12 text-center text-muted-foreground">
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

      {/* 新建/编辑弹窗 */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{editing ? '编辑订单' : '新建订单'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <Label>
                订单编号 <span className="text-red-500">*</span>
              </Label>
              <Input
                className="mt-1"
                value={form.orderNo}
                onChange={(e) => setForm({ ...form, orderNo: e.target.value })}
                disabled={!!editing}
                placeholder="如：PO-2026-0001"
              />
            </div>
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
              <Label>采购金额</Label>
              <Input
                className="mt-1"
                type="number"
                step="0.01"
                value={form.totalAmount}
                onChange={(e) => setForm({ ...form, totalAmount: e.target.value })}
              />
            </div>
            <div>
              <Label>下单日期</Label>
              <Input
                className="mt-1"
                type="date"
                value={form.orderDate}
                onChange={(e) => setForm({ ...form, orderDate: e.target.value })}
              />
            </div>
            <div>
              <Label>状态</Label>
              <NativeSelect
                className="mt-1"
                value={form.status}
                onChange={(e) => setForm({ ...form, status: e.target.value })}
              >
                {Object.entries(ORDER_STATUS).map(([k, v]) => (
                  <option key={k} value={k}>
                    {v.label}
                  </option>
                ))}
              </NativeSelect>
            </div>
            <div>
              <Label>说明</Label>
              <Textarea
                className="mt-1 min-h-[60px] resize-none"
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                placeholder="订单说明"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={handleSave} disabled={saving}>
              {saving ? '保存中...' : '保存'}
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
