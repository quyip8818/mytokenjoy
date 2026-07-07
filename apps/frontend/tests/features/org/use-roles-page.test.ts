import { describe, expect, it, vi } from 'vitest'
import { useRolesPage } from '@/features/org/hooks/use-roles-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitFor } from '@testing-library/react'

const mockRoles = [
  { id: 'r1', name: '管理员', type: 'preset' as const, memberCount: 1, permissions: ['*'] },
]
const mockPermissions = [{ key: 'org.read', label: '查看组织' }]

describe('useRolesPage', () => {
  it('loads roles and selects the first role on mount', async () => {
    const apis = createMockApis({
      roleApi: {
        list: vi.fn().mockResolvedValue(mockRoles),
        getPermissions: vi.fn().mockResolvedValue(mockPermissions),
        getMembers: vi.fn().mockResolvedValue([]),
      },
    })

    const { result } = renderHookWithProviders(() => useRolesPage(apis), { apis })

    await waitFor(() => {
      expect(result.current.roles).toEqual(mockRoles)
    })

    expect(result.current.selectedRoleId).toBe('r1')
    expect(apis.roleApi.getMembers).toHaveBeenCalledWith('r1')
  })
})
