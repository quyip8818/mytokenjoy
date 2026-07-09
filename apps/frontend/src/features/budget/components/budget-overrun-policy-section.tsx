import { useState, useCallback } from 'react'
import { toast } from 'sonner'
import { ApiError } from '@/api/client'
import type { OverrunPolicyConfig } from '@/api/types'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { Textarea } from '@/components/ui/textarea'
import { Input } from '@/components/ui/input'
import { Pencil, X } from 'lucide-react'

interface BudgetOverrunPolicySectionProps {
  policy: OverrunPolicyConfig | undefined
  onUpdate: (data: OverrunPolicyConfig) => Promise<void>
}

const CHANNEL_LABELS = [
  { key: 'notifyEmail' as const, label: '邮件' },
  { key: 'notifyPhone' as const, label: '电话' },
  { key: 'notifyIm' as const, label: 'IM' },
]

export function BudgetOverrunPolicySection({
  policy,
  onUpdate,
}: BudgetOverrunPolicySectionProps) {
  const [editing, setEditing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState<OverrunPolicyConfig | null>(null)
  const [thresholdInput, setThresholdInput] = useState('')

  const startEdit = useCallback(() => {
    if (!policy) return
    setForm({ ...policy })
    setThresholdInput('')
    setEditing(true)
  }, [policy])

  const cancel = useCallback(() => {
    setEditing(false)
    setForm(null)
  }, [])

  const handleSave = useCallback(async () => {
    if (!form) return
    setSaving(true)
    try {
      await onUpdate(form)
      toast.success('超限策略已更新')
      setEditing(false)
      setForm(null)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '保存失败，请重试')
    } finally {
      setSaving(false)
    }
  }, [form, onUpdate])

  const addThreshold = useCallback(() => {
    const val = parseInt(thresholdInput, 10)
    if (!form || isNaN(val) || val <= 0 || val > 200) return
    if (form.thresholds.includes(val)) {
      setThresholdInput('')
      return
    }
    setForm({ ...form, thresholds: [...form.thresholds, val].sort((a, b) => a - b) })
    setThresholdInput('')
  }, [form, thresholdInput])

  const removeThreshold = useCallback(
    (val: number) => {
      if (!form) return
      setForm({ ...form, thresholds: form.thresholds.filter((t) => t !== val) })
    },
    [form],
  )

  if (!policy) return null

  const activeChannels = CHANNEL_LABELS.filter(({ key }) => policy[key])

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-3">
        <CardTitle className="text-sm font-medium">超限策略</CardTitle>
        {!editing && (
          <Button variant="ghost" size="sm" className="h-7 gap-1 px-2 text-xs" onClick={startEdit}>
            <Pencil className="size-3" />
            编辑
          </Button>
        )}
      </CardHeader>
      <CardContent>
        {!editing ? (
          <div className="space-y-3 text-sm">
            <div>
              <span className="text-muted-foreground">预警阈值：</span>
              <span className="ml-1.5 inline-flex flex-wrap gap-1">
                {policy.thresholds.length === 0 ? (
                  <span className="text-muted-foreground">未设置</span>
                ) : (
                  policy.thresholds.map((t) => (
                    <Badge key={t} variant="secondary" className="text-xs">
                      {t}%
                    </Badge>
                  ))
                )}
              </span>
            </div>
            <div>
              <span className="text-muted-foreground">通知渠道：</span>
              <span className="ml-1.5">
                {activeChannels.length === 0 ? (
                  <span className="text-muted-foreground">未启用</span>
                ) : (
                  activeChannels.map(({ label }) => label).join('、')
                )}
              </span>
            </div>
            {policy.blockMessage && (
              <div>
                <span className="text-muted-foreground">拦截提示：</span>
                <span className="ml-1.5 text-foreground">{policy.blockMessage}</span>
              </div>
            )}
          </div>
        ) : (
          <div className="space-y-4">
            <div>
              <label className="mb-1.5 block text-xs font-medium text-muted-foreground">
                预警阈值（%）
              </label>
              <div className="mb-2 flex flex-wrap gap-1">
                {form!.thresholds.map((t) => (
                  <Badge
                    key={t}
                    variant="secondary"
                    className="gap-0.5 pr-1 text-xs"
                  >
                    {t}%
                    <button
                      type="button"
                      className="ml-0.5 rounded-full p-0.5 hover:bg-muted-foreground/20"
                      onClick={() => removeThreshold(t)}
                      aria-label={`移除 ${t}%`}
                    >
                      <X className="size-2.5" />
                    </button>
                  </Badge>
                ))}
              </div>
              <div className="flex items-center gap-2">
                <Input
                  type="number"
                  min={1}
                  max={200}
                  placeholder="输入百分比，如 80"
                  className="h-8 w-32 text-xs"
                  value={thresholdInput}
                  onChange={(e) => setThresholdInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault()
                      addThreshold()
                    }
                  }}
                />
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-8 text-xs"
                  onClick={addThreshold}
                >
                  添加
                </Button>
              </div>
            </div>

            <div>
              <label className="mb-1.5 block text-xs font-medium text-muted-foreground">
                通知渠道
              </label>
              <div className="flex items-center gap-4">
                {CHANNEL_LABELS.map(({ key, label }) => (
                  <label key={key} className="flex items-center gap-1.5 text-sm">
                    <Checkbox
                      checked={form![key]}
                      onCheckedChange={(checked) =>
                        setForm((prev) => (prev ? { ...prev, [key]: !!checked } : null))
                      }
                    />
                    {label}
                  </label>
                ))}
              </div>
            </div>

            <div>
              <label className="mb-1.5 block text-xs font-medium text-muted-foreground">
                拦截提示（429 错误文案）
              </label>
              <Textarea
                className="min-h-[60px] resize-none text-xs"
                placeholder="自定义超限拦截时返回的消息…"
                value={form!.blockMessage}
                onChange={(e) =>
                  setForm((prev) => (prev ? { ...prev, blockMessage: e.target.value } : null))
                }
              />
            </div>

            <div className="flex items-center gap-2 pt-1">
              <Button size="sm" className="h-7 text-xs" onClick={handleSave} disabled={saving}>
                {saving ? '保存中…' : '保存'}
              </Button>
              <Button
                size="sm"
                variant="ghost"
                className="h-7 text-xs text-muted-foreground"
                onClick={cancel}
                disabled={saving}
              >
                取消
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
