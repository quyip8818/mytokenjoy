import { useEffect } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { toast } from 'sonner'
import type { SyncConfig as SyncConfigType } from '@/api/types'
import { useApis } from '@/api/use-apis'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface SyncConfigProps {
  onTriggerSync: () => void
  triggeringSync: boolean
  onSaved?: () => void
}

export function SyncConfigPanel({ onTriggerSync, triggeringSync, onSaved }: SyncConfigProps) {
  const apis = useApis()
  const { register, handleSubmit, setValue, control } = useForm<SyncConfigType>({
    defaultValues: {
      enabled: false,
      startTime: '02:00',
      frequencyHours: 24,
      deleteMemberThreshold: 10,
      deleteDepartmentThreshold: 3,
      notifyPhone: true,
      notifyEmail: false,
      notifyIm: false,
    },
  })

  const enabled = useWatch({ control, name: 'enabled' })
  const frequencyHours = useWatch({ control, name: 'frequencyHours' })
  const notifyPhone = useWatch({ control, name: 'notifyPhone' })
  const notifyEmail = useWatch({ control, name: 'notifyEmail' })
  const notifyIm = useWatch({ control, name: 'notifyIm' })

  useEffect(() => {
    apis.syncApi.getConfig().then((config) => {
      const fields: (keyof SyncConfigType)[] = [
        'enabled',
        'startTime',
        'frequencyHours',
        'deleteMemberThreshold',
        'deleteDepartmentThreshold',
        'notifyPhone',
        'notifyEmail',
        'notifyIm',
      ]
      fields.forEach((key) => {
        setValue(key, config[key] as never)
      })
    })
  }, [apis, setValue])

  const onSubmit = async (data: SyncConfigType) => {
    await apis.syncApi.saveConfig(data)
    toast.success('同步策略已保存')
    onSaved?.()
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* Enable toggle */}
      <div className="flex items-center gap-3">
        <Switch checked={enabled} onCheckedChange={(checked) => setValue('enabled', checked)} />
        <Label className="text-sm font-medium">
          {enabled ? '自动同步已开启' : '自动同步已关闭'}
        </Label>
      </div>

      {enabled && (
        <div className="space-y-4 pl-1">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label className="mb-1 block">每日开始时间</Label>
              <Input type="time" {...register('startTime')} />
            </div>
            <div>
              <Label className="mb-1 block">同步频率</Label>
              <Select
                value={String(frequencyHours)}
                onValueChange={(val) => setValue('frequencyHours', Number(val) as 6 | 12 | 24)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="选择频率" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="6">每 6 小时</SelectItem>
                  <SelectItem value="12">每 12 小时</SelectItem>
                  <SelectItem value="24">每 24 小时</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label className="mb-1 block">删除成员保护阈值</Label>
              <Input
                type="number"
                {...register('deleteMemberThreshold', { valueAsNumber: true, min: 0 })}
              />
              <p className="text-xs text-muted-foreground mt-1">单次同步删除成员超过此数将暂停</p>
            </div>
            <div>
              <Label className="mb-1 block">删除部门保护阈值</Label>
              <Input
                type="number"
                {...register('deleteDepartmentThreshold', { valueAsNumber: true, min: 0 })}
              />
              <p className="text-xs text-muted-foreground mt-1">单次同步删除部门超过此数将暂停</p>
            </div>
          </div>

          <div>
            <Label className="mb-2 block">通知方式</Label>
            <div className="flex gap-4">
              <div className="flex items-center gap-2">
                <Checkbox
                  id="notifyPhone"
                  checked={notifyPhone}
                  onCheckedChange={(checked) => setValue('notifyPhone', !!checked)}
                />
                <Label htmlFor="notifyPhone" className="text-sm cursor-pointer">
                  手机号
                </Label>
              </div>
              <div className="flex items-center gap-2">
                <Checkbox
                  id="notifyEmail"
                  checked={notifyEmail}
                  onCheckedChange={(checked) => setValue('notifyEmail', !!checked)}
                />
                <Label htmlFor="notifyEmail" className="text-sm cursor-pointer">
                  邮箱
                </Label>
              </div>
              <div className="flex items-center gap-2">
                <Checkbox
                  id="notifyIm"
                  checked={notifyIm}
                  onCheckedChange={(checked) => setValue('notifyIm', !!checked)}
                />
                <Label htmlFor="notifyIm" className="text-sm cursor-pointer">
                  IM
                </Label>
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="flex gap-3 pt-2">
        <Button type="submit">保存配置</Button>
        <Button type="button" variant="outline" onClick={onTriggerSync} disabled={triggeringSync}>
          {triggeringSync ? '同步中...' : '立即同步一次'}
        </Button>
      </div>
    </form>
  )
}
