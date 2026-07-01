import { describe, expect, it } from 'vitest'
import { SessionContextSchema } from '@/api/schemas/session'

describe('SessionContextSchema', () => {
  it('accepts valid session payloads', () => {
    const result = SessionContextSchema.safeParse({
      companyId: 1,
      member: {
        id: 'm1',
        companyId: 1,
        name: 'Admin',
        phone: '13800000000',
        email: 'admin@test.com',
        departmentId: 'd1',
        departmentName: 'HQ',
        status: 'active',
        roles: ['超级管理员'],
        source: 'manual',
      },
      permissions: ['org.structure'],
      readOnly: false,
    })

    expect(result.success).toBe(true)
  })

  it('rejects invalid session payloads', () => {
    const result = SessionContextSchema.safeParse({
      companyId: 1,
      member: { id: 'm1', companyId: 1 },
      permissions: [],
      readOnly: false,
    })

    expect(result.success).toBe(false)
  })
})
