import type { CallLog, OperationLog } from '@/api/types'

export const mockCallLogs: CallLog[] = [
  {
    id: 'cl-1',
    caller: '张三',
    callerId: 'm-1',
    callerType: 'member',
    model: 'gpt-4o',
    provider: 'openai',
    inputTokens: 100,
    outputTokens: 50,
    latencyMs: 200,
    status: 'success',
    cost: 1.5,
    createdAt: '2026-06-19 10:00',
    previewSnippet: 'hello world',
  },
]

export const mockOperationLogs: OperationLog[] = [
  {
    id: 'op-1',
    createdAt: '2026-06-19 09:00',
    action: 'key_create',
    operator: 'Bob',
    operatorId: 'm-2',
    target: 'key-1',
    detail: 'Created platform key',
    ip: '10.0.0.1',
  },
]
