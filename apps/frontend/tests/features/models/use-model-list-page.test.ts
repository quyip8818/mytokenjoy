import { describe, expect, it, vi } from 'vitest'
import { useModelListPage } from '@/features/models/hooks/use-model-list-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'
import { mockModels } from '@tests/fixtures/models'

// ponytail: mock IS_SAAS at module level — toggled per test via mockReturnValue
const isSaasMock = vi.fn(() => false)
vi.mock('@/config/app', async (importOriginal) => {
  const actual = (await importOriginal()) as Record<string, unknown>
  return {
    ...actual,
    get IS_SAAS() {
      return isSaasMock()
    },
  }
})

describe('useModelListPage', () => {
  it('loads models on mount', async () => {
    const apis = createMockApis({
      modelsApi: {
        list: vi.fn().mockResolvedValue(mockModels),
      },
    })

    const { result } = renderHookWithProviders(() => useModelListPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.modelsApi.list).toHaveBeenCalled()
    expect(result.current.models.length).toBeGreaterThan(0)
  })

  it('filters to builtin only in SaaS mode', async () => {
    isSaasMock.mockReturnValue(true)
    const apis = createMockApis({
      modelsApi: {
        list: vi.fn().mockResolvedValue(mockModels),
      },
    })

    const { result } = renderHookWithProviders(() => useModelListPage(apis), {
      apis,
      companyType: 'saas',
    })

    await waitForLoaded(result, 'loading')

    // custom models should be filtered out in SaaS mode
    const hasCustom = result.current.models.some((m) => m.provider === 'custom')
    expect(hasCustom).toBe(false)
    isSaasMock.mockReturnValue(false)
  })

  it('shows all models including custom in local deploy', async () => {
    isSaasMock.mockReturnValue(false)
    const apis = createMockApis({
      modelsApi: {
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

  it('builds discountMap from billing discounts', async () => {
    const discounts = [
      { modelType: 'deepseek-v4-flash', discount: 0.9, note: '限时优惠' },
      { modelType: '*', discount: 0.95 },
    ]
    const apis = createMockApis({
      modelsApi: { list: vi.fn().mockResolvedValue(mockModels) },
      billingApi: { myDiscounts: vi.fn().mockResolvedValue(discounts) },
    })

    const { result } = renderHookWithProviders(() => useModelListPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(result.current.discountMap.size).toBe(2)
    expect(result.current.discountMap.get('deepseek-v4-flash')?.discount).toBe(0.9)
    expect(result.current.discountMap.get('*')?.discount).toBe(0.95)
  })

  it('returns empty discountMap when no discounts configured', async () => {
    const apis = createMockApis({
      modelsApi: { list: vi.fn().mockResolvedValue(mockModels) },
      billingApi: { myDiscounts: vi.fn().mockResolvedValue([]) },
    })

    const { result } = renderHookWithProviders(() => useModelListPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(result.current.discountMap.size).toBe(0)
  })
})
