import { useEffect } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import type { Permission, Role } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'

interface RoleFormData {
  name: string
  permissions: string[]
}

interface RoleFormProps {
  open: boolean
  role: Role | null // null = create mode
  permissions: Permission[]
  onSubmit: (data: RoleFormData) => void
  onCancel: () => void
}

export function RoleForm({ open, role, permissions, onSubmit, onCancel }: RoleFormProps) {
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    control,
    formState: { errors },
  } = useForm<RoleFormData>({
    defaultValues: { name: '', permissions: [] },
  })

  const watchedPermissions = useWatch({ control, name: 'permissions' })

  useEffect(() => {
    if (open) {
      reset({
        name: role?.name ?? '',
        permissions: role?.permissions ?? [],
      })
    }
  }, [open, role, reset])

  // Group permissions by group field
  const grouped = permissions.reduce<Record<string, Permission[]>>((acc, p) => {
    if (!acc[p.group]) acc[p.group] = []
    acc[p.group].push(p)
    return acc
  }, {})

  const togglePermission = (permId: string) => {
    const current = watchedPermissions || []
    const next = current.includes(permId)
      ? current.filter((id) => id !== permId)
      : [...current, permId]
    setValue('permissions', next, { shouldDirty: true })
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!o) onCancel()
      }}
    >
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit(onSubmit)}>
          <DialogHeader>
            <DialogTitle>{role ? '编辑角色' : '创建角色'}</DialogTitle>
          </DialogHeader>

          {/* Name field */}
          <div className="mt-5 mb-5">
            <Label htmlFor="role-name" className="text-xs font-medium text-muted-foreground">
              角色名称
            </Label>
            <Input
              id="role-name"
              {...register('name', { required: '请输入角色名称' })}
              placeholder="输入角色名称"
              className="mt-1.5"
            />
            {errors.name && <p className="mt-1 text-xs text-destructive">{errors.name.message}</p>}
          </div>

          {/* Permissions */}
          <div className="mb-5">
            <Label className="text-xs font-medium text-muted-foreground mb-3 block">权限配置</Label>
            <div className="max-h-64 overflow-y-auto border border-border rounded-md p-3 space-y-4">
              {Object.entries(grouped).map(([group, perms]) => (
                <div key={group}>
                  <p className="text-xs font-medium text-foreground mb-2">{group}</p>
                  <div className="space-y-2 pl-1">
                    {perms.map((perm) => (
                      <div key={perm.id} className="flex items-center gap-2">
                        <Checkbox
                          id={`perm-${perm.id}`}
                          checked={(watchedPermissions || []).includes(perm.id)}
                          onCheckedChange={() => togglePermission(perm.id)}
                        />
                        <Label
                          htmlFor={`perm-${perm.id}`}
                          className="text-sm font-normal cursor-pointer text-foreground"
                        >
                          {perm.name}
                        </Label>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onCancel}>
              取消
            </Button>
            <Button type="submit">{role ? '保存' : '创建'}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
