import { describe, expect, it, vi, beforeEach } from 'vitest'
import { act, waitFor } from '@testing-library/react'
import { useSimulateConsumeDialog } from '@/features/dev/hooks/use-simulate-consume-dialog'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import * as simulateConsume from '@/features/dev/lib/simulate-consume'

vi.mock('@/features/dev/lib/simulate-consume', async (importOriginal) => {
  const actual = await importOriginal<typeof simulateConsume>()
  return {
    ...actual,
    postChatCompletions: vi.fn(),
  }
})

describe('useSimulateConsumeDialog', () => {
  beforeEach(() => {
    vi.mocked(simulateConsume.postChatCompletions).mockReset()
    sessionStorage.clear()
  })

  it('selects platform key from dropdown and submits gateway call', async () => {
    const baselineCall = {
      id: 'call-old',
      caller: '张三',
      callerId: 'm-1',
      callerType: 'platform_key' as const,
      model: 'local-test-model',
      provider: 'custom' as const,
      inputTokens: 1,
      outputTokens: 1,
      latencyMs: 10,
      status: 'success' as const,
      cost: 1,
      createdAt: '2026-06-01T00:00:00Z',
      previewSnippet: '',
    }
    const newCall = {
      ...baselineCall,
      id: 'call-new',
      inputTokens: 12_000_000,
      outputTokens: 8_000_000,
    }

    const getCalls = vi
      .fn()
      .mockResolvedValueOnce({ items: [baselineCall], total: 1, page: 1, pageSize: 50 })
      .mockResolvedValueOnce({ items: [newCall, baselineCall], total: 2, page: 1, pageSize: 50 })

    const getPlatformKeyBearer = vi.fn().mockResolvedValue({ bearer: 'sk-test-key' })

    const apis = createMockApis({
      auditApi: { getCalls },
      devApi: { getPlatformKeyBearer },
      platformKeyApi: {
        list: vi.fn().mockResolvedValue({
          items: [
            {
              id: 'plk-1',
              name: '张三-开发调试',
              keyPrefix: 'sk-gate',
              status: 'active',
              scope: 'member',
            },
          ],
          total: 1,
        }),
      },
    })

    vi.mocked(simulateConsume.postChatCompletions).mockResolvedValue()

    const { result } = renderHookWithProviders(() => useSimulateConsumeDialog(true, apis), { apis })

    await waitFor(() => {
      expect(result.current.selectedKeyId).toBe('plk-1')
    })
    expect(getPlatformKeyBearer).toHaveBeenCalledWith('plk-1')

    await act(async () => {
      await result.current.handleSubmit()
    })

    await waitFor(() => {
      expect(result.current.matchedCall?.id).toBe('call-new')
    })

    expect(simulateConsume.postChatCompletions).toHaveBeenCalledWith({
      bearer: 'sk-test-key',
      inputTokens: 12_000_000,
      outputTokens: 8_000_000,
    })
    expect(getCalls).toHaveBeenCalled()
  })

  it('rejects submit when input tokens is zero', async () => {
    const getPlatformKeyBearer = vi.fn().mockResolvedValue({ bearer: 'sk-test-key' })
    const apis = createMockApis({
      devApi: { getPlatformKeyBearer },
      platformKeyApi: {
        list: vi.fn().mockResolvedValue({
          items: [
            {
              id: 'plk-1',
              name: '张三-开发调试',
              keyPrefix: 'sk-gate',
              status: 'active',
              scope: 'member',
            },
          ],
          total: 1,
        }),
      },
    })

    const { result } = renderHookWithProviders(() => useSimulateConsumeDialog(true, apis), { apis })

    await waitFor(() => {
      expect(result.current.selectedKeyId).toBe('plk-1')
    })

    act(() => {
      result.current.setInputTokensText('0')
    })

    await act(async () => {
      await result.current.handleSubmit()
    })

    expect(simulateConsume.postChatCompletions).not.toHaveBeenCalled()
    expect(result.current.error).toBe('Input tokens 须 ≥ 1')
  })
})
