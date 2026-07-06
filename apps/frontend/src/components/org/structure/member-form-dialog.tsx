import { useEffect } from 'react'
import { useForm, Controller } from 'react-hook-form'
import type { Member, Department } from '@/api/types'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface MemberFormData {
  name: string
  phone: string
  email: string
  username: string
  employeeId: string
  jobTitle: string
  hireDate: string
  departmentId: string
}

interface MemberFormDialogProps {
  open: boolean
  member?: Member | null
  departments: Department[]
  onSubmit: (data: MemberFormData) => void
  onClose: () => void
}

function flattenDepartments(
  departments: Department[],
  level = 0
): { id: string; name: string; level: number }[] {
  const result: { id: string; name: string; level: number }[] = []
  for (const dept of departments) {
    result.push({ id: dept.id, name: dept.name, level })
    if (dept.children) {
      result.push(...flattenDepartments(dept.children, level + 1))
    }
  }
  return result
}

export function MemberFormDialog({
  open,
  member,
  departments,
  onSubmit,
  onClose,
}: MemberFormDialogProps) {
  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors },
  } = useForm<MemberFormData>()

  useEffect(() => {
    if (open && member) {
      reset({
        name: member.name,
        phone: member.phone,
        email: member.email,
        username: member.username,
        employeeId: member.employeeId,
        jobTitle: member.jobTitle,
        hireDate: member.hireDate,
        departmentId: member.departmentId,
      })
    } else if (open) {
      reset({
        name: '', phone: '', email: '', username: '',
        employeeId: '', jobTitle: '', hireDate: '', departmentId: '',
      })
    }
  }, [open, member, reset])

  const flatDepts = flattenDepartments(departments)

  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose() }}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{member ? '编辑成员' : '添加成员'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>姓名 <span className="text-destructive">*</span></Label>
              <Input {...register('name', { required: '请输入姓名' })} />
              {errors.name && (
                <p className="text-xs text-destructive">{errors.name.message}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label>主部门 <span className="text-destructive">*</span></Label>
              <Controller
                name="departmentId"
                control={control}
                rules={{ required: '请选择部门' }}
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="请选择部门" />
                    </SelectTrigger>
                    <SelectContent>
                      {flatDepts.map((d) => (
                        <SelectItem key={d.id} value={d.id}>
                          {'　'.repeat(d.level)}
                          {d.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              />
              {errors.departmentId && (
                <p className="text-xs text-destructive">{errors.departmentId.message}</p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>手机号 <span className="text-destructive">*</span></Label>
              <Input {...register('phone', { required: '请输入手机号' })} />
              {errors.phone && (
                <p className="text-xs text-destructive">{errors.phone.message}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label>邮箱 <span className="text-destructive">*</span></Label>
              <Input
                type="email"
                {...register('email', { required: '请输入邮箱' })}
              />
              {errors.email && (
                <p className="text-xs text-destructive">{errors.email.message}</p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>用户名</Label>
              <Input placeholder="登录用户名" {...register('username')} />
            </div>
            <div className="space-y-1.5">
              <Label>工号</Label>
              <Input placeholder="员工工号" {...register('employeeId')} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label>职位</Label>
              <Input placeholder="如：高级工程师" {...register('jobTitle')} />
            </div>
            <div className="space-y-1.5">
              <Label>入职时间</Label>
              <Input type="date" {...register('hireDate')} />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              取消
            </Button>
            <Button type="submit">{member ? '保存' : '添加'}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
