import type { KeyApproval } from '@/api/types'

export const mockApprovals: KeyApproval[] = [
  {
    id: 'a1',
    type: 'key',
    applicant: '张三',
    applicantId: 'm1',
    department: '研发部',
    reason: '需要 API 访问',
    requestedQuota: 0,
    requestedModels: ['gpt-4'],
    status: 'pending',
    approver: null,
    createdAt: '2026-01-01',
    resolvedAt: null,
  },
]
