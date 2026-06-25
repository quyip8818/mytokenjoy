import { describe, expect, it } from 'vitest'
import {
  buildCallAuditCsvRows,
  buildOperationAuditCsvRows,
  CALL_AUDIT_CSV_HEADERS,
  OPERATION_AUDIT_CSV_HEADERS,
} from '@/lib/audit-export'
import type { CallLog, OperationLog } from '@/api/types'

const sampleCallLog: CallLog = {
  id: 'call-1',
  createdAt: '2026-06-19 10:30',
  caller: 'Alice',
  callerId: 'm-1',
  callerType: 'member',
  model: 'gpt-4',
  provider: 'openai',
  inputTokens: 100,
  outputTokens: 50,
  latencyMs: 1200,
  status: 'success',
  cost: 0.05,
  inputPreview: 'hello',
  outputPreview: 'world',
}

const sampleOperationLog: OperationLog = {
  id: 'op-1',
  createdAt: '2026-06-19 09:00',
  action: 'key_create',
  operator: 'Bob',
  operatorId: 'm-2',
  target: 'key-1',
  detail: 'Created platform key',
  ip: '10.0.0.1',
}

describe('audit-export', () => {
  it('defines stable CSV headers', () => {
    expect(CALL_AUDIT_CSV_HEADERS).toHaveLength(8)
    expect(OPERATION_AUDIT_CSV_HEADERS).toHaveLength(6)
  })

  it('buildCallAuditCsvRows maps call log fields', () => {
    expect(buildCallAuditCsvRows([sampleCallLog])).toEqual([
      ['2026-06-19 10:30', 'Alice', 'gpt-4', '100', '50', '1200', 'success', '0.05'],
    ])
  })

  it('buildOperationAuditCsvRows maps action labels', () => {
    expect(buildOperationAuditCsvRows([sampleOperationLog])).toEqual([
      ['2026-06-19 09:00', 'Key 创建', 'Bob', 'key-1', 'Created platform key', '10.0.0.1'],
    ])
  })
})
