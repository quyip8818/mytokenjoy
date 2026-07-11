import { useCallback, useState } from 'react'
import type { BudgetNode, MemberBudgetQuota, UpdateMemberBudgetInput } from '@/api/types'
import { ApiError } from '@/api/client'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { displayToPoints, formatDisplayCurrency, pointsToDisplay } from '@/lib/points'
import { cn } from '@/lib/utils'
import { Users, Pencil, Check, X, Loader2, Search } from 'lucide-react'
import { useAsyncFetch, useMemberBudgetQuotas } from '@/features/budget'

const emptyMemberBudgets: MemberBudgetQuota[] = []

interface BudgetEditMemberBudgetProps {
  node: BudgetNode
  onUpdated: () => void
  getMemberBudgets: (departmentId: string) => Promise<MemberBudgetQuota[]>
  updateMemberBudget: (
    memberId: string,
    data: UpdateMemberBudgetInput,
  ) => Promise<MemberBudgetQuota>
  applyAverageBudget: (
    departmentId: string,
    data: { personalBudget: number; recursive: boolean },
  ) => Promise<void>
}

export function BudgetEditMemberBudget({
  node,
  onUpdated,
  getMemberBudgets,
  updateMemberBudget,
  applyAverageBudget,
}: BudgetEditMemberBudgetProps) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const {
    loading: loadingDisplay,
    data: members,
    refresh,
  } = useMemberBudgetQuotas(node.id, getMemberBudgets)

  const averageBudget =
    members.length > 0 ? members.reduce((sum, m) => sum + m.personalBudget, 0) / members.length : 0

  return (
    <div className="rounded-lg border border-border p-4">
      <div className="mb-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Users className="size-4 text-muted-foreground" />
          <h4 className="text-sm font-semibold text-foreground">成员额度</h4>
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 gap-1.5 text-xs text-muted-foreground"
          onClick={() => setDialogOpen(true)}
        >
          <Pencil className="size-3.5" />
          编辑
        </Button>
      </div>

      <div className="flex items-center gap-3 rounded-md bg-muted/50 px-3 py-2.5">
        <Users className="size-4 shrink-0 text-muted-foreground" />
        <div className="min-w-0 flex-1">
          <p className="text-xs text-muted-foreground">人均额度</p>
          <p className="text-sm font-medium tabular-nums">
            {loadingDisplay ? '—' : formatDisplayCurrency(averageBudget)}
          </p>
        </div>
      </div>

      <MemberBudgetEditDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        departmentId={node.id}
        getMemberBudgets={getMemberBudgets}
        updateMemberBudget={updateMemberBudget}
        applyAverageBudget={applyAverageBudget}
        onUpdated={() => {
          onUpdated()
          void refresh()
        }}
      />
    </div>
  )
}

/* --- MemberBudgetEditDialog --- */

interface MemberBudgetEditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  departmentId: string
  getMemberBudgets: (departmentId: string) => Promise<MemberBudgetQuota[]>
  updateMemberBudget: (
    memberId: string,
    data: UpdateMemberBudgetInput,
  ) => Promise<MemberBudgetQuota>
  applyAverageBudget: (
    departmentId: string,
    data: { personalBudget: number; recursive: boolean },
  ) => Promise<void>
  onUpdated: () => void
}

function MemberBudgetEditDialog({
  open,
  onOpenChange,
  departmentId,
  getMemberBudgets,
  updateMemberBudget,
  applyAverageBudget,
  onUpdated,
}: MemberBudgetEditDialogProps) {
  const [averageDraft, setAverageDraft] = useState('')
  const [savingAverage, setSavingAverage] = useState(false)
  const [individualMode, setIndividualMode] = useState(false)
  const [search, setSearch] = useState('')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [draft, setDraft] = useState('')
  const [saving, setSaving] = useState(false)

  const fetchMembers = useCallback(
    () =>
      getMemberBudgets(departmentId)
        .then((data) => data ?? [])
        .catch((err) => {
          toast.error(err instanceof ApiError ? err.message : '加载成员额度失败')
          return []
        }),
    [departmentId, getMemberBudgets],
  )

  const fetchKey = open && individualMode ? `${departmentId}:${individualMode}` : ''
  const {
    loading,
    data: members,
    replace: replaceMembers,
  } = useAsyncFetch(fetchKey, fetchMembers, open && individualMode, emptyMemberBudgets)

  function handleClose() {
    setAverageDraft('')
    setIndividualMode(false)
    setSearch('')
    setEditingId(null)
    setDraft('')
    onOpenChange(false)
  }

  async function handleSaveAverage() {
    const value = parseFloat(averageDraft)
    if (Number.isNaN(value) || value < 0) {
      toast.error('请输入有效的额度数值')
      return
    }
    setSavingAverage(true)
    try {
      const points = displayToPoints(value)
      await applyAverageBudget(departmentId, { personalBudget: points, recursive: true })
      onUpdated()
      toast.success(`已将所有成员额度设置为 ${formatDisplayCurrency(points)}/人`)
      handleClose()
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
        replaceMembers(members.map((m) => (m.memberId === memberId ? updated : m)))
        setEditingId(null)
        setDraft('')
        toast.success('成员额度已更新')
        onUpdated()
      } catch (err) {
        toast.error(err instanceof ApiError ? err.message : '修改失败，请重试')
      } finally {
        setSaving(false)
      }
    },
    [draft, members, onUpdated, replaceMembers, updateMemberBudget],
  )

  const filteredMembers = search.trim()
    ? members.filter((m) => m.memberName.includes(search.trim()))
    : members

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!v) handleClose()
        else onOpenChange(v)
      }}
    >
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>成员额度设置</DialogTitle>
        </DialogHeader>

        {/* Average quota setting */}
        <div className="mb-4">
          <Label className="mb-1.5 block text-xs text-muted-foreground">
            统一设置人均额度（元）
          </Label>
          <div className="flex items-center gap-2">
            <Input
              type="number"
              min={0}
              value={averageDraft}
              onChange={(e) => setAverageDraft(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') void handleSaveAverage()
              }}
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
            onCheckedChange={setIndividualMode}
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
                      <tr
                        key={member.memberId}
                        className={cn('h-10', editingId === member.memberId && 'bg-muted/30')}
                      >
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
                              <Button
                                variant="ghost"
                                size="icon"
                                className="size-7"
                                onClick={cancelEdit}
                                disabled={saving}
                                aria-label="取消"
                              >
                                <X className="size-3.5" />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="size-7"
                                onClick={() => void handleSaveMember(member.memberId)}
                                disabled={saving}
                                aria-label="保存"
                              >
                                {saving ? (
                                  <Loader2 className="size-3.5 animate-spin" />
                                ) : (
                                  <Check className="size-3.5" />
                                )}
                              </Button>
                            </div>
                          ) : (
                            <Button
                              variant="ghost"
                              size="icon"
                              className="size-7 text-muted-foreground"
                              onClick={() => startEdit(member)}
                              aria-label={`编辑${member.memberName}的额度`}
                            >
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
      </DialogContent>
    </Dialog>
  )
}
