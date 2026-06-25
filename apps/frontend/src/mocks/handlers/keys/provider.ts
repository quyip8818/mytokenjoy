import { http, HttpResponse, delay } from 'msw'
import { API_BASE_PATH } from '@/config/app'
import { DEMO_TODAY } from '@/mocks/lib/demo-clock'
import { mockProviderKeys } from '../../data'

export const providerKeysHandlers = [
  http.get(`${API_BASE_PATH}/keys/provider`, () => {
    return HttpResponse.json(mockProviderKeys)
  }),
  http.post(`${API_BASE_PATH}/keys/provider`, async ({ request }) => {
    await delay(500)
    const body = (await request.json()) as Record<string, unknown>
    return HttpResponse.json({
      id: `pk-${Date.now()}`,
      ...body,
      keyPrefix: 'sk-new...',
      status: 'active',
      balance: null,
      lastUsed: null,
      createdAt: DEMO_TODAY,
      rotateEnabled: false,
    })
  }),
  http.put(`${API_BASE_PATH}/keys/provider/:id/toggle`, async () => {
    await delay(300)
    return HttpResponse.json(null, { status: 200 })
  }),
  http.post(`${API_BASE_PATH}/keys/provider/:id/rotate`, async ({ params }) => {
    await delay(1000)
    const idx = mockProviderKeys.findIndex((k) => k.id === params.id)
    if (idx === -1) {
      return HttpResponse.json({ message: 'Not found' }, { status: 404 })
    }
    const updated = {
      ...mockProviderKeys[idx],
      keyPrefix: `sk-rot-${Date.now().toString(36)}...`,
      lastUsed: new Date().toISOString().slice(0, 16).replace('T', ' '),
    }
    mockProviderKeys[idx] = updated
    return HttpResponse.json(updated)
  }),
  http.delete(`${API_BASE_PATH}/keys/provider/:id`, () => {
    return HttpResponse.json(null, { status: 200 })
  }),
]
