import { useState } from 'react'
import { Plus, Pencil, Trash2, Upload, Download, FileText } from 'lucide-react'
import { toast } from 'sonner'
import { useFilteredQuery, useInjectedMutation, queryKeys } from '@/features/query'
import { useSession } from '@/features/session'
import { useSupplierOptions } from '@/features/suppliers'
import { useApis } from '@/api/use-apis'
import { CONTRACT_STATUS } from '@/config/enums'
import { StatusBadge, Field, Pagination } from '@/components/ui'
import { PageShell } from '@/components/layout/page-shell'
import { PageHeader } from '@/components/layout/page-header'
import type { Contract, ContractDetail, ContractAttachment } from '@/api/contracts'

function daysUntil(endDate?: string): number | null {
  if (!endDate) return null
  const end = new Date(endDate).getTime()
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  return Math.ceil((end - now.getTime()) / (24 * 3600 * 1000))
}

function formatAmount(amount?: number): string {
  if (amount === undefined || amount === null) return '-'
  return Number(amount).toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

export function ContractsPage() {
  const { user } = useSession()
  const canEdit = user?.role === 'admin' || user?.role === 'buyer'
  const apis = useApis()
  const suppliers = useSupplierOptions()

  const { data, loading, filter, setFilter, search } = useFilteredQuery({
    initialFilter: { page: 1, pageSize: 10, keyword: '', supplierId: 0, status: '' },
    queryKeyFactory: (f) => queryKeys.contracts.list(f),
    fetcher: (a, f) => a.contractsApi.list(f),
  })

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<Contract | null>(null)
  const [form, setForm] = useState({
    contractNo: '',
    supplierId: 0,
    title: '',
    amount: '',
    signDate: '',
    startDate: '',
    endDate: '',
    status: 'draft',
    remarks: '',
  })
  const [saving, setSaving] = useState(false)

  const [detailOpen, setDetailOpen] = useState(false)
  const [detail, setDetail] = useState<ContractDetail | null>(null)

  const deleteMut = useInjectedMutation<void, number>({
    mutationFn: (a, id) => a.contractsApi.delete(id),
    invalidateKeys: [queryKeys.contracts.all],
    onSuccess: () => toast.success('删除成功'),
  })

  const openCreate = () => {
    setEditing(null)
    setForm({
      contractNo: '',
      supplierId: 0,
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
    if (!confirm(`确定删除合同「${row.title}」吗？`)) return
    deleteMut.mutate(row.id)
  }

  const openDetail = async (row: Contract) => {
    const d = await apis.contractsApi.detail(row.id)
    setDetail(d)
    setDetailOpen(true)
  }

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file || !detail) return
    try {
      await apis.contractsApi.uploadAttachment(detail.id, file)
      toast.success('上传成功')
      const d = await apis.contractsApi.detail(detail.id)
      setDetail(d)
      search({})
    } catch (err: any) {
      toast.error(err.message)
    }
    e.target.value = ''
  }

  const handleDeleteAttachment = async (att: ContractAttachment) => {
    if (!detail || !confirm(`确定删除附件「${att.fileName}」吗？`)) return
    await apis.contractsApi.deleteAttachment(detail.id, att.id)
    toast.success('删除成功')
    const d = await apis.contractsApi.detail(detail.id)
    setDetail(d)
  }

  const totalPages = Math.ceil((data?.total ?? 0) / filter.pageSize)

  return (
    <PageShell>
      <PageHeader
        title="合同管理"
        actions={
          canEdit ? (
            <button
              onClick={openCreate}
              className="inline-flex h-9 items-center gap-1.5 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            >
              <Plus className="h-4 w-4" /> 新建合同
            </button>
          ) : undefined
        }
      />

      <div className="flex items-center gap-2">
          <input
            className="h-9 w-48 rounded-md border px-3 text-sm"
            placeholder="合同编号 / 标题"
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
            {Object.entries(CONTRACT_STATUS).map(([k, v]) => (
              <option key={k} value={k}>
                {v.label}
              </option>
            ))}
          </select>
      </div>

      <div className="rounded-lg border bg-white">
        <table className="w-full text-sm">
          <thead className="border-b bg-muted/40">
            <tr>
              <th className="px-4 py-3 text-left font-medium">合同编号</th>
              <th className="px-4 py-3 text-left font-medium">标题</th>
              <th className="px-4 py-3 text-left font-medium">供应商</th>
              <th className="px-4 py-3 text-right font-medium">金额</th>
              <th className="px-4 py-3 text-left font-medium">到期日</th>
              <th className="px-4 py-3 text-right font-medium">剩余</th>
              <th className="px-4 py-3 text-left font-medium">状态</th>
              <th className="px-4 py-3 text-right font-medium">操作</th>
            </tr>
          </thead>
          <tbody>
            {data?.items.map((c) => {
              const days = daysUntil(c.endDate)
              return (
                <tr key={c.id} className="border-b last:border-0 hover:bg-muted/20">
                  <td className="px-4 py-3 text-muted-foreground">{c.contractNo}</td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => openDetail(c)}
                      className="text-left text-primary hover:underline"
                    >
                      {c.title}
                    </button>
                  </td>
                  <td className="px-4 py-3">{c.supplierName ?? '-'}</td>
                  <td className="px-4 py-3 text-right">{formatAmount(c.amount)}</td>
                  <td className="px-4 py-3 text-muted-foreground">{c.endDate ?? '-'}</td>
                  <td
                    className={`px-4 py-3 text-right text-xs font-medium ${days === null ? '' : days < 0 ? 'text-red-500' : days <= 30 ? 'text-yellow-600' : 'text-muted-foreground'}`}
                  >
                    {days === null ? '-' : days < 0 ? '已过期' : `${days} 天`}
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge status={c.status} map={CONTRACT_STATUS} />
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => openDetail(c)}
                      className="px-1.5 py-1 text-xs text-muted-foreground hover:text-primary"
                    >
                      <FileText className="inline h-3.5 w-3.5" />
                    </button>
                    {canEdit && (
                      <>
                        <button
                          onClick={() => openEdit(c)}
                          className="px-1.5 py-1 text-xs text-muted-foreground hover:text-primary"
                        >
                          <Pencil className="inline h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => handleDelete(c)}
                          className="px-1.5 py-1 text-xs text-muted-foreground hover:text-red-500"
                        >
                          <Trash2 className="inline h-3.5 w-3.5" />
                        </button>
                      </>
                    )}
                  </td>
                </tr>
              )
            })}
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

      {/* 新建/编辑弹窗 */}
      {dialogOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          onClick={() => setDialogOpen(false)}
        >
          <div
            className="w-full max-w-lg rounded-lg bg-white p-6 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="mb-4 text-lg font-semibold">{editing ? '编辑合同' : '新建合同'}</h2>
            <div className="space-y-3">
              <Field label="合同编号" required>
                <input
                  className="input"
                  value={form.contractNo}
                  onChange={(e) => setForm({ ...form, contractNo: e.target.value })}
                  disabled={!!editing}
                  placeholder="如：HT-2026-001"
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
              <Field label="合同标题" required>
                <input
                  className="input"
                  value={form.title}
                  onChange={(e) => setForm({ ...form, title: e.target.value })}
                />
              </Field>
              <Field label="合同金额">
                <input
                  className="input"
                  type="number"
                  step="0.01"
                  value={form.amount}
                  onChange={(e) => setForm({ ...form, amount: e.target.value })}
                />
              </Field>
              <Field label="签订日期">
                <input
                  className="input"
                  type="date"
                  value={form.signDate}
                  onChange={(e) => setForm({ ...form, signDate: e.target.value })}
                />
              </Field>
              <div className="grid grid-cols-2 gap-3">
                <Field label="生效日期">
                  <input
                    className="input"
                    type="date"
                    value={form.startDate}
                    onChange={(e) => setForm({ ...form, startDate: e.target.value })}
                  />
                </Field>
                <Field label="到期日期">
                  <input
                    className="input"
                    type="date"
                    value={form.endDate}
                    onChange={(e) => setForm({ ...form, endDate: e.target.value })}
                  />
                </Field>
              </div>
              <Field label="状态">
                <select
                  className="input"
                  value={form.status}
                  onChange={(e) => setForm({ ...form, status: e.target.value })}
                >
                  {Object.entries(CONTRACT_STATUS).map(([k, v]) => (
                    <option key={k} value={k}>
                      {v.label}
                    </option>
                  ))}
                </select>
              </Field>
              <Field label="备注">
                <textarea
                  className="input min-h-[60px] resize-none"
                  value={form.remarks}
                  onChange={(e) => setForm({ ...form, remarks: e.target.value })}
                  placeholder="合同备注说明"
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

      {/* 详情侧抽屉 */}
      {detailOpen && detail && (
        <div
          className="fixed inset-0 z-50 flex justify-end bg-black/30"
          onClick={() => setDetailOpen(false)}
        >
          <div
            className="h-full w-full max-w-lg overflow-auto bg-white p-6 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-lg font-semibold">{detail.title}</h2>
              <button
                onClick={() => setDetailOpen(false)}
                className="text-lg text-muted-foreground hover:text-foreground"
              >
                &times;
              </button>
            </div>
            <div className="mb-6 grid grid-cols-2 gap-3 text-sm">
              <div>
                <span className="text-muted-foreground">编号：</span>
                {detail.contractNo}
              </div>
              <div>
                <span className="text-muted-foreground">供应商：</span>
                {detail.supplierName}
              </div>
              <div>
                <span className="text-muted-foreground">金额：</span>
                {formatAmount(detail.amount)}
              </div>
              <div>
                <span className="text-muted-foreground">状态：</span>
                {CONTRACT_STATUS[detail.status]?.label}
              </div>
              <div>
                <span className="text-muted-foreground">签订：</span>
                {detail.signDate ?? '-'}
              </div>
              <div>
                <span className="text-muted-foreground">生效：</span>
                {detail.startDate ?? '-'}
              </div>
              <div>
                <span className="text-muted-foreground">到期：</span>
                {detail.endDate ?? '-'}
              </div>
              <div>
                <span className="text-muted-foreground">备注：</span>
                {detail.remarks ?? '-'}
              </div>
            </div>

            <div>
              <div className="mb-3 flex items-center justify-between">
                <h3 className="text-sm font-semibold">附件 ({detail.attachments.length})</h3>
                {canEdit && (
                  <label className="inline-flex cursor-pointer items-center gap-1 rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground">
                    <Upload className="h-3.5 w-3.5" /> 上传
                    <input type="file" className="hidden" onChange={handleUpload} />
                  </label>
                )}
              </div>
              {detail.attachments.length === 0 ? (
                <div className="py-8 text-center text-sm text-muted-foreground">暂无附件</div>
              ) : (
                <div className="space-y-2">
                  {detail.attachments.map((att) => (
                    <div
                      key={att.id}
                      className="flex items-center justify-between rounded border p-2.5"
                    >
                      <div className="flex min-w-0 items-center gap-2">
                        <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                        <div className="truncate text-sm">{att.fileName}</div>
                        <span className="text-xs text-muted-foreground">
                          {(att.fileSize / 1024).toFixed(0)} KB
                        </span>
                      </div>
                      <div className="flex items-center gap-1">
                        <a
                          href={`/api/contracts/${detail.id}/attachments/${att.id}/download`}
                          className="p-1 text-muted-foreground hover:text-primary"
                        >
                          <Download className="h-3.5 w-3.5" />
                        </a>
                        {canEdit && (
                          <button
                            onClick={() => handleDeleteAttachment(att)}
                            className="p-1 text-muted-foreground hover:text-red-500"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </PageShell>
  )
}
