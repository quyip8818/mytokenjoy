import { useState } from 'react'
import type { AlertRuleView } from '@/features/budget/lib/alerts'

export type AlertRuleFormState = {
  targetType: 'team' | 'project'
  targetId: string
  targetName: string
  thresholds: number[]
  notifyRoleIds: string[]
}

export function buildAlertRulePayload(
  state: AlertRuleFormState,
  rule: AlertRuleView | null,
): AlertRuleView {
  return {
    id: rule?.id ?? '',
    nodeId: state.targetId,
    nodeName: state.targetName,
    targetType: state.targetType,
    targetId: state.targetId,
    targetName: state.targetName,
    thresholds: state.thresholds,
    notifyRoleIds: state.notifyRoleIds,
    enabled: rule?.enabled ?? true,
  }
}

export function validateAlertRuleForm(state: AlertRuleFormState): string | null {
  if (!state.targetId) return '请选择监控对象'
  if (state.thresholds.length === 0) return '请至少设置一个阈值'
  if (state.notifyRoleIds.length === 0) return '请选择通知角色'
  return null
}

export function useAlertRuleForm(rule: AlertRuleView | null) {
  const [targetType, setTargetType] = useState<'team' | 'project'>(rule?.targetType ?? 'team')
  const [targetId, setTargetId] = useState(rule?.targetId ?? '')
  const [targetName, setTargetName] = useState(rule?.targetName ?? '')
  const [thresholds, setThresholds] = useState<number[]>(
    rule?.thresholds ? [...rule.thresholds] : [80, 90, 100],
  )
  const [customThreshold, setCustomThreshold] = useState('')
  const [notifyRoleIds, setNotifyRoleIds] = useState<string[]>(
    rule?.notifyRoleIds ? [...rule.notifyRoleIds] : [],
  )

  function resetTarget() {
    setTargetId('')
    setTargetName('')
  }

  function setTargetTypeAndReset(value: 'team' | 'project') {
    setTargetType(value)
    resetTarget()
  }

  function setTarget(id: string, name: string) {
    setTargetId(id)
    setTargetName(name)
  }

  function addCustomThreshold() {
    const num = parseInt(customThreshold, 10)
    if (Number.isNaN(num) || num <= 0 || num > 100) return
    if (thresholds.includes(num)) return
    setThresholds([...thresholds, num].sort((a, b) => a - b))
    setCustomThreshold('')
  }

  function removeThreshold(value: number) {
    setThresholds(thresholds.filter((threshold) => threshold !== value))
  }

  function toggleRole(roleId: string) {
    setNotifyRoleIds((prev) =>
      prev.includes(roleId) ? prev.filter((id) => id !== roleId) : [...prev, roleId],
    )
  }

  function selectPreset(preset: number[]) {
    setThresholds([...preset])
  }

  const state: AlertRuleFormState = {
    targetType,
    targetId,
    targetName,
    thresholds,
    notifyRoleIds,
  }

  return {
    state,
    customThreshold,
    setCustomThreshold,
    setTargetTypeAndReset,
    setTarget,
    addCustomThreshold,
    removeThreshold,
    toggleRole,
    selectPreset,
    validate: () => validateAlertRuleForm(state),
    buildPayload: () => buildAlertRulePayload(state, rule),
  }
}
