import type { CallLog, OperationLog } from '@/api/types'
import { OPERATION_ACTION_LABELS } from '@/features/audit/lib/labels'

export const CALL_AUDIT_CSV_HEADERS = [
  '时间',
  '调用人',
  '模型',
  '输入 Token',
  '输出 Token',
  '延迟(ms)',
  '状态',
  '费用',
] as const

export const OPERATION_AUDIT_CSV_HEADERS = [
  '时间',
  '操作类型',
  '操作人',
  '操作对象',
  '详情',
  'IP',
] as const

export function buildCallAuditCsvRows(logs: CallLog[]): string[][] {
  return logs.map((log) => [
    log.createdAt,
    log.caller,
    log.model,
    String(log.inputTokens),
    String(log.outputTokens),
    String(log.latencyMs),
    log.status,
    String(log.cost),
  ])
}

export function buildOperationAuditCsvRows(logs: OperationLog[]): string[][] {
  return logs.map((log) => [
    log.createdAt,
    OPERATION_ACTION_LABELS[log.action] ?? log.action,
    log.operator,
    log.target,
    log.detail,
    log.ip,
  ])
}
