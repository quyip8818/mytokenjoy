import { describe, expect, it, vi } from 'vitest'
import { useModelRoutingPage } from '@/features/models/hooks/use-model-routing-page'
import { createMockApis, mockDepartments, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useModelRoutingPage', () => {
  it('loads routing rules and departments on mount', async () => {
    const rules = [
      {
        nodeId: 'd1',
        allowedModels: ['gpt-4'],
        whitelistEnabled: false,
        whitelistedMemberIds: [],
      },
    ]
    const apis = createMockApis({
      routingApi: {
        getRules: vi.fn().mockResolvedValue(rules),
      },
      departmentApi: {
        getTree: vi.fn().mockResolvedValue(mockDepartments),
      },
    })

    const { result } = renderHookWithProviders(() => useModelRoutingPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.routingApi.getRules).toHaveBeenCalled()
    expect(apis.departmentApi.getTree).toHaveBeenCalled()
    expect(result.current.rules).toEqual(rules)
  })
})
