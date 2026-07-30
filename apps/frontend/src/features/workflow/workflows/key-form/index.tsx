import { useEffect, useRef, useState } from 'react'
import { toast } from 'sonner'
import { Check, Copy } from 'lucide-react'
import type { Member, ModelInfo, PlatformKeyScope } from '@/api/types'
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
  const { memberId } = useSession()
  const { billingCurrency } = useBillingExchange()
  const currencyLabel = currencySymbol(billingCurrency)
  const { resolveAllowedModelIds: resolveAllModels } = useMemberWhitelist()

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
  const [createdFullKey, setCreatedFullKey] = useState<string | null>(null)

  // Fetch existing key names to prevent duplicates
  const [existingNames, setExistingNames] = useState<Set<string>>(new Set())
  useEffect(() => {
    let cancelled = false
    void apis.keysApi.platform.list({ memberId: effectiveMemberId || memberId }).then((res) => {
      if (!cancelled) {
        setExistingNames(new Set(res.items.map((k) => k.name)))
      }
    })
    return () => {
      cancelled = true
    }
  }, [apis, effectiveMemberId, memberId])

  const nameDuplicate = isCreate && !!name.trim() && existingNames.has(name.trim())

  // Fetch available models for the picker
  const [availableModels, setAvailableModels] = useState<ModelInfo[]>([])
  useEffect(() => {
    // ponytail: memberId 为空时 session 还在加载，跳过避免无效请求被 cancel 浪费时间
    if (!memberId) return
    let cancelled = false
    const resolve = async () => {
      const allModels = await apis.modelsApi.list()
      const enabled = allModels.filter((m) => m.active)
      const allowedIds = await resolveAllModels()
      if (!allowedIds) return enabled
      const allowed = new Set(allowedIds)
      return enabled.filter((m) => allowed.has(m.modelId))
    }
    void resolve()
      .then((models) => {
        if (!cancelled) setAvailableModels(models)
      })
      .catch(() => {
        // Fallback: show all active models rather than stuck on "加载中"
        if (!cancelled) {
          void apis.modelsApi.list().then((all) => setAvailableModels(all.filter((m) => m.active)))
        }
      })
    return () => {
      cancelled = true
    }
  }, [memberId, apis.modelsApi, resolveAllModels])

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

  // Default budget to user's remaining quota — only on first load, not after user edits
  const budgetInitialized = useRef(!isCreate || !!budget)
  useEffect(() => {
    if (budgetInitialized.current) return
    if (budgetSummary && budgetSummary.remaining > 0) {
      setBudget(String(budgetSummary.remaining))
      budgetInitialized.current = true
    }
  }, [budgetSummary, setBudget])

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

  const budgetInvalid = budget === '' || budgetAmount <= 0
  const formInvalid =
    !name.trim() ||
    nameDuplicate ||
    models.length === 0 ||
    budgetInvalid ||
    (requiresMemberPick && !targetMemberId)

  // ponytail: 优先级从上到下，返回第一个成立的原因
  const formDisabledReason = submitting
    ? '请求中…'
    : !name.trim()
      ? '请输入 Key 名称'
      : nameDuplicate
        ? '已存在同名 Key'
        : models.length === 0
          ? '请至少选择一个模型'
          : budgetInvalid
            ? '请输入有效额度'
            : requiresMemberPick && !targetMemberId
              ? '请先选择绑定成员'
              : budgetInsufficient
                ? '额度不足，请先申请追加'
                : budgetExceedsRemaining
                  ? '额度不能超过剩余可用'
                  : projectBudgetExceeds
                    ? '额度不能超过项目剩余'
                    : subBudgetExceeds
                      ? '额度不能超过成员子额度剩余'
                      : undefined

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
      const created = await apis.keysApi.platform.create({
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
      setCreatedFullKey(created.fullKey)
      onSetDirty(false)
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
      await apis.keysApi.platform.update(key.id, {
        name,
        budget: budgetAmount,
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

  // --- Created key result view ---
  if (createdFullKey) {
    return (
      <WorkflowPanelChrome
        title="创建成功"
        onClose={onClose}
        footer={<WorkflowPanelFooter primaryLabel="完成" onPrimary={onClose} />}
      >
        <div className="space-y-6 py-4">
          <div className="flex items-center gap-2 text-emerald-600">
            <Check className="size-5" />
            <span className="text-sm font-medium">Key 已创建，请立即复制保存</span>
          </div>
          <div className="space-y-2">
            <Label>API Key</Label>
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded-md border bg-muted px-3 py-2 font-mono text-sm break-all select-all">
                {createdFullKey}
              </code>
              <CopyKeyButton text={createdFullKey} />
            </div>
            <p className="text-xs text-muted-foreground">
              关闭后将无法再次查看完整 Key，请妥善保管。
            </p>
          </div>
        </div>
      </WorkflowPanelChrome>
    )
  }

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
              formInvalid ||
              budgetInsufficient ||
              budgetExceedsRemaining ||
              projectBudgetExceeds ||
              subBudgetExceeds
            }
            primaryDisabledReason={formDisabledReason}
          />
        ) : (
          <WorkflowPanelFooter
            onCancel={onClose}
            primaryLabel={submitting ? '保存中...' : '保存'}
            onPrimary={handleSave}
            primaryDisabled={submitting || formInvalid}
            primaryDisabledReason={formDisabledReason}
          />
        )
      }
    >
      <div className="space-y-5">
        {requiresMemberPick && (
          <div className="space-y-1.5">
            <Label>绑定成员</Label>
            <Button variant="outline" className="w-full justify-start" onClick={openPickMember}>
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
            maxLength={64}
          />
          {nameDuplicate && <p className="text-xs text-destructive">已存在同名 Key</p>}
        </div>
        <div className="space-y-1.5">
          <Label>额度 ({currencyLabel})</Label>
          <Input
            type="number"
            min="1"
            value={budget}
            onChange={(e) => {
              setBudget(e.target.value)
              onSetDirty(true)
            }}
          />
          {budget !== '' && budgetAmount <= 0 && (
            <p className="text-xs text-destructive">额度必须大于 0</p>
          )}
        </div>

        <InlineModelPicker
          value={models}
          onChange={handleModelsChange}
          models={availableModels}
          hint="至少选择一个模型"
        />
      </div>
    </WorkflowPanelChrome>
  )
}

function CopyKeyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  return (
    <Button
      variant="outline"
      size="icon"
      className="shrink-0"
      aria-label="复制"
      onClick={() => {
        void navigator.clipboard.writeText(text).then(() => {
          setCopied(true)
          setTimeout(() => setCopied(false), 2000)
        })
      }}
    >
      {copied ? <Check className="size-4 text-emerald-600" /> : <Copy className="size-4" />}
    </Button>
  )
}
