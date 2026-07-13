import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import { buildChatCompletionResponse, parseDevUsage } from '../src/server.mjs'

describe('parseDevUsage', () => {
  it('uses dev_usage when provided', () => {
    const usage = parseDevUsage({
      dev_usage: { prompt_tokens: 100, completion_tokens: 50 },
    })
    assert.equal(usage.prompt_tokens, 100)
    assert.equal(usage.completion_tokens, 50)
  })

  it('falls back to defaults when dev_usage missing', () => {
    const usage = parseDevUsage({})
    assert.equal(usage.prompt_tokens, 10)
    assert.equal(usage.completion_tokens, 5)
  })
})

describe('buildChatCompletionResponse', () => {
  it('returns openai-compatible shape', () => {
    const payload = buildChatCompletionResponse({ prompt_tokens: 12, completion_tokens: 8 })
    assert.equal(payload.object, 'chat.completion')
    assert.equal(payload.usage.total_tokens, 20)
    assert.equal(payload.choices[0].message.content, 'ok')
  })
})
