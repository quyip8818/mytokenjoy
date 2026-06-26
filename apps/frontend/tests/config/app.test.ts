import { describe, expect, it } from 'vitest'
import { resolveUseMocks } from '@/config/app'

describe('resolveUseMocks', () => {
  it('enables mocks in dev when env is unset', () => {
    expect(resolveUseMocks({ DEV: true })).toBe(true)
  })

  it('disables mocks in dev when VITE_ENABLE_MOCKS is false', () => {
    expect(resolveUseMocks({ DEV: true, VITE_ENABLE_MOCKS: 'false' })).toBe(false)
  })

  it('enables mocks in production when VITE_ENABLE_MOCKS is true', () => {
    expect(resolveUseMocks({ DEV: false, VITE_ENABLE_MOCKS: 'true' })).toBe(true)
  })

  it('disables mocks in production when VITE_ENABLE_MOCKS is unset', () => {
    expect(resolveUseMocks({ DEV: false })).toBe(false)
  })

  it('disables mocks in production when VITE_ENABLE_MOCKS is false', () => {
    expect(resolveUseMocks({ DEV: false, VITE_ENABLE_MOCKS: 'false' })).toBe(false)
  })
})
