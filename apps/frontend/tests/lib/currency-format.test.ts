import { describe, expect, it } from 'vitest'
import { currencySymbol, formatCurrencyAmount } from '@/lib/currency-format'

describe('currency-format', () => {
  it('maps common currency codes to symbols', () => {
    expect(currencySymbol('CNY')).toBe('¥')
    expect(currencySymbol('USD')).toBe('$')
    expect(currencySymbol('EUR')).toBe('€')
    expect(currencySymbol('JPY')).toBe('JPY ')
  })

  it('formats display amounts with currency symbol', () => {
    expect(formatCurrencyAmount(10, 'USD')).toBe('$10')
    expect(formatCurrencyAmount(10.25, 'CNY')).toBe('¥10.25')
  })
})
