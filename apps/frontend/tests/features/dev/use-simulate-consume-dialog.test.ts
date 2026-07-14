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

const activeKey = {
  id: 'plk-1',
  name: '张三-开发调试',
  keyPrefix: 'sk-gate',
  status: 'active' as const,
  scope: 'member' as const,
}

describe('useSimulateConsumeDialog', () => {
  beforeEach(() => {
    vi.mocked(simulateConsume.postChatCompletions).mockReset()
    sessionStorage.clear()
  })

  it('submits gateway call then invokes onSuccess', async () => {
    const getPlatformKeyBearer = vi.fn().mockResolvedValue({ bearer: 'sk-test-key' })
    const onSuccess = vi.fn()
    const apis = createMockApis({
      devApi: { getPlatformKeyBearer },
      platformKeyApi: {
        list: vi.fn().mockResolvedValue({ items: [activeKey], total: 1 }),
      },
    })
    vi.mocked(simulateConsume.postChatCompletions).mockResolvedValue()

    const { result } = renderHookWithProviders(
      () => useSimulateConsumeDialog(true, apis, onSuccess),
      { apis },
    )

    await waitFor(() => {
      expect(result.current.selectedKeyId).toBe('plk-1')
    })

    await act(async () => {
      await result.current.handleSubmit()
    })

    expect(simulateConsume.postChatCompletions).toHaveBeenCalledWith({
      bearer: 'sk-test-key',
      inputTokens: 12_000_000,
      outputTokens: 8_000_000,
    })
    expect(onSuccess).toHaveBeenCalledTimes(1)
    expect(result.current.error).toBeNull()
    expect(result.current.busy).toBe(false)
  })

  it('rejects submit when input tokens is zero', async () => {
    const getPlatformKeyBearer = vi.fn().mockResolvedValue({ bearer: 'sk-test-key' })
    const onSuccess = vi.fn()
    const apis = createMockApis({
      devApi: { getPlatformKeyBearer },
      platformKeyApi: {
        list: vi.fn().mockResolvedValue({ items: [activeKey], total: 1 }),
      },
    })

    const { result } = renderHookWithProviders(
      () => useSimulateConsumeDialog(true, apis, onSuccess),
      { apis },
    )

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
    expect(onSuccess).not.toHaveBeenCalled()
    expect(result.current.error).toBe('Input tokens 须 ≥ 1')
  })
})
