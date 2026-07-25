import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useInjectedQuery } from '@/features/query'
import { useApis } from '@/api/use-apis'
import { Badge, Field } from '@/components/ui'
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

  const handleDelete = async (u: User) => {
    if (u.username === 'admin') {
      toast.error('不能删除管理员账户')
      return
    }
    if (!confirm(`确定删除用户「${u.realName}（${u.username}）」吗？`)) return
    try {
      await apis.usersApi.delete(u.id)
      toast.success('删除成功')
      refresh()
    } catch (e: any) {
      toast.error(e.message)
    }
  }

  const roleLabel = (code: string) => roles.find((r) => r.code === code)?.name ?? code

  const roleBadgeVariant = (role: string) => {
    if (role === 'admin') return 'destructive' as const
    if (role === 'buyer') return 'default' as const
    return 'outline' as const
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <input
          className="h-9 w-56 rounded-md border px-3 text-sm"
          placeholder="用户名 / 姓名"
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
        />
        <button
          onClick={openCreate}
          className="inline-flex h-9 items-center gap-1.5 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground"
        >
          <Plus className="h-4 w-4" /> 新建用户
        </button>
      </div>

      <div className="rounded-lg border bg-white">
        <table className="w-full text-sm">
          <thead className="border-b bg-muted/40">
            <tr>
              <th className="px-4 py-3 text-left font-medium">用户名</th>
              <th className="px-4 py-3 text-left font-medium">姓名</th>
              <th className="px-4 py-3 text-left font-medium">邮箱</th>
              <th className="px-4 py-3 text-left font-medium">角色</th>
              <th className="px-4 py-3 text-left font-medium">状态</th>
              <th className="px-4 py-3 text-right font-medium">操作</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((u) => (
              <tr key={u.id} className="border-b last:border-0 hover:bg-muted/20">
                <td className="px-4 py-3">{u.username}</td>
                <td className="px-4 py-3">{u.realName}</td>
                <td className="px-4 py-3 text-muted-foreground">{u.email ?? '-'}</td>
                <td className="px-4 py-3">
                  <Badge variant={roleBadgeVariant(u.role)}>{roleLabel(u.role)}</Badge>
                </td>
                <td className="px-4 py-3">
                  <span
                    className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${u.status === 1 ? 'bg-green-50 text-green-700' : 'bg-gray-100 text-gray-500'}`}
                  >
                    {u.status === 1 ? '启用' : '停用'}
                  </span>
                </td>
                <td className="px-4 py-3 text-right">
                  <button
                    onClick={() => openEdit(u)}
                    className="px-1.5 py-1 text-muted-foreground hover:text-primary"
                  >
                    <Pencil className="inline h-3.5 w-3.5" />
                  </button>
                  <button
                    onClick={() => handleDelete(u)}
                    disabled={u.username === 'admin'}
                    className="px-1.5 py-1 text-muted-foreground hover:text-red-500 disabled:opacity-30"
                  >
                    <Trash2 className="inline h-3.5 w-3.5" />
                  </button>
                </td>
              </tr>
            ))}
            {!loading && filtered.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-12 text-center text-muted-foreground">
                  暂无数据
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {dialogOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          onClick={() => setDialogOpen(false)}
        >
          <div
            className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="mb-4 text-lg font-semibold">
              {editing ? `编辑用户（${editing.username}）` : '新建用户'}
            </h2>
            <div className="space-y-3">
              <Field label="用户名" required>
                <input
                  className="input"
                  value={form.username}
                  onChange={(e) => setForm({ ...form, username: e.target.value })}
                  disabled={!!editing}
                />
              </Field>
              <Field label="姓名" required>
                <input
                  className="input"
                  value={form.realName}
                  onChange={(e) => setForm({ ...form, realName: e.target.value })}
                />
              </Field>
              <Field label={editing ? '重置密码' : '密码'} required={!editing}>
                <input
                  className="input"
                  type="password"
                  value={form.password}
                  onChange={(e) => setForm({ ...form, password: e.target.value })}
                  placeholder={editing ? '留空则不修改' : '至少 6 位'}
                />
              </Field>
              <Field label="邮箱">
                <input
                  className="input"
                  value={form.email}
                  onChange={(e) => setForm({ ...form, email: e.target.value })}
                />
              </Field>
              <Field label="角色" required>
                <select
                  className="input"
                  value={form.role}
                  onChange={(e) => setForm({ ...form, role: e.target.value })}
                >
                  {roles.map((r) => (
                    <option key={r.code} value={r.code}>
                      {r.name}
                    </option>
                  ))}
                </select>
              </Field>
              <Field label="状态">
                <select
                  className="input"
                  value={form.status}
                  onChange={(e) => setForm({ ...form, status: Number(e.target.value) })}
                >
                  <option value={1}>启用</option>
                  <option value={0}>停用</option>
                </select>
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
