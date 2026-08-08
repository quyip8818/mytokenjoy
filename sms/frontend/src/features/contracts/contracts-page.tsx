import { useState } from 'react'
import { Plus, Pencil, Trash2, FileText } from 'lucide-react'
import { toast } from 'sonner'
import { useFilteredQuery, useInjectedMutation, queryKeys } from '@/features/query'
import { useSession } from '@/features/session'
import { useSupplierOptions } from '@/features/suppliers'
import { useWorkflow } from '@/features/workflow'
import { useApis } from '@/api/use-apis'
import { CONTRACT_STATUS } from '@/config/enums'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { StatusBadge } from '@/components/ui/badge'
import { ActionIcon } from '@/components/ui/action-icon'
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
import { daysUntil, formatAmount } from '@/lib/utils'
import type { Contract } from '@/api/contracts'

export function ContractsPage() {
  const { user } = useSession()
  const canEdit = user?.role === 'admin' || user?.role === 'buyer'
  const apis = useApis()
  const suppliers = useSupplierOptions()
  const { open } = useWorkflow()

  const { data, loading, filter, setFilter, search } = useFilteredQuery({
    initialFilter: { page: 1, pageSize: 10, keyword: '', supplierId: '', status: '' },
    queryKeyFactory: (f) => queryKeys.contracts.list(f),
    fetcher: (a, f) => a.contractsApi.list(f),
  })

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<Contract | null>(null)
  const [form, setForm] = useState({
    contractNo: '',
    supplierId: '',
    title: '',
    amount: '',
    signDate: '',
    startDate: '',
    endDate: '',
    status: 'draft',
    remarks: '',
  })
  const [saving, setSaving] = useState(false)
  const [confirmState, setConfirmState] = useState<ConfirmActionState | null>(null)

  const deleteMut = useInjectedMutation<void, string>({
    mutationFn: (a, id) => a.contractsApi.delete(id),
    invalidateKeys: [queryKeys.contracts.all],
    onSuccess: () => toast.success('删除成功'),
  })

  const openCreate = () => {
    setEditing(null)
    setForm({
      contractNo: '',
      supplierId: '',
      title: '',
      amount: '',
      signDate: '',
      startDate: '',
      endDate: '',
      status: 'draft',
      remarks: '',
    })
    setDialogOpen(true)
  }

  const openEdit = (row: Contract) => {
    setEditing(row)
    setForm({
      contractNo: row.contractNo,
      supplierId: row.supplierId,
      title: row.title,
      amount: row.amount?.toString() ?? '',
      signDate: row.signDate ?? '',
      startDate: row.startDate ?? '',
      endDate: row.endDate ?? '',
      status: row.status,
      remarks: row.remarks ?? '',
    })
    setDialogOpen(true)
  }

  const handleSave = async () => {
    if (!form.contractNo || !form.title || !form.supplierId) {
      toast.error('合同编号、标题和供应商不能为空')
      return
    }
    setSaving(true)
    try {
      const payload = {
        ...form,
        amount: form.amount ? Number(form.amount) : undefined,
        signDate: form.signDate || undefined,
        startDate: form.startDate || undefined,
        endDate: form.endDate || undefined,
        remarks: form.remarks || undefined,
      }
      if (editing) {
        await apis.contractsApi.update(editing.id, payload)
        toast.success('更新成功')
      } else {
        await apis.contractsApi.create(payload)
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

  const handleDelete = (row: Contract) => {
    setConfirmState({
      open: true,
      title: '确认删除',
      desc: `确定删除合同「${row.title}」吗？`,
      variant: 'danger',
      confirmLabel: '删除',
      onConfirm: () => {
        deleteMut.mutate(row.id)
        setConfirmState(null)
      },
    })
  }

  const openDetail = async (row: Contract) => {
    const d = await apis.contractsApi.detail(row.id)
    open('contract-detail', { contract: d, canEdit, onRefresh: () => search({}) })
  }

  const totalPages = Math.ceil((data?.total ?? 0) / filter.pageSize)

  return (
    <PageShell>
      <PageHeader
        title="合同管理"
        actions={
          canEdit ? (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" /> 新建合同
            </Button>
          ) : undefined
        }
      />

      <div className="flex items-center gap-2">
        <Input
          className="h-9 w-48"
          placeholder="合同编号 / 标题"
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
          {Object.entries(CONTRACT_STATUS).map(([k, v]) => (
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
              <TableHead>合同编号</TableHead>
              <TableHead>标题</TableHead>
              <TableHead>供应商</TableHead>
              <TableHead className="text-right">金额</TableHead>
              <TableHead>到期日</TableHead>
              <TableHead className="text-right">剩余</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-center">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data?.items.map((c) => {
              const days = daysUntil(c.endDate)
              return (
                <TableRow key={c.id}>
                  <TableCell className="text-muted-foreground">{c.contractNo}</TableCell>
                  <TableCell>
                    <button
                      onClick={() => openDetail(c)}
                      className="text-left text-primary hover:underline"
                    >
                      {c.title}
                    </button>
                  </TableCell>
                  <TableCell>{c.supplierName ?? '-'}</TableCell>
                  <TableCell className="text-right">{formatAmount(c.amount)}</TableCell>
                  <TableCell className="text-muted-foreground">{c.endDate ?? '-'}</TableCell>
                  <TableCell
                    className={`text-right text-xs font-medium ${days === null ? '' : days < 0 ? 'text-red-500' : days <= 30 ? 'text-yellow-600' : 'text-muted-foreground'}`}
                  >
                    {days === null ? '-' : days < 0 ? '已过期' : `${days} 天`}
                  </TableCell>
                  <TableCell>
                    <StatusBadge status={c.status} map={CONTRACT_STATUS} />
                  </TableCell>
                  <TableCell className="text-center">
                    <div className="inline-flex items-center gap-1">
                      <ActionIcon hint="查看详情" onClick={() => openDetail(c)}>
                        <FileText className="h-5 w-5" />
                      </ActionIcon>
                      {canEdit && (
                        <>
                          <ActionIcon hint="编辑" onClick={() => openEdit(c)}>
                            <Pencil className="h-5 w-5" />
                          </ActionIcon>
                          <ActionIcon hint="删除" onClick={() => handleDelete(c)}>
                            <Trash2 className="h-5 w-5" />
                          </ActionIcon>
                        </>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              )
            })}
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
            <DialogTitle>{editing ? '编辑合同' : '新建合同'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <Label>
                合同编号 <span className="text-red-500">*</span>
              </Label>
              <Input
                className="mt-1"
                value={form.contractNo}
                onChange={(e) => setForm({ ...form, contractNo: e.target.value })}
                disabled={!!editing}
                placeholder="如：HT-2026-001"
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
              <Label>
                合同标题 <span className="text-red-500">*</span>
              </Label>
              <Input
                className="mt-1"
                value={form.title}
                onChange={(e) => setForm({ ...form, title: e.target.value })}
              />
            </div>
            <div>
              <Label>合同金额</Label>
              <Input
                className="mt-1"
                type="number"
                step="0.01"
                value={form.amount}
                onChange={(e) => setForm({ ...form, amount: e.target.value })}
              />
            </div>
            <div>
              <Label>签订日期</Label>
              <Input
                className="mt-1"
                type="date"
                value={form.signDate}
                onChange={(e) => setForm({ ...form, signDate: e.target.value })}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <Label>生效日期</Label>
                <Input
                  className="mt-1"
                  type="date"
                  value={form.startDate}
                  onChange={(e) => setForm({ ...form, startDate: e.target.value })}
                />
              </div>
              <div>
                <Label>到期日期</Label>
                <Input
                  className="mt-1"
                  type="date"
                  value={form.endDate}
                  onChange={(e) => setForm({ ...form, endDate: e.target.value })}
                />
              </div>
            </div>
            <div>
              <Label>状态</Label>
              <NativeSelect
                className="mt-1"
                value={form.status}
                onChange={(e) => setForm({ ...form, status: e.target.value })}
              >
                {Object.entries(CONTRACT_STATUS).map(([k, v]) => (
                  <option key={k} value={k}>
                    {v.label}
                  </option>
                ))}
              </NativeSelect>
            </div>
            <div>
              <Label>备注</Label>
              <Textarea
                className="mt-1 min-h-[60px] resize-none"
                value={form.remarks}
                onChange={(e) => setForm({ ...form, remarks: e.target.value })}
                placeholder="合同备注说明"
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
