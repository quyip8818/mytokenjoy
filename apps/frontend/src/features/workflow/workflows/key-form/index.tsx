import { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import type { Member, PlatformKeyScope } from '@/api/types'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import { useSession } from '@/features/session'
import { useMemberWhitelist } from '@/features/workflow/hooks/use-member-whitelist'
import type { WorkflowComponentProps, WorkflowStackEntry } from '@/features/workflow/types'
import {
  WorkflowPanelChrome,
  WorkflowPanelFooter,
} from '@/features/workflow/components/workflow-panel-chrome'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { useWorkflow } from '@/features/workflow/hooks/use-workflow'
import { workflowErrorMessage } from '@/features/workflow/lib/error-message'
import { BUDGET_INSUFFICIENT_MESSAGE } from '@/features/keys'
import { InlineModelPicker } from '@/features/models'
import { formatBudgetContext, useKeyFormBudget, useKeyFormState } from './use-key-form-budget'
import { useBillingExchange } from '@/features/session'
import { currencySymbol, formatMoney } from '@/lib/quota-display'

type KeyFormWorkflowProps = WorkflowComponentProps<'key-create' | 'key-edit'> & {
  injectedApis?: AppApis
}

export function KeyFormWorkflow({
  entry,
  onPush,
  onClose,
  onSetDirty,
  injectedApis,
}: KeyFormWorkflowProps) {
  const apis = useInjectedApis(injectedApis)
  const { closeAll } = useWorkflow()
  const { memberId, companyType } = useSession()
  const { billingCurrency } = useBillingExchange()
  const currencyLabel = currencySymbol(billingCurrency)
  const { resolveAllowedModelIds: resolveAllModels } = useMemberWhitelist()
  const isTrialOrDemo = companyType === 'trial' || companyType === 'demo'

  const isCreate = entry.id === 'key-create'
  const key =
    entry.id === 'key-edit' ? (entry as WorkflowStackEntry<'key-edit'>).payload.key : undefined
  const createPayload =
    entry.id === 'key-create' ? (entry as WorkflowStackEntry<'key-create'>).payload : undefined
  const adminCreate = Boolean(createPayload?.adminCreate)
  const projectId = createPayload?.projectId
  const projectName = createPayload?.projectName
  const onSuccess = entry.payload.onSuccess
  const scope: PlatformKeyScope = createPayload?.scope ?? 'member'

  const {
    name,
    setName,
    budget,
    setBudget,
    models,
    setModels,
    targetMemberId,
    setTargetMemberId,
    targetMemberName,
    setTargetMemberName,
    submitting,
    setSubmitting,
  } = useKeyFormState({
    key,
    adminCreate,
    defaultMemberId: memberId,
    initialTargetMemberId: createPayload?.targetMemberId,
    initialName: createPayload?.initialName,
    initialBudget: createPayload?.initialBudget,
  })

  const effectiveMemberId = adminCreate || scope === 'project_member' ? targetMemberId : memberId
  const requiresMemberPick = adminCreate || scope === 'project_member'

  // Resolve allowed model IDs for the picker
  const [allowedModelIds, setAllowedModelIds] = useState<string[]>([])
  useEffect(() => {
    let cancelled = false
    const resolve = async () => {
      if (isTrialOrDemo) {
        const allModels = await apis.modelApi.list()
        return allModels.filter((m) => m.type.startsWith('test-')).map((m) => m.modelId)
      }
      return resolveAllModels()
    }
    void resolve().then((ids) => {
      if (!cancelled && ids) setAllowedModelIds(ids)
    })
    return () => { cancelled = true }
  }, [isTrialOrDemo, resolveAllModels, apis.modelApi])

  const {
    budgetSummary,
    projectBudgetRemaining,
    subBudgetRemaining,
    budgetAmount,
    budgetInsufficient,
    budgetExceedsRemaining,
    projectBudgetExceeds,
    subBudgetExceeds,
  } = useKeyFormBudget({
    isCreate,
    scope,
    effectiveMemberId,
    projectId,
    budget,
    adminCreate,
    injectedApis: apis,
  })

  // Default budget to user's remaining quota once loaded
  useEffect(() => {
    if (isCreate && !budget && budgetSummary && budgetSummary.remaining > 0) {
      setBudget(String(budgetSummary.remaining))
    }
  }, [isCreate, budget, budgetSummary, setBudget])

  const openPickMember = () => {
    onPush('member-search', {
      multi: false,
      excludeIds: [],
      onConfirm: (members: Member[]) => {
        const picked = members[0]
        if (!picked) return
        setTargetMemberId(picked.id)
        setTargetMemberName(picked.alias)
        onSetDirty(true)
      },
    })
  }

  const handleModelsChange = (ids: string[]) => {
    setModels(ids)
    onSetDirty(true)
  }

  const handleCreate = async () => {
    if (budgetInsufficient) {
      toast.error(BUDGET_INSUFFICIENT_MESSAGE)
      return
    }
    if (budgetSummary && budgetAmount > budgetSummary.remaining) {
      toast.error(`额度不能超过剩余 ${formatMoney(budgetSummary.remaining)}`)
      return
    }
    if (projectBudgetExceeds) {
      toast.error(`额度不能超过项目剩余 ${formatMoney(projectBudgetRemaining!)}`)
      return
    }
    if (subBudgetExceeds) {
      toast.error(`额度不能超过成员子额度剩余 ${formatMoney(subBudgetRemaining!)}`)
      return
    }
    setSubmitting(true)
    try {
      const created = await apis.platformKeyApi.create({
        name,
        scope,
        memberId:
          scope === 'member' || scope === 'project_member'
            ? effectiveMemberId || memberId
            : scope === 'project' && adminCreate
              ? effectiveMemberId || undefined
              : undefined,
        projectId: scope === 'project' || scope === 'project_member' ? projectId : undefined,
        budget: budgetAmount,
        modelWhitelist: models,
      })
      toast.success('Key 创建成功')
      onSuccess?.(created.id)
      if (!created.fullKey) {
        toast.error('创建失败：未返回 Key')
        return
      }
      onPush('key-reveal', {
        fullKey: created.fullKey,
        onDone: onSuccess,
      })
    } catch (err) {
      toast.error(workflowErrorMessage(err, '创建失败'))
    } finally {
      setSubmitting(false)
    }
  }

  const handleSave = async () => {
    if (!key) return
    setSubmitting(true)
    try {
      await apis.platformKeyApi.update(key.id, {
        name,
        budget: Number(budget) || 0,
        modelWhitelist: models,
      })
      toast.success('Key 已更新')
      onSuccess?.()
      closeAll()
    } catch (err) {
      toast.error(workflowErrorMessage(err, '保存失败'))
    } finally {
      setSubmitting(false)
    }
  }

  const contextBar = (() => {
    if (!isCreate) return undefined
    if (scope === 'project_member') {
      return `项目：${projectName ?? ''} · 成员：${targetMemberName || '—'} · 子额度剩余 ${formatMoney(subBudgetRemaining ?? 0)}`
    }
    if (scope === 'project') {
      return `项目：${projectName ?? ''} · 剩余可分配 ${formatMoney(projectBudgetRemaining ?? 0)}`
    }
    return formatBudgetContext(
      budgetSummary,
      adminCreate ? targetMemberName || undefined : undefined,
    )
  })()

  return (
    <WorkflowPanelChrome
      title={isCreate ? '创建 Key' : '编辑 Key'}
      onClose={onClose}
      contextBar={contextBar}
      banner={
        budgetInsufficient ? (
          <p className="text-sm text-amber-800">{BUDGET_INSUFFICIENT_MESSAGE}</p>
        ) : budgetExceedsRemaining ? (
          <p className="text-sm text-amber-800">
            申请额度超过剩余 {formatMoney(budgetSummary!.remaining)}
          </p>
        ) : projectBudgetExceeds ? (
          <p className="text-sm text-amber-800">
            申请额度超过项目剩余 {formatMoney(projectBudgetRemaining!)}
          </p>
        ) : subBudgetExceeds ? (
          <p className="text-sm text-amber-800">
            申请额度超过成员子额度剩余 {formatMoney(subBudgetRemaining!)}
          </p>
        ) : undefined
      }
      footer={
        isCreate ? (
          <WorkflowPanelFooter
            onCancel={onClose}
            primaryLabel={submitting ? '创建中...' : '创建 Key'}
            onPrimary={handleCreate}
            primaryDisabled={
              submitting ||
              !name.trim() ||
              (requiresMemberPick && !targetMemberId) ||
              budgetInsufficient ||
              budgetExceedsRemaining ||
              projectBudgetExceeds ||
              subBudgetExceeds
            }
          />
        ) : (
          <WorkflowPanelFooter
            onCancel={onClose}
            primaryLabel={submitting ? '保存中...' : '保存'}
            onPrimary={handleSave}
            primaryDisabled={submitting || !name.trim()}
          />
        )
      }
    >
      <div className="space-y-5">
        {requiresMemberPick && (
          <div className="space-y-1.5">
            <Label>绑定成员</Label>
            <Button
              variant="outline"
              className="w-full justify-start"
              onClick={openPickMember}
            >
              {targetMemberName || '选择成员'}
            </Button>
          </div>
        )}
        <div className="space-y-1.5">
          <Label>Key 名称</Label>
          <Input
            value={name}
            onChange={(e) => {
              setName(e.target.value)
              onSetDirty(true)
            }}
            placeholder="如：开发调试"
          />
        </div>
        <div className="space-y-1.5">
          <Label>额度 ({currencyLabel})</Label>
          <Input
            type="number"
            value={budget}
            onChange={(e) => {
              setBudget(e.target.value)
              onSetDirty(true)
            }}
          />
        </div>

        <InlineModelPicker
          value={models}
          onChange={handleModelsChange}
          allowedModelIds={allowedModelIds.length > 0 ? allowedModelIds : undefined}
          injectedApis={apis}
          hint="不选 = 全部可用"
        />
      </div>
    </WorkflowPanelChrome>
  )
}
