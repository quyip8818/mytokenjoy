import { toast } from 'sonner'
import type { Member, PlatformKeyScope } from '@/api/types'
import type { AppApis } from '@/api/app-apis'
import { useInjectedApis } from '@/api/use-apis'
import { useSession } from '@/features/session'
import { pushModelPicker, useMemberWhitelist } from '@/features/workflow/use-member-whitelist'
import type { WorkflowComponentProps, WorkflowStackEntry } from '@/features/workflow/types'
import {
  WorkflowPanelChrome,
  WorkflowPanelFooter,
} from '@/features/workflow/components/workflow-panel-chrome'
import { WorkflowInfoBox } from '@/features/workflow/components/workflow-info-box'
import { WorkflowStepper } from '@/features/workflow/components/workflow-stepper'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useWorkflow } from '@/features/workflow/use-workflow'
import { workflowErrorMessage } from '@/features/workflow/lib/error-message'
import { BUDGET_INSUFFICIENT_MESSAGE } from '@/features/keys'
import { useModelLabels } from '@/features/models/hooks/use-model-labels'
import { formatBudgetContext, useKeyFormBudget, useKeyFormState } from './use-key-form-budget'
import { formatDisplayCurrency } from '@/lib/points'

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
  const { resolveAllowedModelIds } = useMemberWhitelist()
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
    step,
    setStep,
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

  const { labelFor } = useModelLabels()
  const effectiveMemberId = adminCreate || scope === 'project_member' ? targetMemberId : memberId
  const requiresMemberPick = adminCreate || scope === 'project_member'

  const {
    budgetSummary,
    projectBudgetRemaining,
    subBudgetRemaining,
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

  const openPickMember = () => {
    onPush('member-search', {
      multi: false,
      excludeIds: [],
      onConfirm: (members: Member[]) => {
        const picked = members[0]
        if (!picked) return
        setTargetMemberId(picked.id)
        setTargetMemberName(picked.name)
        onSetDirty(true)
      },
    })
  }

  const handlePickModels = () => {
    void pushModelPicker(onPush, resolveAllowedModelIds, {
      selectedModelIds: models,
      onConfirm: setModels,
      onSetDirty,
    })
  }

  const handleCreate = async () => {
    if (budgetInsufficient) {
      toast.error(BUDGET_INSUFFICIENT_MESSAGE)
      return
    }
    if (budgetSummary && Number(budget) > budgetSummary.remaining) {
      toast.error(`额度不能超过剩余 ${formatDisplayCurrency(budgetSummary.remaining)}`)
      return
    }
    if (projectBudgetExceeds) {
      toast.error(`额度不能超过项目剩余 ${formatDisplayCurrency(projectBudgetRemaining!)}`)
      return
    }
    if (subBudgetExceeds) {
      toast.error(`额度不能超过成员子额度剩余 ${formatDisplayCurrency(subBudgetRemaining!)}`)
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
        budget: Number(budget),
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
        budget: Number(budget),
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
      return `项目：${projectName ?? ''} · 成员：${targetMemberName || '—'} · 子额度剩余 ${formatDisplayCurrency(subBudgetRemaining ?? 0)}`
    }
    if (scope === 'project') {
      return `项目：${projectName ?? ''} · 剩余可分配 ${formatDisplayCurrency(projectBudgetRemaining ?? 0)}`
    }
    return formatBudgetContext(
      budgetSummary,
      adminCreate ? targetMemberName || undefined : undefined,
    )
  })()

  const modelSection = (
    <div className="space-y-3">
      <Label>模型白名单</Label>
      <Button variant="outline" onClick={handlePickModels}>
        选择模型 {models.length > 0 && `(${models.length})`}
      </Button>
      {models.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {models.map((modelId) => (
            <Badge key={modelId} variant="outline" className="text-xs">
              {labelFor(modelId)}
            </Badge>
          ))}
        </div>
      )}
    </div>
  )

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
            申请额度超过剩余 {formatDisplayCurrency(budgetSummary!.remaining)}
          </p>
        ) : projectBudgetExceeds ? (
          <p className="text-sm text-amber-800">
            申请额度超过项目剩余 {formatDisplayCurrency(projectBudgetRemaining!)}
          </p>
        ) : subBudgetExceeds ? (
          <p className="text-sm text-amber-800">
            申请额度超过成员子额度剩余 {formatDisplayCurrency(subBudgetRemaining!)}
          </p>
        ) : undefined
      }
      footer={
        isCreate ? (
          step === 1 ? (
            <WorkflowPanelFooter
              onCancel={onClose}
              primaryLabel="下一步"
              onPrimary={() => setStep(2)}
              primaryDisabled={
                budgetInsufficient ||
                !name.trim() ||
                (requiresMemberPick && !targetMemberId) ||
                budgetExceedsRemaining ||
                projectBudgetExceeds ||
                subBudgetExceeds
              }
            />
          ) : (
            <WorkflowPanelFooter
              onCancel={onClose}
              secondaryLabel="上一步"
              onSecondary={() => setStep(1)}
              primaryLabel={submitting ? '创建中...' : '创建 Key'}
              onPrimary={handleCreate}
              primaryDisabled={
                models.length === 0 ||
                submitting ||
                budgetInsufficient ||
                budgetExceedsRemaining ||
                projectBudgetExceeds ||
                subBudgetExceeds
              }
            />
          )
        ) : (
          <WorkflowPanelFooter
            onCancel={onClose}
            primaryLabel={submitting ? '保存中...' : '保存'}
            onPrimary={handleSave}
            primaryDisabled={submitting || !name.trim() || models.length === 0}
          />
        )
      }
    >
      <div className="grid grid-cols-5 gap-8">
        <div className="col-span-3 space-y-5">
          {isCreate && <WorkflowStepper steps={['基本信息', '模型白名单']} current={step} />}
          {isCreate && step === 1 ? (
            <>
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
                <Label>额度 (¥)</Label>
                <Input
                  type="number"
                  value={budget}
                  onChange={(e) => {
                    setBudget(e.target.value)
                    onSetDirty(true)
                  }}
                />
              </div>
            </>
          ) : (
            <>
              {!isCreate && (
                <>
                  <div className="space-y-1.5">
                    <Label>Key 名称</Label>
                    <Input
                      value={name}
                      onChange={(e) => {
                        setName(e.target.value)
                        onSetDirty(true)
                      }}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label>额度 (¥)</Label>
                    <Input
                      type="number"
                      value={budget}
                      onChange={(e) => {
                        setBudget(e.target.value)
                        onSetDirty(true)
                      }}
                    />
                  </div>
                </>
              )}
              {modelSection}
            </>
          )}
        </div>
        <WorkflowInfoBox fullWidth className="space-y-3">
          <h4 className="font-semibold text-foreground/80">{isCreate ? '摘要' : '当前 Key'}</h4>
          {isCreate ? (
            <div className="space-y-2 text-muted-foreground">
              <p>名称：{name || '—'}</p>
              <p>额度：{formatDisplayCurrency(Number(budget))}</p>
              <p>模型：{models.length} 个</p>
            </div>
          ) : (
            key && (
              <>
                <p className="text-muted-foreground font-mono">{key.keyPrefix}</p>
                <p className="text-muted-foreground">
                  已消耗：{formatDisplayCurrency(key.consumed)}
                </p>
              </>
            )
          )}
        </WorkflowInfoBox>
      </div>
    </WorkflowPanelChrome>
  )
}
