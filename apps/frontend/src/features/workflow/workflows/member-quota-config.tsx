import { useMemo, useState } from 'react'
import type { MemberBudgetQuota } from '@/api/types'
import { ApiError } from '@/api/client'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import { queryKeys, useInjectedQuery } from '@/features/query'
import type { WorkflowComponentProps } from '../types'
import { WorkflowPanelChrome, WorkflowPanelFooter } from '../components/workflow-panel-chrome'
import { WorkflowInfoBox } from '../components/workflow-info-box'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { cn } from '@/lib/utils'

export function MemberQuotaConfigWorkflow({
  entry,
  onClose,
  onSetDirty,
  injectedApis,
}: WorkflowComponentProps<'member-quota-config'> & { injectedApis?: AppApis }) {
  const apis = useInjectedApis(injectedApis)
  const departmentId = entry.payload.departmentId
  const departmentName = entry.payload.departmentName
  const onSuccess = entry.payload.onSuccess

  const {
    data: quotas = [],
    loading,
    refresh,
  } = useInjectedQuery({
    injectedApis: apis,
    queryKey: queryKeys.budget.memberQuotas(departmentId),
    queryFn: (a) => a.budgetApi.getMemberQuotas(departmentId),
  })

  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [draftQuota, setDraftQuota] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const effectiveSelectedId = selectedId ?? quotas[0]?.memberId ?? null

  const selected = useMemo(
    () => quotas.find((q) => q.memberId === effectiveSelectedId) ?? null,
    [quotas, effectiveSelectedId],
  )

  const draftValue = selectedId
    ? draftQuota
    : selected
      ? String(selected.personalQuota)
      : draftQuota

  const handleSelect = (item: MemberBudgetQuota) => {
    setSelectedId(item.memberId)
    setDraftQuota(String(item.personalQuota))
    setError('')
  }

  const handleSave = async () => {
    if (!effectiveSelectedId || !draftValue.trim()) return
    setSaving(true)
    setError('')
    try {
      const updated = await apis.budgetApi.updateMemberQuota(effectiveSelectedId, {
        personalQuota: Number(draftValue),
      })
      setSelectedId(updated.memberId)
      setDraftQuota(String(updated.personalQuota))
      onSetDirty(false)
      onSuccess?.()
      refresh()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <WorkflowPanelChrome
      title="成员额度配置"
      onClose={onClose}
      contextBar={departmentName}
      footer={
        <WorkflowPanelFooter
          onCancel={onClose}
          primaryLabel="保存"
          onPrimary={() => void handleSave()}
          primaryDisabled={!effectiveSelectedId || !draftValue.trim() || saving || !!error}
        />
      }
    >
      <div className="grid grid-cols-5 gap-8">
        <div className="col-span-3 space-y-4">
          {loading ? (
            <p className="text-sm text-muted-foreground">加载中…</p>
          ) : quotas.length === 0 ? (
            <p className="text-sm text-muted-foreground">该部门暂无成员</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow className="hover:bg-transparent">
                  <TableHead>成员</TableHead>
                  <TableHead className="text-right">个人额度</TableHead>
                  <TableHead className="text-right">已分配 Key</TableHead>
                  <TableHead className="text-right">已用</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {quotas.map((item) => (
                  <TableRow
                    key={item.memberId}
                    className={cn(
                      'cursor-pointer',
                      effectiveSelectedId === item.memberId && 'bg-blue-50/80',
                    )}
                    onClick={() => handleSelect(item)}
                  >
                    <TableCell className="font-medium">{item.memberName}</TableCell>
                    <TableCell className="text-right">
                      ¥{item.personalQuota.toLocaleString()}
                    </TableCell>
                    <TableCell className="text-right">¥{item.allocated.toLocaleString()}</TableCell>
                    <TableCell className="text-right">¥{item.used.toLocaleString()}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>
        <WorkflowInfoBox fullWidth className="space-y-4">
          <h4 className="font-semibold text-foreground/80">编辑额度</h4>
          {selected ? (
            <>
              <p className="text-sm text-muted-foreground">{selected.memberName}</p>
              <div className="space-y-1.5">
                <Label>个人额度 (¥)</Label>
                <Input
                  type="number"
                  value={draftValue}
                  onChange={(e) => {
                    setSelectedId(selected.memberId)
                    setDraftQuota(e.target.value)
                    onSetDirty(true)
                    setError('')
                  }}
                />
              </div>
              <p className="text-xs text-muted-foreground">
                已分配 Key：¥{selected.allocated.toLocaleString()}（额度不可低于此值）
              </p>
            </>
          ) : (
            <p className="text-sm text-muted-foreground">选择成员以编辑额度</p>
          )}
          {error && (
            <div className="rounded-md bg-red-50 border border-red-200 px-3 py-2 text-sm text-red-700">
              {error}
            </div>
          )}
        </WorkflowInfoBox>
      </div>
    </WorkflowPanelChrome>
  )
}
