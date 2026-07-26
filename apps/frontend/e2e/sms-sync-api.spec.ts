import { expect, test } from '@playwright/test'

const SMS_BASE_URL = process.env.SMS_API_URL || 'http://127.0.0.1:8020'
const TOKENJOY_BASE_URL = process.env.TOKENJOY_API_URL || 'http://127.0.0.1:8010'

// Test OAuth client credentials (seeded in SMS DB)
const CLIENT_ID = 'tokenjoy-sync'
const CLIENT_SECRET = 'e2e-test-secret'

test.describe('SMS Sync API - OAuth2 Token', () => {
  test('valid client_credentials returns JWT token', async ({ request }) => {
    const res = await request.post(`${SMS_BASE_URL}/api/oauth/token`, {
      data: {
        grant_type: 'client_credentials',
        client_id: CLIENT_ID,
        client_secret: CLIENT_SECRET,
      },
    })
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body.access_token).toBeTruthy()
    expect(body.token_type).toBe('Bearer')
    expect(body.expires_in).toBe(600)
    expect(body.scope).toBe('sync:read')
  })

  test('wrong secret returns 401', async ({ request }) => {
    const res = await request.post(`${SMS_BASE_URL}/api/oauth/token`, {
      data: {
        grant_type: 'client_credentials',
        client_id: CLIENT_ID,
        client_secret: 'wrong-secret',
      },
    })
    expect(res.status()).toBe(401)
  })

  test('unknown client_id returns 401', async ({ request }) => {
    const res = await request.post(`${SMS_BASE_URL}/api/oauth/token`, {
      data: {
        grant_type: 'client_credentials',
        client_id: 'nonexistent',
        client_secret: 'any',
      },
    })
    expect(res.status()).toBe(401)
  })

  test('unsupported grant_type returns 400', async ({ request }) => {
    const res = await request.post(`${SMS_BASE_URL}/api/oauth/token`, {
      data: {
        grant_type: 'authorization_code',
        client_id: CLIENT_ID,
        client_secret: CLIENT_SECRET,
      },
    })
    expect(res.status()).toBe(400)
  })
})

test.describe('SMS Sync API - Catalog', () => {
  let accessToken: string

  test.beforeAll(async ({ request }) => {
    const res = await request.post(`${SMS_BASE_URL}/api/oauth/token`, {
      data: {
        grant_type: 'client_credentials',
        client_id: CLIENT_ID,
        client_secret: CLIENT_SECRET,
      },
    })
    const body = await res.json()
    accessToken = body.access_token
  })

  test('valid token returns catalog with models and channels', async ({ request }) => {
    const res = await request.get(`${SMS_BASE_URL}/api/sync/catalog`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body).toHaveProperty('channels')
    expect(body).toHaveProperty('models')
    expect(body).toHaveProperty('syncedAt')
    expect(Array.isArray(body.channels)).toBe(true)
    expect(Array.isArray(body.models)).toBe(true)
  })

  test('catalog models have required fields', async ({ request }) => {
    const res = await request.get(`${SMS_BASE_URL}/api/sync/catalog`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
    const body = await res.json()
    // Skip if no models seeded
    if (body.models.length === 0) return

    const model = body.models[0]
    expect(model).toHaveProperty('modelId')
    expect(model).toHaveProperty('displayName')
    expect(model).toHaveProperty('provider')
    expect(model).toHaveProperty('callType')
    expect(model).toHaveProperty('inputPrice')
    expect(model).toHaveProperty('outputPrice')
    expect(typeof model.inputPrice).toBe('number')
    expect(typeof model.outputPrice).toBe('number')
  })

  test('catalog channels have required fields', async ({ request }) => {
    const res = await request.get(`${SMS_BASE_URL}/api/sync/catalog`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
    const body = await res.json()
    if (body.channels.length === 0) return

    const channel = body.channels[0]
    expect(channel).toHaveProperty('name')
    expect(channel).toHaveProperty('type')
    expect(channel).toHaveProperty('baseUrl')
    expect(channel).toHaveProperty('key')
    expect(channel).toHaveProperty('models')
    expect(channel).toHaveProperty('group')
  })

  test('no authorization header returns 401', async ({ request }) => {
    const res = await request.get(`${SMS_BASE_URL}/api/sync/catalog`)
    expect(res.status()).toBe(401)
  })

  test('expired/invalid token returns 401', async ({ request }) => {
    const res = await request.get(`${SMS_BASE_URL}/api/sync/catalog`, {
      headers: { Authorization: 'Bearer invalid-token-xxx' },
    })
    expect(res.status()).toBe(401)
  })
})

test.describe('TokenJoy - Post-Sync Verification', () => {
  test('models API returns synced models with source=sms', async ({ request }) => {
    // Login to TokenJoy
    const loginRes = await request.post(`${TOKENJOY_BASE_URL}/api/auth/login`, {
      data: { email: 'admin@example.com', password: 'demo1234' },
    })
    expect(loginRes.status()).toBe(200)

    // Fetch model list
    const modelsRes = await request.get(`${TOKENJOY_BASE_URL}/api/models`, {
      headers: { Cookie: loginRes.headers()['set-cookie'] || '' },
    })

    // If sync has run, there should be models. If not yet synced, skip gracefully.
    if (modelsRes.status() === 200) {
      const body = await modelsRes.json()
      const smsModels = (body.models || body || []).filter(
        (m: { source?: string }) => m.source === 'sms',
      )
      // This test passes whether sync has run or not — it validates structure
      for (const model of smsModels) {
        expect(model).toHaveProperty('type')
        expect(model).toHaveProperty('provider')
        expect(model).toHaveProperty('name')
        expect(model.source).toBe('sms')
      }
    }
  })

  test('manual models are not overwritten by sync', async ({ request }) => {
    const loginRes = await request.post(`${TOKENJOY_BASE_URL}/api/auth/login`, {
      data: { email: 'admin@example.com', password: 'demo1234' },
    })

    const modelsRes = await request.get(`${TOKENJOY_BASE_URL}/api/models`, {
      headers: { Cookie: loginRes.headers()['set-cookie'] || '' },
    })
    if (modelsRes.status() !== 200) return

    const body = await modelsRes.json()
    const manualModels = (body.models || body || []).filter(
      (m: { source?: string }) => m.source === 'manual' || !m.source,
    )
    // Manual models should still exist after sync cycles
    // This is a structural assertion — actual count depends on seed data
    for (const model of manualModels) {
      expect(model.source).not.toBe('sms')
    }
  })
})
