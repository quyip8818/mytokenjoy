import { displayToPoints, formatDisplayCurrency } from '@/lib/points'

/** Catalog type shared by TokenJoy seed, NewAPI channel, and Gateway calls. */
export const LOCAL_TEST_MODEL = 'local-test-model'

export const DEFAULT_INPUT_TOKENS = 12_000_000
export const DEFAULT_OUTPUT_TOKENS = 8_000_000

export const PLATFORM_KEY_ID_SESSION_KEY = 'tokenjoy.local-test-model.platformKeyId'

/** Per-1M-token display prices — same tier as gpt-4o-mini seed. */
const INPUT_PRICE_DISPLAY = 0.15
const OUTPUT_PRICE_DISPLAY = 0.6
const TOKENS_PER_MILLION = 1_000_000

export function estimateConsumePoints(inputTokens: number, outputTokens: number): number {
  const inputCost = (inputTokens / TOKENS_PER_MILLION) * displayToPoints(INPUT_PRICE_DISPLAY)
  const outputCost = (outputTokens / TOKENS_PER_MILLION) * displayToPoints(OUTPUT_PRICE_DISPLAY)
  return Math.round(inputCost + outputCost)
}

export function formatEstimatedConsume(inputTokens: number, outputTokens: number): string {
  return formatDisplayCurrency(estimateConsumePoints(inputTokens, outputTokens))
}
