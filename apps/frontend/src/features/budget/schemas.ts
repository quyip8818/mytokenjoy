import { z } from 'zod'
import type { BudgetNode } from '@/api/types'

const OverrunPolicySchema = z.enum(['hard_reject', 'approval', 'downgrade'])

export const BudgetNodeSchema: z.ZodType<BudgetNode> = z.lazy(() =>
  z.object({
    id: z.string(),
    name: z.string(),
    parentId: z.string().nullable(),
    budget: z.number(),
    consumed: z.number(),
    reservedPool: z.number().optional(),
    children: z.array(BudgetNodeSchema).optional(),
    period: z.string(),
  }),
)

export const BudgetGroupSchema = z.object({
  id: z.string(),
  name: z.string(),
  budget: z.number(),
  consumed: z.number(),
  memberIds: z.array(z.string()),
  departmentIds: z.array(z.string()),
})

export const BudgetApprovalSchema = z.object({
  id: z.string(),
  applicantName: z.string(),
  departmentName: z.string(),
  amount: z.number(),
  reason: z.string(),
  status: z.enum(['pending', 'approved', 'rejected']),
  createdAt: z.string(),
  resolvedAt: z.string().optional(),
  rejectReason: z.string().optional(),
})

export const BudgetTreeResponseSchema = z.array(BudgetNodeSchema)
export const BudgetGroupsResponseSchema = z.array(BudgetGroupSchema)
export const BudgetApprovalsResponseSchema = z.array(BudgetApprovalSchema)

export const BudgetProjectViewSchema = z.object({
  id: z.string(),
  name: z.string(),
  budget: z.number(),
  consumed: z.number(),
  memberIds: z.array(z.string()),
  departmentId: z.string(),
  departmentName: z.string(),
  overrunPolicy: OverrunPolicySchema,
  period: z.string(),
})
