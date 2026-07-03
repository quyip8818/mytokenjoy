import { z } from 'zod'

const MemberStatusSchema = z.enum(['active', 'inactive', 'pending'])
const MemberSourceSchema = z.enum(['imported', 'manual', 'invited'])

export const MemberSchema = z.object({
  id: z.string(),
  companyId: z.number(),
  name: z.string(),
  phone: z.string(),
  email: z.string(),
  departmentId: z.string(),
  departmentName: z.string(),
  status: MemberStatusSchema,
  roles: z.array(z.string()),
  source: MemberSourceSchema,
})

export const SessionContextSchema = z.object({
  companyId: z.number(),
  authzRevision: z.number(),
  member: MemberSchema,
  permissions: z.array(z.string()),
  readOnly: z.boolean(),
})

export type SessionContextDto = z.infer<typeof SessionContextSchema>
