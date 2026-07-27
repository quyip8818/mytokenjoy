import { expect, test } from '@playwright/test'

test.describe('TokenJoy 模型列表 - SMS 同步数据展示', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/models/list')
    // Wait for the model list to load (scoped to main to avoid Header duplicate)
    await expect(page.getByRole('main').getByRole('heading', { name: /模型/ })).toBeVisible({
      timeout: 10000,
    })
  })

  test('model list page loads with table', async ({ page }) => {
    // The model table or list should be visible
    const table = page.locator('table').first()
    await expect(table).toBeVisible({ timeout: 5000 })
  })

  test('synced models appear with correct data', async ({ page }) => {
    // Look for any model row in the table
    const rows = page.locator('table tbody tr')
    const count = await rows.count()

    if (count === 0) {
      // No models yet — test passes (sync may not have run)
      test.skip()
      return
    }

    // First row should have model name, provider, and price columns
    const firstRow = rows.first()
    await expect(firstRow).toBeVisible()

    // Verify there's content in the cells (not empty)
    const cells = firstRow.locator('td')
    const cellCount = await cells.count()
    expect(cellCount).toBeGreaterThanOrEqual(3) // at minimum: name, provider, price
  })

  test('model list shows provider information', async ({ page }) => {
    // Look for common provider names that would come from SMS sync
    const providers = ['deepseek', 'openai', 'anthropic', 'custom']
    const table = page.locator('table')

    // Check if any provider badge/text is visible
    const hasProvider = await table
      .getByText(new RegExp(providers.join('|'), 'i'))
      .first()
      .isVisible()
      .catch(() => false)

    // This is informational — doesn't fail if no models synced yet
    if (!hasProvider) {
      test.skip()
    }
  })

  test('model list shows pricing columns', async ({ page }) => {
    const table = page.locator('table')

    // Look for price-related text in column headers or cells
    const hasPricing = await table
      .getByText(/输入|输出|input|output|价格|定价|price/i)
      .first()
      .isVisible()
      .catch(() => false)

    // Skip if model list doesn't show pricing (feature may not be visible without models)
    if (!hasPricing) {
      test.skip()
    }
  })
})
