import { expect, test } from '@playwright/test'

test.describe('数据源', () => {
  test('shows platform selection', async ({ page }) => {
    await page.goto('/org/data-source')
    await expect(page.getByRole('heading', { name: '数据源' })).toBeVisible()
    await expect(page.getByRole('radio', { name: /飞书/ })).toBeVisible()
    await expect(page.getByRole('radio', { name: /钉钉/ })).toBeVisible()
    await expect(page.getByRole('radio', { name: /企业微信/ })).toBeVisible()
  })
})
