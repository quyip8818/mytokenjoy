import { useEffect, useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { toast } from 'sonner'
import type { SyncConfig } from '@/api/types'
import type { AppApis } from '@/api/app-apis'
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
import { Loader2 } from 'lucide-react'

interface StepSyncScheduleProps {
  syncApi: AppApis['syncApi']
  onComplete: () => void
  onBack: () => void
}

export function StepSyncSchedule({ syncApi, onComplete, onBack }: StepSyncScheduleProps) {
  const [saving, setSaving] = useState(false)

  const { register, setValue, control, handleSubmit } = useForm<SyncConfig>({
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

  const [enabled, frequencyHours, notifyPhone, notifyEmail, notifyIm] = useWatch({
    control,
    name: ['enabled', 'frequencyHours', 'notifyPhone', 'notifyEmail', 'notifyIm'],
  })

  useEffect(() => {
    void syncApi.getConfig().then((config) => {
      const fields: (keyof SyncConfig)[] = [
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
  }, [syncApi, setValue])

  const onSubmit = async (data: SyncConfig) => {
    setSaving(true)
    try {
      await syncApi.saveConfig(data)
      toast.success('同步配置已保存，数据源配置完成')
      onComplete()
    } catch {
      toast.error('保存同步配置失败，请稍后重试')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-semibold">定时同步配置</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          配置自动同步策略，系统将按计划定期从数据源拉取最新数据
        </p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="max-w-lg space-y-5">
        <div className="flex items-center justify-between rounded-lg border p-4">
          <div>
            <Label htmlFor="sync-enabled" className="text-sm font-medium">
              自动同步
            </Label>
            <p className="mt-0.5 text-xs text-muted-foreground">
              {enabled ? '系统将按以下计划自动同步组织数据' : '关闭时仅支持手动触发同步'}
            </p>
          </div>
          <Switch
            id="sync-enabled"
            checked={enabled}
            onCheckedChange={(checked) => setValue('enabled', !!checked)}
          />
        </div>

        {enabled && (
          <div className="space-y-4 rounded-lg border bg-muted/20 p-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label htmlFor="sync-start-time">每日开始时间</Label>
                <Input id="sync-start-time" type="time" {...register('startTime')} />
              </div>
              <div className="space-y-1.5">
                <Label>同步频率</Label>
                <Select
                  value={String(frequencyHours)}
                  onValueChange={(val) => setValue('frequencyHours', Number(val) as 6 | 12 | 24)}
                >
                  <SelectTrigger className="w-full">
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
              <div className="space-y-1.5">
                <Label htmlFor="delete-member-threshold">删除成员保护阈值</Label>
                <Input
                  id="delete-member-threshold"
                  type="number"
                  min={0}
                  {...register('deleteMemberThreshold', {
                    valueAsNumber: true,
                    min: 0,
                  })}
                />
                <p className="text-xs text-muted-foreground">单次删除超过此数将暂停同步</p>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="delete-department-threshold">删除部门保护阈值</Label>
                <Input
                  id="delete-department-threshold"
                  type="number"
                  min={0}
                  {...register('deleteDepartmentThreshold', {
                    valueAsNumber: true,
                    min: 0,
                  })}
                />
                <p className="text-xs text-muted-foreground">单次删除超过此数将暂停同步</p>
              </div>
            </div>

            <div className="space-y-2">
              <Label>异常通知方式</Label>
              <div className="flex gap-4">
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="notifyPhone"
                    checked={notifyPhone}
                    onCheckedChange={(checked) => setValue('notifyPhone', !!checked)}
                  />
                  <Label htmlFor="notifyPhone" className="cursor-pointer text-sm font-normal">
                    手机短信
                  </Label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="notifyEmail"
                    checked={notifyEmail}
                    onCheckedChange={(checked) => setValue('notifyEmail', !!checked)}
                  />
                  <Label htmlFor="notifyEmail" className="cursor-pointer text-sm font-normal">
                    邮箱
                  </Label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="notifyIm"
                    checked={notifyIm}
                    onCheckedChange={(checked) => setValue('notifyIm', !!checked)}
                  />
                  <Label htmlFor="notifyIm" className="cursor-pointer text-sm font-normal">
                    IM 通知
                  </Label>
                </div>
              </div>
            </div>
          </div>
        )}

        <div className="flex items-center gap-3 border-t pt-4">
          <Button type="button" variant="outline" onClick={onBack}>
            上一步
          </Button>
          <Button type="submit" disabled={saving}>
            {saving && <Loader2 className="size-4 animate-spin" />}
            {saving ? '保存中...' : '保存并完成'}
          </Button>
        </div>
      </form>
    </div>
  )
}
