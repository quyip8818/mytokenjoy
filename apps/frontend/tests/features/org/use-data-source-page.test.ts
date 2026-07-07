import { describe, expect, it, vi } from 'vitest'
import { useDataSourcePage } from '@/features/org/hooks/use-data-source-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

describe('useDataSourcePage', () => {
  it('enters select phase when data source is not connected', async () => {
    const apis = createMockApis({
      dataSourceApi: {
        getStatus: vi.fn().mockResolvedValue({ connected: false, platform: null }),
      },
    })

    const { result } = renderHookWithProviders(() => useDataSourcePage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(result.current.phase).toBe('select')
    expect(apis.dataSourceApi.getStatus).toHaveBeenCalled()
  })

  it('enters connected phase when data source is connected', async () => {
    const apis = createMockApis({
      dataSourceApi: {
        getStatus: vi.fn().mockResolvedValue({ connected: true, platform: 'feishu' }),
      },
    })

    const { result } = renderHookWithProviders(() => useDataSourcePage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(result.current.phase).toBe('connected')
    expect(result.current.platform).toBe('feishu')
  })
})
