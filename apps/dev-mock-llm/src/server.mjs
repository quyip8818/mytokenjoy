import http from 'node:http'
import { pathToFileURL } from 'node:url'

const DEFAULT_PORT = 8765
const DEFAULT_PROMPT_TOKENS = 10
const DEFAULT_COMPLETION_TOKENS = 5

export function parseDevUsage(body) {
  const raw = body?.dev_usage
  const prompt = Number(raw?.prompt_tokens ?? DEFAULT_PROMPT_TOKENS)
  const completion = Number(raw?.completion_tokens ?? DEFAULT_COMPLETION_TOKENS)
  return {
    prompt_tokens:
      Number.isFinite(prompt) && prompt >= 0 ? Math.floor(prompt) : DEFAULT_PROMPT_TOKENS,
    completion_tokens:
      Number.isFinite(completion) && completion >= 0
        ? Math.floor(completion)
        : DEFAULT_COMPLETION_TOKENS,
  }
}

export function buildChatCompletionResponse(usage) {
  const total = usage.prompt_tokens + usage.completion_tokens
  return {
    id: 'chatcmpl-local-test-model',
    object: 'chat.completion',
    created: Math.floor(Date.now() / 1000),
    model: 'local-test-model',
    choices: [
      {
        index: 0,
        message: { role: 'assistant', content: 'ok' },
        finish_reason: 'stop',
      },
    ],
    usage: {
      prompt_tokens: usage.prompt_tokens,
      completion_tokens: usage.completion_tokens,
      total_tokens: total,
    },
  }
}

function readJson(req) {
  return new Promise((resolve, reject) => {
    const chunks = []
    req.on('data', (chunk) => chunks.push(chunk))
    req.on('end', () => {
      const raw = Buffer.concat(chunks).toString('utf8')
      if (!raw.trim()) {
        resolve({})
        return
      }
      try {
        resolve(JSON.parse(raw))
      } catch (err) {
        reject(err)
      }
    })
    req.on('error', reject)
  })
}

function sendJson(res, status, payload) {
  const body = JSON.stringify(payload)
  res.writeHead(status, { 'Content-Type': 'application/json' })
  res.end(body)
}

export function createServer() {
  return http.createServer(async (req, res) => {
    if (req.method === 'GET' && req.url === '/healthz') {
      sendJson(res, 200, { status: 'ok' })
      return
    }

    if (req.method === 'POST' && req.url === '/v1/chat/completions') {
      try {
        const body = await readJson(req)
        const usage = parseDevUsage(body)
        sendJson(res, 200, buildChatCompletionResponse(usage))
      } catch {
        sendJson(res, 400, { error: { message: 'invalid json' } })
      }
      return
    }

    sendJson(res, 404, { error: { message: 'not found' } })
  })
}

export function startServer(port = DEFAULT_PORT) {
  const server = createServer()
  server.listen(port, '127.0.0.1', () => {
    console.log(`[dev-mock-llm] listening on http://127.0.0.1:${port}`)
  })
  return server
}

const isMain = process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href
if (isMain) {
  const port = Number(process.env.PORT ?? DEFAULT_PORT)
  startServer(Number.isFinite(port) ? port : DEFAULT_PORT)
}
