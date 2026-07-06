import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import type { SyncConfig } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'
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
import { CheckCircle2, Loader2 } from 'lucide-react'

interface StepSyncScheduleProps {
  onComplete: () => void
  onBack: () => void
}

export function StepSyncSchedule({ onComplete, onBack }: StepSyncScheduleProps) {
  const apis = useInjectedApis()
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)

  const { register, setValue, watch, handleSubmit } = useForm<SyncConfig>({
    defaultValues: {
      enabled: false,
      startTime: '02:00',
      frequencyHours: 24,
      deleteMemberThreshold: 10,
      deleteDepartmentThreshold: 3,
      notifyPhone: true,
      notifyEmail: false,
    },
  })

  const enabled = watch('enabled')
  const frequencyHours = watch('frequencyHours')
  const notifyPhone = watch('notifyPhone')
  const notifyEmail = watch('notifyEmail')

  useEffect(() => {
    void apis.syncApi.getConfig().then((config) => {
      const fields: (keyof SyncConfig)[] = [
        'enabled', 'startTime', 'frequencyHours',
        'deleteMemberThreshold', 'deleteDepartmentThreshold',
        'notifyPhone', 'notifyEmail',
      ]
      fields.forEach((key) => {
        setValue(key, config[key] as never)
      })
    })
  }, [apis, setValue])

  const onSubmit = async (data: SyncConfig) => {
    setSaving(true)
    try {
      await apis.syncApi.saveConfig(data)
      setSaved(true)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-sm font-semibold">定时同步配置</h3>
        <p className="text-sm text-muted-foreground mt-1">
          配置自动同步策略，系统将按计划定期从数据源拉取最新数据
        </p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-5 max-w-lg">
        {/* Enable toggle */}
        <div className="flex items-center gap-3">
          <Switch
            checked={enabled}
            onCheckedChange={(checked) => setValue('enabled', !!checked)}
          />
          <Label className="text-sm font-medium">
            {enabled ? '自动同步已开启' : '自动同步已关闭'}
          </Label>
        </div>

        {enabled && (
          <div className="space-y-4 rounded-lg border bg-muted/20 p-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>每日开始时间</Label>
                <Input type="time" {...register('startTime')} />
              </div>
              <div className="space-y-1.5">
                <Label>同步频率</Label>
                <Select
                  value={String(frequencyHours)}
                  onValueChange={(val) =>
                    setValue('frequencyHours', Number(val) as 6 | 12 | 24)
                  }
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
              <div className="space-y-1.5">
                <Label>删除成员保护阈值</Label>
                <Input
                  type="number"
                  {...register('deleteMemberThreshold', {
                    valueAsNumber: true,
                    min: 0,
                  })}
                />
                <p className="text-xs text-muted-foreground">
                  单次删除超过此数将暂停
                </p>
              </div>
              <div className="space-y-1.5">
                <Label>删除部门保护阈值</Label>
                <Input
                  type="number"
                  {...register('deleteDepartmentThreshold', {
                    valueAsNumber: true,
                    min: 0,
                  })}
                />
                <p className="text-xs text-muted-foreground">
                  单次删除超过此数将暂停
                </p>
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
                  <Label htmlFor="notifyPhone" className="text-sm cursor-pointer">
                    手机短信
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
              </div>
            </div>
          </div>
        )}

        {/* Save feedback */}
        {saved && (
          <div className="flex items-center gap-2 text-sm text-emerald-700 bg-emerald-50 border border-emerald-200 rounded-md px-4 py-2.5">
            <CheckCircle2 className="size-4" />
            配置已保存
          </div>
        )}

        {/* Actions */}
        <div className="flex items-center gap-3 pt-2">
          <Button type="button" variant="outline" onClick={onBack}>
            上一步
          </Button>
          {saved ? (
            <Button type="button" onClick={onComplete}>
              完成配置
            </Button>
          ) : (
            <Button type="submit" disabled={saving}>
              {saving && <Loader2 className="size-4 animate-spin" />}
              保存并完成
            </Button>
          )}
        </div>
      </form>
    </div>
  )
}
