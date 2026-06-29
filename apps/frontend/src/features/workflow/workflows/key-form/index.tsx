import { toast } from 'sonner'
import type { Member } from '@/api/types'
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
import { QUOTA_INSUFFICIENT_MESSAGE } from '@/features/workflow/constants'
import { formatQuotaContext, useKeyFormQuota, useKeyFormState } from './use-key-form-quota'

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
  const { resolveWhitelist } = useMemberWhitelist()
  const isCreate = entry.id === 'key-create'
  const key =
    entry.id === 'key-edit' ? (entry as WorkflowStackEntry<'key-edit'>).payload.key : undefined
  const adminCreate = isCreate ? Boolean(entry.payload.adminCreate) : false
  const budgetGroupId = isCreate ? entry.payload.budgetGroupId : undefined
  const budgetGroupName = isCreate ? entry.payload.budgetGroupName : undefined
  const onSuccess = entry.payload.onSuccess

  const {
    step,
    setStep,
    name,
    setName,
    quota,
    setQuota,
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
    initialTargetMemberId:
      isCreate && entry.id === 'key-create' ? entry.payload.targetMemberId : undefined,
  })

  const effectiveMemberId = adminCreate ? targetMemberId : memberId
  const isGroupKey = Boolean(budgetGroupId)

  const {
    quotaSummary,
    groupQuotaRemaining,
    quotaInsufficient,
    quotaExceedsRemaining,
    groupQuotaExceeds,
  } = useKeyFormQuota({
    isCreate,
    isGroupKey,
    effectiveMemberId,
    budgetGroupId,
    quota,
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
    void pushModelPicker(onPush, resolveWhitelist, {
      selectedModels: models,
      onConfirm: setModels,
      onSetDirty,
    })
  }

  const handleCreate = async () => {
    if (quotaInsufficient) {
      toast.error(QUOTA_INSUFFICIENT_MESSAGE)
      return
    }
    if (quotaSummary && Number(quota) > quotaSummary.remaining) {
      toast.error(`额度不能超过剩余 ¥${quotaSummary.remaining.toLocaleString()}`)
      return
    }
    if (groupQuotaExceeds) {
      toast.error(`额度不能超过预算组剩余 ¥${groupQuotaRemaining!.toLocaleString()}`)
      return
    }
    setSubmitting(true)
    try {
      const created = await apis.platformKeyApi.create({
        name,
        memberId: isGroupKey ? effectiveMemberId || memberId : effectiveMemberId,
        budgetGroupId,
        quota: Number(quota),
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
    } catch {
      toast.error('创建失败')
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
        quota: Number(quota),
        modelWhitelist: models,
      })
      toast.success('Key 已更新')
      onSuccess?.()
      closeAll()
    } catch {
      toast.error('保存失败')
    } finally {
      setSubmitting(false)
    }
  }

  const modelSection = (
    <div className="space-y-3">
      <Label>模型白名单</Label>
      <Button variant="outline" onClick={handlePickModels}>
        选择模型 {models.length > 0 && `(${models.length})`}
      </Button>
      {models.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {models.map((m) => (
            <Badge key={m} variant="outline" className="text-xs">
              {m}
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
      contextBar={
        isCreate
          ? isGroupKey
            ? `预算组：${budgetGroupName ?? ''} · 剩余可分配 ¥${(groupQuotaRemaining ?? 0).toLocaleString()}`
            : formatQuotaContext(
                quotaSummary,
                adminCreate ? targetMemberName || undefined : undefined,
              )
          : undefined
      }
      banner={
        quotaInsufficient ? (
          <p className="text-sm text-amber-800">{QUOTA_INSUFFICIENT_MESSAGE}</p>
        ) : quotaExceedsRemaining ? (
          <p className="text-sm text-amber-800">
            申请额度超过剩余 ¥{quotaSummary!.remaining.toLocaleString()}
          </p>
        ) : groupQuotaExceeds ? (
          <p className="text-sm text-amber-800">
            申请额度超过预算组剩余 ¥{groupQuotaRemaining!.toLocaleString()}
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
                quotaInsufficient ||
                !name.trim() ||
                (adminCreate && !isGroupKey && !targetMemberId) ||
                quotaExceedsRemaining ||
                groupQuotaExceeds
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
                quotaInsufficient ||
                quotaExceedsRemaining ||
                groupQuotaExceeds
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
              {adminCreate && (
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
                  value={quota}
                  onChange={(e) => {
                    setQuota(e.target.value)
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
                      value={quota}
                      onChange={(e) => {
                        setQuota(e.target.value)
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
              <p>额度：¥{Number(quota).toLocaleString()}</p>
              <p>模型：{models.length} 个</p>
            </div>
          ) : (
            key && (
              <>
                <p className="text-muted-foreground font-mono">{key.keyPrefix}</p>
                <p className="text-muted-foreground">已用：¥{key.used.toLocaleString()}</p>
              </>
            )
          )}
        </WorkflowInfoBox>
      </div>
    </WorkflowPanelChrome>
  )
}
