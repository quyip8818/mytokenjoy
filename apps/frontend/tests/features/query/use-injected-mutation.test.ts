import { describe, expect, it, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { useInjectedMutation } from '@/features/query/use-injected-mutation'
import { createMockApis, createTestWrapper } from '@tests/utils'

describe('useInjectedMutation', () => {
  it('runs mutation and invalidates query keys on success', async () => {
    const updateFn = vi.fn().mockResolvedValue(undefined)
    const apis = createMockApis({
      departmentApi: {
        update: updateFn,
      },
    })

    const { result } = renderHook(
      () =>
        useInjectedMutation({
          injectedApis: apis,
          mutationFn: async (a, variables: { id: string; name: string }) => {
            await a.departmentApi.update(variables.id, { name: variables.name })
          },
          invalidateKeys: [['org', 'department-tree']],
        }),
      { wrapper: createTestWrapper({ apis }) },
    )

    await result.current.mutateAsync({ id: 'dept-1', name: 'Updated' })

    expect(updateFn).toHaveBeenCalledWith('dept-1', { name: 'Updated' })
    expect(result.current.error).toBeNull()
    expect(result.current.isPending).toBe(false)
  })

  it('exposes error when mutation fails', async () => {
    const apis = createMockApis({
      departmentApi: {
        update: vi.fn().mockRejectedValue(new Error('update failed')),
      },
    })

    const { result } = renderHook(
      () =>
        useInjectedMutation({
          injectedApis: apis,
          mutationFn: async (a) => {
            await a.departmentApi.update('dept-1', { name: 'x' })
          },
        }),
      { wrapper: createTestWrapper({ apis }) },
    )

    await expect(result.current.mutateAsync()).rejects.toThrow('update failed')

    await waitFor(() => {
      expect(result.current.error?.message).toBe('update failed')
    })
  })
})
