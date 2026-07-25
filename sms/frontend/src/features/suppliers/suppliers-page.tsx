import { useState } from 'react'
import { Link } from 'react-router'
import { Plus, Trash2, Pencil, Eye } from 'lucide-react'
import { toast } from 'sonner'
import { useFilteredQuery, useInjectedMutation, queryKeys } from '@/features/query'
import { useSession } from '@/features/session'
import { useApis } from '@/api/use-apis'
import { SUPPLIER_STATUS, CATEGORIES } from '@/config/enums'
import { StatusBadge, Field, Pagination } from '@/components/ui'
import type { Supplier } from '@/api/suppliers'

export function SuppliersPage() {
  const { user } = useSession()
  const canEdit = user?.role === 'admin' || user?.role === 'buyer'
  const apis = useApis()

  const { data, loading, filter, setFilter, search } = useFilteredQuery({
    initialFilter: { page: 1, pageSize: 10, keyword: '', status: '', category: '' },
    queryKeyFactory: (f) => queryKeys.suppliers.list(f),
    fetcher: (a, f) => a.suppliersApi.list(f),
  })

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<Supplier | null>(null)
  const [form, setForm] = useState({
    name: '',
    code: '',
    category: '',
    website: '',
    status: 'potential',
    description: '',
  })
  const [saving, setSaving] = useState(false)

  const deleteMut = useInjectedMutation<void, number>({
    mutationFn: (a, id) => a.suppliersApi.delete(id),
    invalidateKeys: [queryKeys.suppliers.all],
    onSuccess: () => toast.success('删除成功'),
  })

  const openCreate = () => {
    setEditing(null)
    setForm({ name: '', code: '', category: '', website: '', status: 'potential', description: '' })
    setDialogOpen(true)
  }

  const openEdit = (row: Supplier) => {
    setEditing(row)
    setForm({
      name: row.name,
      code: row.code,
      category: row.category ?? '',
      website: row.website ?? '',
      status: row.status,
      description: row.description ?? '',
    })
    setDialogOpen(true)
  }

  const handleSave = async () => {
    if (!form.name || !form.code) {
      toast.error('名称和编码不能为空')
      return
    }
    setSaving(true)
    try {
      if (editing) {
        await apis.suppliersApi.update(editing.id, form)
        toast.success('更新成功')
      } else {
        await apis.suppliersApi.create(form)
        toast.success('创建成功')
      }
      setDialogOpen(false)
      search({})
    } catch (e: any) {
      toast.error(e.message || '操作失败')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = (row: Supplier) => {
    if (!confirm(`确定删除供应商「${row.name}」吗？`)) return
    deleteMut.mutate(row.id)
  }

  const totalPages = Math.ceil((data?.total ?? 0) / filter.pageSize)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <input
            className="h-9 w-48 rounded-md border px-3 text-sm"
            placeholder="名称 / 编码"
            value={filter.keyword}
            onChange={(e) => search({ keyword: e.target.value })}
          />
          <select
            className="h-9 rounded-md border px-2 text-sm"
            value={filter.status}
            onChange={(e) => search({ status: e.target.value })}
          >
            <option value="">全部状态</option>
            {Object.entries(SUPPLIER_STATUS).map(([k, v]) => (
              <option key={k} value={k}>
                {v.label}
              </option>
            ))}
          </select>
          <select
            className="h-9 rounded-md border px-2 text-sm"
            value={filter.category}
            onChange={(e) => search({ category: e.target.value })}
          >
            <option value="">全部分类</option>
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>
                {c}
              </option>
            ))}
          </select>
        </div>
        {canEdit && (
          <button
            onClick={openCreate}
            className="inline-flex h-9 items-center gap-1.5 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            <Plus className="h-4 w-4" /> 新建供应商
          </button>
        )}
      </div>

      <div className="rounded-lg border bg-white">
        <table className="w-full text-sm">
          <thead className="border-b bg-muted/40">
            <tr>
              <th className="px-4 py-3 text-left font-medium">厂商名称</th>
              <th className="px-4 py-3 text-left font-medium">编码</th>
              <th className="px-4 py-3 text-left font-medium">分类</th>
              <th className="px-4 py-3 text-left font-medium">状态</th>
              <th className="px-4 py-3 text-left font-medium">官网</th>
              <th className="px-4 py-3 text-right font-medium">操作</th>
            </tr>
          </thead>
          <tbody>
            {data?.items.map((s) => (
              <tr key={s.id} className="border-b last:border-0 hover:bg-muted/20">
                <td className="px-4 py-3">
                  <Link to={`/suppliers/${s.id}`} className="text-primary hover:underline">
                    {s.name}
                  </Link>
                </td>
                <td className="px-4 py-3 text-muted-foreground">{s.code}</td>
                <td className="px-4 py-3">{s.category ?? '-'}</td>
                <td className="px-4 py-3">
                  <StatusBadge status={s.status} map={SUPPLIER_STATUS} />
                </td>
                <td className="px-4 py-3 max-w-[180px] truncate text-muted-foreground">
                  {s.website ?? '-'}
                </td>
                <td className="px-4 py-3 text-right">
                  <Link
                    to={`/suppliers/${s.id}`}
                    className="inline-flex items-center gap-1 px-2 py-1 text-xs text-muted-foreground hover:text-primary"
                  >
                    <Eye className="h-3.5 w-3.5" />
                  </Link>
                  {canEdit && (
                    <>
                      <button
                        onClick={() => openEdit(s)}
                        className="inline-flex items-center gap-1 px-2 py-1 text-xs text-muted-foreground hover:text-primary"
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </button>
                      <button
                        onClick={() => handleDelete(s)}
                        className="inline-flex items-center gap-1 px-2 py-1 text-xs text-muted-foreground hover:text-red-500"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </>
                  )}
                </td>
              </tr>
            ))}
            {!loading && data?.items.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-12 text-center text-muted-foreground">
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
            <h2 className="mb-4 text-lg font-semibold">{editing ? '编辑供应商' : '新建供应商'}</h2>
            <div className="space-y-3">
              <Field label="厂商名称" required>
                <input
                  className="input"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="如：OpenAI"
                />
              </Field>
              <Field label="厂商编码" required>
                <input
                  className="input"
                  value={form.code}
                  onChange={(e) => setForm({ ...form, code: e.target.value })}
                  placeholder="如：OPENAI"
                  disabled={!!editing}
                />
              </Field>
              <Field label="分类">
                <div className="flex gap-3">
                  {CATEGORIES.map((c) => (
                    <label key={c} className="flex items-center gap-1.5 text-sm">
                      <input
                        type="radio"
                        name="category"
                        value={c}
                        checked={form.category === c}
                        onChange={() => setForm({ ...form, category: c })}
                      />
                      {c}
                    </label>
                  ))}
                </div>
              </Field>
              <Field label="状态">
                <select
                  className="input"
                  value={form.status}
                  onChange={(e) => setForm({ ...form, status: e.target.value })}
                >
                  {Object.entries(SUPPLIER_STATUS).map(([k, v]) => (
                    <option key={k} value={k}>
                      {v.label}
                    </option>
                  ))}
                </select>
              </Field>
              <Field label="官网">
                <input
                  className="input"
                  value={form.website}
                  onChange={(e) => setForm({ ...form, website: e.target.value })}
                  placeholder="https://"
                />
              </Field>
              <Field label="备注说明">
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
