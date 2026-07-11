import { useCallback, useState } from 'react'
import type { BudgetNode, MemberBudgetQuota, UpdateMemberBudgetInput } from '@/api/types'
import { ApiError } from '@/api/client'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { displayToPoints, formatDisplayCurrency, pointsToDisplay } from '@/lib/points'
import { cn } from '@/lib/utils'
import { Users, Pencil, Check, X, Loader2, Search } from 'lucide-react'

interface BudgetEditMemberBudgetProps {
  node: BudgetNode
  onUpdated: () => void
  getMemberBudgets: (departmentId: string) => Promise<MemberBudgetQuota[]>
  updateMemberBudget: (
    memberId: string,
    data: UpdateMemberBudgetInput,
  ) => Promise<MemberBudgetQuota>
}

export function BudgetEditMemberBudget({
  node,
  onUpdated,
  getMemberBudgets,
  updateMemberBudget,
}: BudgetEditMemberBudgetProps) {
  const [averageDraft, setAverageDraft] = useState('')
  const [savingAverage, setSavingAverage] = useState(false)
  const [individualMode, setIndividualMode] = useState(false)
  const [members, setMembers] = useState<MemberBudgetQuota[]>([])
  const [loading, setLoading] = useState(false)
  const [search, setSearch] = useState('')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [draft, setDraft] = useState('')
  const [saving, setSaving] = useState(false)

  const handleIndividualModeChange = useCallback(
    (checked: boolean) => {
      setIndividualMode(checked)
      if (!checked) return
      setLoading(true)
      getMemberBudgets(node.id)
        .then((data) => setMembers(data ?? []))
        .catch((err) => toast.error(err instanceof ApiError ? err.message : '加载成员额度失败'))
        .finally(() => setLoading(false))
    },
    [getMemberBudgets, node.id],
  )

  async function handleSaveAverage() {
    const value = parseFloat(averageDraft)
    if (Number.isNaN(value) || value < 0) {
      toast.error('请输入有效的额度数值')
      return
    }
    setSavingAverage(true)
    try {
      const budgets = await getMemberBudgets(node.id)
      const points = displayToPoints(value)
      for (const member of budgets) {
        if (member.personalBudget !== points) {
          await updateMemberBudget(member.memberId, { personalBudget: points })
        }
      }
      onUpdated()
      toast.success(`已将所有成员额度设置为 ${formatDisplayCurrency(points)}/人`)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : '设置失败，请重试')
    } finally {
      setSavingAverage(false)
    }
  }

  const startEdit = useCallback((member: MemberBudgetQuota) => {
    setEditingId(member.memberId)
    setDraft(String(pointsToDisplay(member.personalBudget)))
  }, [])

  const cancelEdit = useCallback(() => {
    setEditingId(null)
    setDraft('')
  }, [])

  const handleSaveMember = useCallback(
    async (memberId: string) => {
      const value = parseFloat(draft)
      if (Number.isNaN(value) || value < 0) {
        toast.error('请输入有效的额度数值')
        return
      }
      setSaving(true)
      try {
        const updated = await updateMemberBudget(memberId, {
          personalBudget: displayToPoints(value),
        })
        setMembers((prev) =>
          prev.map((m) => (m.memberId === memberId ? updated : m)),
        )
        setEditingId(null)
        setDraft('')
        toast.success('成员额度已更新')
      } catch (err) {
        toast.error(err instanceof ApiError ? err.message : '修改失败，请重试')
      } finally {
        setSaving(false)
      }
    },
    [draft, updateMemberBudget],
  )

  const filteredMembers = search.trim()
    ? members.filter((m) => m.memberName.includes(search.trim()))
    : members

  return (
    <div className="rounded-lg border border-border p-4">
      <div className="mb-3 flex items-center gap-2">
        <Users className="size-4 text-muted-foreground" />
        <h4 className="text-sm font-semibold text-foreground">成员额度</h4>
      </div>

      {/* Average quota setting */}
      <div className="mb-4">
        <Label className="mb-1.5 block text-xs text-muted-foreground">平均额度/人（元）</Label>
        <div className="flex items-center gap-2">
          <Input
            type="number"
            min={0}
            value={averageDraft}
            onChange={(e) => setAverageDraft(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter') void handleSaveAverage() }}
            className="h-8 w-40 tabular-nums"
            placeholder="输入统一额度"
          />
          <Button
            size="sm"
            className="h-8"
            onClick={handleSaveAverage}
            disabled={savingAverage || !averageDraft.trim()}
          >
            {savingAverage ? '设置中…' : '应用到全部'}
          </Button>
        </div>
      </div>

      {/* Individual mode toggle */}
      <div className="flex items-center gap-2 border-t border-border pt-3">
        <Switch
          id="individual-mode"
          checked={individualMode}
          onCheckedChange={handleIndividualModeChange}
        />
        <Label htmlFor="individual-mode" className="cursor-pointer text-xs text-muted-foreground">
          单独配置成员额度
        </Label>
      </div>

      {/* Individual member list */}
      {individualMode && (
        <div className="mt-3">
          <div className="relative mb-2">
            <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="搜索成员..."
              className="h-8 pl-8 text-sm"
            />
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-6">
              <Loader2 className="size-4 animate-spin text-muted-foreground" />
            </div>
          ) : filteredMembers.length === 0 ? (
            <p className="py-4 text-center text-xs text-muted-foreground">
              {search.trim() ? '未找到匹配成员' : '暂无成员'}
            </p>
          ) : (
            <div className="max-h-64 overflow-y-auto">
              <table className="w-full text-sm">
                <thead className="sticky top-0 bg-background">
                  <tr className="border-b border-border text-left text-xs text-muted-foreground">
                    <th className="pb-2 font-medium">成员</th>
                    <th className="pb-2 font-medium">个人额度</th>
                    <th className="pb-2 font-medium">已用</th>
                    <th className="pb-2 text-right font-medium">操作</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {filteredMembers.map((member) => (
                    <tr key={member.memberId} className={cn('h-10', editingId === member.memberId && 'bg-muted/30')}>
                      <td className="py-2 font-medium text-foreground">{member.memberName}</td>
                      <td className="py-2 tabular-nums">
                        {editingId === member.memberId ? (
                          <Input
                            type="number"
                            min={0}
                            value={draft}
                            onChange={(e) => setDraft(e.target.value)}
                            onKeyDown={(e) => {
                              if (e.key === 'Enter') void handleSaveMember(member.memberId)
                              if (e.key === 'Escape') cancelEdit()
                            }}
                            className="h-7 w-28 tabular-nums"
                            placeholder="元"
                            autoFocus
                          />
                        ) : (
                          <span className="text-muted-foreground">
                            {formatDisplayCurrency(member.personalBudget)}
                          </span>
                        )}
                      </td>
                      <td className="py-2 tabular-nums text-muted-foreground">
                        {formatDisplayCurrency(member.used)}
                      </td>
                      <td className="py-2 text-right">
                        {editingId === member.memberId ? (
                          <div className="flex items-center justify-end gap-1">
                            <Button variant="ghost" size="icon" className="size-7" onClick={cancelEdit} disabled={saving} aria-label="取消">
                              <X className="size-3.5" />
                            </Button>
                            <Button variant="ghost" size="icon" className="size-7" onClick={() => void handleSaveMember(member.memberId)} disabled={saving} aria-label="保存">
                              {saving ? <Loader2 className="size-3.5 animate-spin" /> : <Check className="size-3.5" />}
                            </Button>
                          </div>
                        ) : (
                          <Button variant="ghost" size="icon" className="size-7 text-muted-foreground" onClick={() => startEdit(member)} aria-label={`编辑${member.memberName}的额度`}>
                            <Pencil className="size-3.5" />
                          </Button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
