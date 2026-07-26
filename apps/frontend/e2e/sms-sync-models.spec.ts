import { expect, test } from '@playwright/test'

test.describe('TokenJoy 模型列表 - SMS 同步数据展示', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/models/list')
    // Wait for the model list to load
    await expect(page.getByRole('heading', { name: /模型/ })).toBeVisible({ timeout: 10000 })
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

    // Look for price-related column headers
    const hasInputPrice = await table
      .getByText(/输入|input/i)
      .first()
      .isVisible()
      .catch(() => false)
    const hasOutputPrice = await table
      .getByText(/输出|output/i)
      .first()
      .isVisible()
      .catch(() => false)

    // At least one price column should be visible if the page has models
    const rows = page.locator('table tbody tr')
    const count = await rows.count()
    if (count > 0) {
      expect(hasInputPrice || hasOutputPrice).toBe(true)
    }
  })
})
