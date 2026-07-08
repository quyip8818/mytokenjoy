import { describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'
import { useAuditModelOptions } from '@/features/audit/hooks/use-audit-model-options'
import { createMockApis, renderHookWithProviders } from '@tests/utils'

describe('useAuditModelOptions', () => {
  it('loads model options on mount', async () => {
    const models = [{ modelId: 1, type: 'gpt-4', name: 'GPT-4', provider: 'openai', enabled: true }]
    const apis = createMockApis({
      modelApi: {
        list: vi.fn().mockResolvedValue(models),
      },
    })

    const { result } = renderHookWithProviders(() => useAuditModelOptions(apis), { apis })

    await waitFor(() => {
      expect(result.current.models).toEqual(models)
    })

    expect(apis.modelApi.list).toHaveBeenCalled()
  })
})
