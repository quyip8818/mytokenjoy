import { describe, expect, it, vi, beforeEach } from 'vitest'
import { notificationApi } from '@/api/notifications'

// Mock the client module
vi.mock('@/api/client', () => ({
  request: vi.fn(),
  buildQuery: vi.fn((params: object) => {
    const search = new URLSearchParams()
    for (const [key, value] of Object.entries(params)) {
      if (value === undefined || value === null || value === '') continue
      search.set(key, String(value))
    }
    const qs = search.toString()
    return qs ? `?${qs}` : ''
  }),
}))

import { request } from '@/api/client'
const mockRequest = vi.mocked(request)

describe('notificationApi', () => {
  beforeEach(() => {
    mockRequest.mockReset()
  })

  it('list calls /notifications with query params', async () => {
    mockRequest.mockResolvedValue([])
    await notificationApi.list({ limit: 10, offset: 5 })
    expect(mockRequest).toHaveBeenCalledWith('/notifications?limit=10&offset=5')
  })

  it('list calls /notifications without params', async () => {
    mockRequest.mockResolvedValue([])
    await notificationApi.list()
    expect(mockRequest).toHaveBeenCalledWith('/notifications')
  })

  it('unreadCount calls /notifications/unread-count', async () => {
    mockRequest.mockResolvedValue({ count: 5 })
    const result = await notificationApi.unreadCount()
    expect(mockRequest).toHaveBeenCalledWith('/notifications/unread-count')
    expect(result).toEqual({ count: 5 })
  })

  it('markRead calls PATCH with id', async () => {
    mockRequest.mockResolvedValue(undefined)
    await notificationApi.markRead('ntf-123')
    expect(mockRequest).toHaveBeenCalledWith('/notifications/ntf-123/read', { method: 'PATCH' })
  })

  it('markAllRead calls POST', async () => {
    mockRequest.mockResolvedValue(undefined)
    await notificationApi.markAllRead()
    expect(mockRequest).toHaveBeenCalledWith('/notifications/read-all', { method: 'POST' })
  })

  it('getCapabilities calls correct endpoint', async () => {
    const caps = {
      channels: ['in_app'],
      emailConfigured: false,
      smsConfigured: false,
      inAppConfigured: true,
    }
    mockRequest.mockResolvedValue(caps)
    const result = await notificationApi.getCapabilities()
    expect(mockRequest).toHaveBeenCalledWith('/notifications/capabilities')
    expect(result.inAppConfigured).toBe(true)
  })

  it('getPreferences calls correct endpoint', async () => {
    mockRequest.mockResolvedValue({ preferences: [] })
    await notificationApi.getPreferences()
    expect(mockRequest).toHaveBeenCalledWith('/notifications/preferences')
  })

  it('updatePreferences sends PUT with body', async () => {
    mockRequest.mockResolvedValue(undefined)
    const data = { preferences: [{ category: 'budget_alert', channel: 'email', enabled: false }] }
    await notificationApi.updatePreferences(data)
    expect(mockRequest).toHaveBeenCalledWith('/notifications/preferences', {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  })

  it('resetPreferences sends POST', async () => {
    mockRequest.mockResolvedValue(undefined)
    await notificationApi.resetPreferences()
    expect(mockRequest).toHaveBeenCalledWith('/notifications/preferences/reset', { method: 'POST' })
  })
})
