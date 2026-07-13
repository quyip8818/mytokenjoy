import type { AppApis } from '@/api/app-apis'
import type { CallLog } from '@/api/types'
import { LOCAL_TEST_MODEL, POLL_INTERVAL_MS, POLL_MAX_ATTEMPTS } from './constants'

export class GatewayClientError extends Error {
  readonly status: number
  readonly body: string

  constructor(status: number, body: string) {
    super(body || `Gateway request failed (${status})`)
    this.name = 'GatewayClientError'
    this.status = status
    this.body = body
  }
}

export async function postChatCompletions(input: {
  bearer: string
  inputTokens: number
  outputTokens: number
}): Promise<void> {
  const response = await fetch('/v1/chat/completions', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${input.bearer}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      model: LOCAL_TEST_MODEL,
      messages: [{ role: 'user', content: 'tokenjoy local-test-model' }],
      max_tokens: 1,
      dev_usage: {
        prompt_tokens: input.inputTokens,
        completion_tokens: input.outputTokens,
      },
    }),
  })

  const body = await response.text()
  if (!response.ok) {
    throw new GatewayClientError(response.status, body)
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms)
  })
}

async function listEchoCalls(auditApi: AppApis['auditApi']): Promise<CallLog[]> {
  const page = await auditApi.getCalls({ page: 1, pageSize: 50, model: LOCAL_TEST_MODEL })
  return page.items
}

export async function fetchBaselineCallIds(auditApi: AppApis['auditApi']): Promise<Set<string>> {
  const items = await listEchoCalls(auditApi)
  return new Set(items.map((item) => item.id))
}

export async function pollForNewCall(
  auditApi: AppApis['auditApi'],
  baselineIds: Set<string>,
  inputTokens: number,
  outputTokens: number,
): Promise<CallLog | null> {
  for (let attempt = 0; attempt < POLL_MAX_ATTEMPTS; attempt += 1) {
    const match = (await listEchoCalls(auditApi)).find(
      (call) =>
        !baselineIds.has(call.id) &&
        call.model === LOCAL_TEST_MODEL &&
        call.inputTokens === inputTokens &&
        call.outputTokens === outputTokens,
    )
    if (match) {
      return match
    }
    await sleep(POLL_INTERVAL_MS)
  }
  return null
}
