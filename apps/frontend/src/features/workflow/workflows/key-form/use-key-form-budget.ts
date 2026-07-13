import { useEffect, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { MemberBudgetSummary, PlatformKey, PlatformKeyScope } from '@/api/types'
import { useInjectedApis } from '@/api/use-apis'

export function formatBudgetContext(
  summary: MemberBudgetSummary | null,
  department?: string,
): string {
  if (!summary) return department ? `部门：${department}` : ''
  const parts = [`剩余额度 ¥${summary.remaining.toLocaleString()}`]
  if (department) parts.unshift(department)
  return parts.join(' · ')
}

export interface UseKeyFormBudgetOptions {
  isCreate: boolean
  scope: PlatformKeyScope
  effectiveMemberId: string
  projectId?: string
  budget: string
  adminCreate: boolean
  injectedApis?: AppApis
}

export function useKeyFormBudget({
  isCreate,
  scope,
  effectiveMemberId,
  projectId,
  budget,
  adminCreate,
  injectedApis,
}: UseKeyFormBudgetOptions) {
  const apis = useInjectedApis(injectedApis)
  const [budgetState, setBudgetState] = useState<{
    memberId: string
    summary: MemberBudgetSummary
  } | null>(null)
  const [projectBudgetRemaining, setProjectBudgetRemaining] = useState<number | null>(null)
  const [subBudgetRemaining, setSubBudgetRemaining] = useState<number | null>(null)

  useEffect(() => {
    if (!isCreate || scope !== 'member' || !effectiveMemberId) return
    let cancelled = false
    void apis.platformKeyApi.getBudgetSummary(effectiveMemberId).then((summary) => {
      if (!cancelled) setBudgetState({ memberId: effectiveMemberId, summary })
    })
    return () => {
      cancelled = true
    }
  }, [apis, isCreate, scope, effectiveMemberId])

  useEffect(() => {
    if (!isCreate || scope !== 'project' || !projectId) return
    let cancelled = false
    void Promise.all([apis.budgetApi.getProjects(), apis.platformKeyApi.list({ projectId })]).then(
      ([groups, keysRes]) => {
        if (cancelled) return
        const group = groups.find((g) => g.id === projectId)
        if (!group) {
          setProjectBudgetRemaining(null)
          return
        }
        const allocated = keysRes.items
          .filter((k) => k.status === 'active')
          .reduce((sum, k) => sum + k.budget, 0)
        setProjectBudgetRemaining(Math.max(0, group.budget - group.consumed - allocated))
      },
    )
    return () => {
      cancelled = true
    }
  }, [apis, isCreate, scope, projectId])

  useEffect(() => {
    if (!isCreate || scope !== 'project_member' || !projectId || !effectiveMemberId) return
    let cancelled = false
    void Promise.all([
      apis.budgetApi.getProjects(),
      apis.platformKeyApi.list({ projectId, scope: 'project_member', memberId: effectiveMemberId }),
    ]).then(([groups, keysRes]) => {
      if (cancelled) return
      const group = groups.find((g) => g.id === projectId)
      if (!group) {
        setSubBudgetRemaining(null)
        return
      }
      const subCap = group.memberBudgets?.[effectiveMemberId] ?? 0
      const allocated = keysRes.items
        .filter((k: PlatformKey) => k.status === 'active' && k.memberId === effectiveMemberId)
        .reduce((sum, k) => sum + k.budget, 0)
      setSubBudgetRemaining(Math.max(0, subCap - allocated))
    })
    return () => {
      cancelled = true
    }
  }, [apis, effectiveMemberId, isCreate, projectId, scope])

  const budgetSummary = budgetState?.memberId === effectiveMemberId ? budgetState.summary : null
  const budgetInsufficient =
    isCreate &&
    scope === 'member' &&
    !adminCreate &&
    budgetSummary !== null &&
    budgetSummary.remaining <= 0
  const budgetExceedsRemaining =
    isCreate &&
    scope === 'member' &&
    budgetSummary !== null &&
    Number(budget) > budgetSummary.remaining
  const projectBudgetExceeds =
    isCreate &&
    scope === 'project' &&
    projectBudgetRemaining !== null &&
    Number(budget) > projectBudgetRemaining
  const subBudgetExceeds =
    isCreate &&
    scope === 'project_member' &&
    subBudgetRemaining !== null &&
    Number(budget) > subBudgetRemaining

  return {
    budgetSummary,
    projectBudgetRemaining,
    subBudgetRemaining,
    budgetInsufficient,
    budgetExceedsRemaining,
    projectBudgetExceeds,
    subBudgetExceeds,
  }
}

export interface UseKeyFormStateOptions {
  key?: PlatformKey
  adminCreate: boolean
  defaultMemberId: string
  initialTargetMemberId?: string
  initialName?: string
  initialBudget?: string
}

export function useKeyFormState({
  key,
  adminCreate,
  defaultMemberId,
  initialTargetMemberId,
  initialName,
  initialBudget,
}: UseKeyFormStateOptions) {
  const [step, setStep] = useState(1)
  const [name, setName] = useState(key?.name ?? initialName ?? '')
  const [budget, setBudget] = useState(String(key?.budget ?? initialBudget ?? '5000'))
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
  }
}
