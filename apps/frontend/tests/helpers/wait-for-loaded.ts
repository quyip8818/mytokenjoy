import { waitFor } from '@testing-library/react'
import { expect } from 'vitest'

export async function waitForLoaded(result: { current: Record<string, unknown> }, key: string) {
  await waitFor(() => {
    expect(result.current[key]).toBe(false)
  })
}
