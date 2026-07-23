import type { ApprovalRequest } from '@/api/types'

export const mockApprovals: ApprovalRequest[] = [
  {
    id: 'a1',
    type: 'key',
    status: 'pending',
    applicantId: 'm1',
    applicantName: '张三',
    scopeId: 'd1',
    canResolve: true,
    metadata: {
      reason: '需要 API 访问',
      requestedBudget: 5000,
      requestedModels: ['00000000-0000-7000-8000-0000000000b1'],
      departmentId: 'd1',
      departmentName: '研发部',
    },
    createdAt: '2026-01-01',
  },
]
