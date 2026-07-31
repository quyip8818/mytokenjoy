import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useInjectedQuery } from '@/features/query'
import { useApis } from '@/api/use-apis'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { PasswordInput } from '@/components/ui/password-input'
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
import type { User } from '@/api/auth'
import type { Role } from '@/api/users'

export function UsersPage() {
  const apis = useApis()
  const [keyword, setKeyword] = useState('')
  const [roles, setRoles] = useState<Role[]>([])
  useEffect(() => {
    apis.usersApi.roles().then(setRoles)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const { data, loading, refresh } = useInjectedQuery({
    queryKey: ['users', keyword],
    queryFn: (a) => a.usersApi.list(),
  })

  const filtered =
    data?.filter((u) => !keyword || u.username.includes(keyword) || u.realName.includes(keyword)) ??
    []

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<User | null>(null)
  const [form, setForm] = useState({
    username: '',
    password: '',
    realName: '',
    email: '',
    role: 'viewer',
    status: 1,
  })
  const [saving, setSaving] = useState(false)
  const [confirmState, setConfirmState] = useState<ConfirmActionState | null>(null)

  const openCreate = () => {
    setEditing(null)
    setForm({ username: '', password: '', realName: '', email: '', role: 'viewer', status: 1 })
    setDialogOpen(true)
  }

  const openEdit = (u: User) => {
    setEditing(u)
    setForm({
      username: u.username,
      password: '',
      realName: u.realName,
      email: u.email ?? '',
      role: u.role,
      status: u.status,
    })
    setDialogOpen(true)
  }

  const handleSave = async () => {
    if (!form.username || !form.realName) {
      toast.error('用户名和姓名不能为空')
      return
    }
    if (!editing && !form.password) {
      toast.error('请输入密码')
      return
    }
    setSaving(true)
    try {
      if (editing) {
        await apis.usersApi.update(editing.id, {
          realName: form.realName,
          email: form.email || undefined,
          role: form.role,
          password: form.password || undefined,
          status: form.status,
        })
        toast.success('更新成功')
      } else {
        await apis.usersApi.create({
          username: form.username,
          password: form.password,
          realName: form.realName,
          email: form.email || undefined,
          role: form.role,
          status: form.status,
        })
        toast.success('创建成功')
      }
      setDialogOpen(false)
      refresh()
    } catch (e: any) {
      toast.error(e.message)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = (u: User) => {
    if (u.username === 'admin') {
      toast.error('不能删除管理员账户')
      return
    }
    setConfirmState({
      open: true,
      title: '确认删除',
      desc: `确定删除用户「${u.realName}（${u.username}）」吗？`,
      variant: 'danger',
      confirmLabel: '删除',
      onConfirm: async () => {
        try {
          await apis.usersApi.delete(u.id)
          toast.success('删除成功')
          refresh()
        } catch (e: any) {
          toast.error(e.message)
        }
        setConfirmState(null)
      },
    })
  }

  const roleLabel = (code: string) => roles.find((r) => r.code === code)?.name ?? code

  const roleBadgeVariant = (role: string) => {
    if (role === 'admin') return 'destructive' as const
    if (role === 'buyer') return 'default' as const
    return 'outline' as const
  }

  return (
    <PageShell>
      <PageHeader
        title="用户管理"
        actions={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" /> 新建用户
          </Button>
        }
      />

      <Input
        className="h-9 w-56"
        placeholder="用户名 / 姓名"
        value={keyword}
        onChange={(e) => setKeyword(e.target.value)}
      />

      <div className="rounded-lg border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>用户名</TableHead>
              <TableHead>姓名</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead>角色</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.map((u) => (
              <TableRow key={u.id}>
                <TableCell>{u.username}</TableCell>
                <TableCell>{u.realName}</TableCell>
                <TableCell className="text-muted-foreground">{u.email ?? '-'}</TableCell>
                <TableCell>
                  <Badge variant={roleBadgeVariant(u.role)}>{roleLabel(u.role)}</Badge>
                </TableCell>
                <TableCell>
                  <span
                    className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${u.status === 1 ? 'bg-green-50 text-green-700' : 'bg-gray-100 text-gray-500'}`}
                  >
                    {u.status === 1 ? '启用' : '停用'}
                  </span>
                </TableCell>
                <TableCell className="text-right">
                  <Button variant="ghost" size="icon-sm" onClick={() => openEdit(u)}>
                    <Pencil className="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => handleDelete(u)}
                    disabled={u.username === 'admin'}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
            {!loading && filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={6} className="py-12 text-center text-muted-foreground">
                  暂无数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* 新建/编辑弹窗 */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{editing ? `编辑用户（${editing.username}）` : '新建用户'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <Label>
                用户名 <span className="text-red-500">*</span>
              </Label>
              <Input
                className="mt-1"
                value={form.username}
                onChange={(e) => setForm({ ...form, username: e.target.value })}
                disabled={!!editing}
              />
            </div>
            <div>
              <Label>
                姓名 <span className="text-red-500">*</span>
              </Label>
              <Input
                className="mt-1"
                value={form.realName}
                onChange={(e) => setForm({ ...form, realName: e.target.value })}
              />
            </div>
            <div>
              <Label>
                {editing ? '重置密码' : '密码'}{' '}
                {!editing && <span className="text-red-500">*</span>}
              </Label>
              <PasswordInput
                className="mt-1"
                value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })}
                placeholder={editing ? '留空则不修改' : '至少 6 位'}
              />
            </div>
            <div>
              <Label>邮箱</Label>
              <Input
                className="mt-1"
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
              />
            </div>
            <div>
              <Label>
                角色 <span className="text-red-500">*</span>
              </Label>
              <NativeSelect
                className="mt-1"
                value={form.role}
                onChange={(e) => setForm({ ...form, role: e.target.value })}
              >
                {roles.map((r) => (
                  <option key={r.code} value={r.code}>
                    {r.name}
                  </option>
                ))}
              </NativeSelect>
            </div>
            <div>
              <Label>状态</Label>
              <NativeSelect
                className="mt-1"
                value={form.status}
                onChange={(e) => setForm({ ...form, status: Number(e.target.value) })}
              >
                <option value={1}>启用</option>
                <option value={0}>停用</option>
              </NativeSelect>
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
