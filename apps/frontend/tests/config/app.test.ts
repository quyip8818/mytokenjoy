import { describe, expect, it } from 'vitest'
import { API_BASE_PATH } from '@/config/app'

describe('app config', () => {
  it('uses /api as the default API base path', () => {
    expect(API_BASE_PATH.endsWith('/api')).toBe(true)
  })
})
