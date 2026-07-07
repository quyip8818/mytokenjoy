import { useEffect, useState } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { toast } from 'sonner'
import type { Member, Department } from '@/api/types'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowFormLayout } from '../components/workflow-form-layout'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useWorkflow } from '../use-workflow'
import { flattenDepartments } from '@/features/org'

interface MemberFormData {
  name: string
  phone: string
  email: string
  departmentId: string
}

export function MemberFormWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'member-form'>) {
  const { closeAll } = useWorkflow()
  const member = entry.payload.member as Member | null | undefined
  const departments = (entry.payload.departments as Department[]) ?? []
  const onSubmit = entry.payload.onSubmit as ((data: MemberFormData) => Promise<void>) | undefined
  const defaultDeptId = (entry.payload.defaultDeptId as string) ?? ''

  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors },
  } = useForm<MemberFormData>()

  useEffect(() => {
    if (member) {
      reset({
        name: member.name,
        phone: member.phone,
        email: member.email,
        departmentId: member.departmentId,
      })
    } else {
      reset({ name: '', phone: '', email: '', departmentId: defaultDeptId })
    }
  }, [member, defaultDeptId, reset])

  const flatDepts = flattenDepartments(departments)

  const onFormSubmit = async (data: MemberFormData) => {
    try {
      await onSubmit?.(data)
      toast.success(member ? '成员已更新' : '成员已添加')
      closeAll()
    } catch {
      toast.error('操作失败')
    }
  }

  return (
    <WorkflowPanelChrome
      title={member ? '编辑成员' : '添加成员'}
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel={member ? '保存' : '添加'}
          onPrimary={handleSubmit(onFormSubmit)}
        />
      }
    >
      <WorkflowFormLayout as="form" onChange={() => onSetDirty(true)}>
        <div className="space-y-1.5">
          <Label>姓名</Label>
          <Input {...register('name', { required: '请输入姓名' })} />
          {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label>手机号</Label>
          <Input {...register('phone', { required: '请输入手机号' })} />
          {errors.phone && <p className="text-xs text-destructive">{errors.phone.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label>邮箱</Label>
          <Input {...register('email', { required: '请输入邮箱' })} type="email" />
          {errors.email && <p className="text-xs text-destructive">{errors.email.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label>部门</Label>
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
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}

export function MemberInviteWorkflow({
  entry,
  onClose,
  onSetDirty,
}: WorkflowComponentProps<'member-invite'>) {
  const { closeAll } = useWorkflow()
  const onSubmit = entry.payload.onSubmit as ((value: string) => Promise<void>) | undefined
  const [value, setValue] = useState('')

  const handleSubmit = async () => {
    if (!value.trim()) return
    try {
      await onSubmit?.(value)
      toast.success('邀请已发送')
      closeAll()
    } catch {
      toast.error('邀请失败')
    }
  }

  return (
    <WorkflowPanelChrome
      title="邀请成员"
      onClose={onClose}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel="发送邀请"
          onPrimary={handleSubmit}
          primaryDisabled={!value.trim()}
        />
      }
    >
      <WorkflowFormLayout className="space-y-3">
        <Label>手机号或邮箱</Label>
        <Input
          value={value}
          onChange={(e) => {
            setValue(e.target.value)
            onSetDirty(true)
          }}
          placeholder="输入手机号或邮箱"
        />
      </WorkflowFormLayout>
    </WorkflowPanelChrome>
  )
}
