import { describe, expect, it, vi } from 'vitest'
import { useModelListPage } from '@/features/models/hooks/use-model-list-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'
import { mockModels } from '@tests/fixtures/models'

describe('useModelListPage', () => {
  it('loads models on mount', async () => {
    const apis = createMockApis({
      modelApi: {
        list: vi.fn().mockResolvedValue(mockModels),
      },
    })

    const { result } = renderHookWithProviders(() => useModelListPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.modelApi.list).toHaveBeenCalled()
    expect(result.current.models.length).toBeGreaterThan(0)
  })

  it('filters to builtin only when not selfhosted', async () => {
    const apis = createMockApis({
      modelApi: {
        list: vi.fn().mockResolvedValue(mockModels),
      },
    })

    const { result } = renderHookWithProviders(() => useModelListPage(apis), {
      apis,
      companyType: 'standard',
    })

    await waitForLoaded(result, 'loading')

    // custom models should be filtered out in SaaS mode
    const hasCustom = result.current.models.some((m) => m.provider === 'custom')
    expect(hasCustom).toBe(false)
  })

  it('returns isSelfHosted true for selfhosted company', async () => {
    const apis = createMockApis({
      modelApi: {
        list: vi.fn().mockResolvedValue(mockModels),
      },
    })

    const { result } = renderHookWithProviders(() => useModelListPage(apis), {
      apis,
      companyType: 'selfhosted',
    })

    await waitForLoaded(result, 'loading')

    expect(result.current.isSelfHosted).toBe(true)
  })

  it('shows all models including custom when selfhosted', async () => {
    const apis = createMockApis({
      modelApi: {
        list: vi.fn().mockResolvedValue(mockModels),
      },
    })

    const { result } = renderHookWithProviders(() => useModelListPage(apis), {
      apis,
      companyType: 'selfhosted',
    })

    await waitForLoaded(result, 'loading')

    expect(result.current.models).toEqual(mockModels)
  })
})
