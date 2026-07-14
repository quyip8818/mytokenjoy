import { LOCAL_TEST_MODEL } from './constants'

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
