import { useEffect, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { MemberQuotaSummary, PlatformKey } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'

export function formatQuotaContext(
  summary: MemberQuotaSummary | null,
  department?: string,
): string {
  if (!summary) return department ? `部门：${department}` : ''
  const parts = [`剩余额度 ¥${summary.remaining.toLocaleString()}`]
  if (department) parts.unshift(department)
  return parts.join(' · ')
}

export interface UseKeyFormQuotaOptions {
  isCreate: boolean
  isGroupKey: boolean
  effectiveMemberId: string
  budgetGroupId?: string
  quota: string
  adminCreate: boolean
  injectedApis?: AppApis
}

export function useKeyFormQuota({
  isCreate,
  isGroupKey,
  effectiveMemberId,
  budgetGroupId,
  quota,
  adminCreate,
  injectedApis,
}: UseKeyFormQuotaOptions) {
  const apis = useInjectedApis(injectedApis)
  const [quotaState, setQuotaState] = useState<{
    memberId: string
    summary: MemberQuotaSummary
  } | null>(null)
  const [groupQuotaRemaining, setGroupQuotaRemaining] = useState<number | null>(null)

  useEffect(() => {
    if (!isCreate || isGroupKey || !effectiveMemberId) return
    let cancelled = false
    void apis.platformKeyApi.getQuotaSummary(effectiveMemberId).then((summary) => {
      if (!cancelled) setQuotaState({ memberId: effectiveMemberId, summary })
    })
    return () => {
      cancelled = true
    }
  }, [apis, isCreate, isGroupKey, effectiveMemberId])

  useEffect(() => {
    if (!isCreate || !budgetGroupId) return
    let cancelled = false
    void Promise.all([
      apis.budgetApi.getGroups(),
      apis.platformKeyApi.list({ budgetGroupId }),
    ]).then(([groups, keysRes]) => {
      if (cancelled) return
      const group = groups.find((g) => g.id === budgetGroupId)
      if (!group) {
        setGroupQuotaRemaining(null)
        return
      }
      const allocated = keysRes.items
        .filter((k) => k.status === 'active')
        .reduce((sum, k) => sum + k.quota, 0)
      setGroupQuotaRemaining(Math.max(0, group.budget - group.consumed - allocated))
    })
    return () => {
      cancelled = true
    }
  }, [apis, isCreate, budgetGroupId])

  const quotaSummary = quotaState?.memberId === effectiveMemberId ? quotaState.summary : null
  const quotaInsufficient =
    isCreate && !isGroupKey && !adminCreate && quotaSummary !== null && quotaSummary.remaining <= 0
  const quotaExceedsRemaining =
    isCreate && !isGroupKey && quotaSummary !== null && Number(quota) > quotaSummary.remaining
  const groupQuotaExceeds =
    isCreate && isGroupKey && groupQuotaRemaining !== null && Number(quota) > groupQuotaRemaining

  return {
    quotaSummary,
    groupQuotaRemaining,
    quotaInsufficient,
    quotaExceedsRemaining,
    groupQuotaExceeds,
  }
}

export interface UseKeyFormStateOptions {
  key?: PlatformKey
  adminCreate: boolean
  defaultMemberId: string
  initialTargetMemberId?: string
  initialName?: string
  initialQuota?: string
}

export function useKeyFormState({
  key,
  adminCreate,
  defaultMemberId,
  initialTargetMemberId,
  initialName,
  initialQuota,
}: UseKeyFormStateOptions) {
  const [step, setStep] = useState(1)
  const [name, setName] = useState(key?.name ?? initialName ?? '')
  const [quota, setQuota] = useState(String(key?.quota ?? initialQuota ?? '5000'))
  const [models, setModels] = useState<number[]>(key?.modelWhitelist ?? [])
  const [targetMemberId, setTargetMemberId] = useState(
    adminCreate ? (initialTargetMemberId ?? '') : defaultMemberId,
  )
  const [targetMemberName, setTargetMemberName] = useState('')
  const [submitting, setSubmitting] = useState(false)

  return {
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
  }
}
