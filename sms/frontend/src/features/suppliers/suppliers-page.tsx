import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { Plus, Trash2, Pencil, Eye } from 'lucide-react'
import { toast } from 'sonner'
import { useFilteredQuery, useInjectedMutation, queryKeys } from '@/features/query'
import { useSession } from '@/features/session'
import { useApis } from '@/api/use-apis'
import { SUPPLIER_STATUS, CATEGORIES } from '@/config/enums'
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
  const [confirmState, setConfirmState] = useState<ConfirmActionState | null>(null)

  const deleteMut = useInjectedMutation<void, string>({
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
    setConfirmState({
      open: true,
      title: '确认删除',
      desc: `确定删除供应商「${row.name}」吗？`,
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
        title="供应商管理"
        actions={
          canEdit ? (
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" /> 新建供应商
            </Button>
          ) : undefined
        }
      />

      <div className="flex items-center gap-2">
        <Input
          className="h-9 w-48"
          placeholder="名称 / 编码"
          value={filter.keyword}
          onChange={(e) => search({ keyword: e.target.value })}
        />
        <NativeSelect
          className="h-9 w-auto"
          value={filter.status}
          onChange={(e) => search({ status: e.target.value })}
        >
          <option value="">全部状态</option>
          {Object.entries(SUPPLIER_STATUS).map(([k, v]) => (
            <option key={k} value={k}>
              {v.label}
            </option>
          ))}
        </NativeSelect>
        <NativeSelect
          className="h-9 w-auto"
          value={filter.category}
          onChange={(e) => search({ category: e.target.value })}
        >
          <option value="">全部分类</option>
          {CATEGORIES.map((c) => (
            <option key={c} value={c}>
              {c}
            </option>
          ))}
        </NativeSelect>
      </div>

      <div className="rounded-lg border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>厂商名称</TableHead>
              <TableHead>编码</TableHead>
              <TableHead>分类</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>官网</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data?.items.map((s) => (
              <TableRow key={s.id}>
                <TableCell>
                  <Link
                    to="/suppliers/$id"
                    params={{ id: s.id }}
                    className="text-primary hover:underline"
                  >
                    {s.name}
                  </Link>
                </TableCell>
                <TableCell className="text-muted-foreground">{s.code}</TableCell>
                <TableCell>{s.category ?? '-'}</TableCell>
                <TableCell>
                  <StatusBadge status={s.status} map={SUPPLIER_STATUS} />
                </TableCell>
                <TableCell className="max-w-[180px] truncate text-muted-foreground">
                  {s.website ?? '-'}
                </TableCell>
                <TableCell className="text-right">
                  <Button variant="ghost" size="icon-sm" asChild>
                    <Link to="/suppliers/$id" params={{ id: s.id }}>
                      <Eye className="h-3.5 w-3.5" />
                    </Link>
                  </Button>
                  {canEdit && (
                    <>
                      <Button variant="ghost" size="icon-sm" onClick={() => openEdit(s)}>
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      <Button variant="ghost" size="icon-sm" onClick={() => handleDelete(s)}>
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </>
                  )}
                </TableCell>
              </TableRow>
            ))}
            {!loading && data?.items.length === 0 && (
              <TableRow>
                <TableCell colSpan={6} className="py-12 text-center text-muted-foreground">
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
            <DialogTitle>{editing ? '编辑供应商' : '新建供应商'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <Label>
                厂商名称 <span className="text-red-500">*</span>
              </Label>
              <Input
                className="mt-1"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="如：OpenAI"
              />
            </div>
            <div>
              <Label>
                厂商编码 <span className="text-red-500">*</span>
              </Label>
              <Input
                className="mt-1"
                value={form.code}
                onChange={(e) => setForm({ ...form, code: e.target.value })}
                placeholder="如：OPENAI"
                disabled={!!editing}
              />
            </div>
            <div>
              <Label>分类</Label>
              <div className="mt-1 flex gap-3">
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
            </div>
            <div>
              <Label>状态</Label>
              <NativeSelect
                className="mt-1"
                value={form.status}
                onChange={(e) => setForm({ ...form, status: e.target.value })}
              >
                {Object.entries(SUPPLIER_STATUS).map(([k, v]) => (
                  <option key={k} value={k}>
                    {v.label}
                  </option>
                ))}
              </NativeSelect>
            </div>
            <div>
              <Label>官网</Label>
              <Input
                className="mt-1"
                value={form.website}
                onChange={(e) => setForm({ ...form, website: e.target.value })}
                placeholder="https://"
              />
            </div>
            <div>
              <Label>备注说明</Label>
              <Textarea
                className="mt-1 min-h-[60px] resize-none"
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
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
