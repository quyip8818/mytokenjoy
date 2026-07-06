import { expect, test } from '@playwright/test'
import { loginAsAdmin } from './helpers/auth'

test('navigates to org structure and shows seed department', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/')

  await page.getByRole('link', { name: '组织架构' }).click()
  await expect(page).toHaveURL(/\/org\/structure$/)

  await expect(page.getByRole('main').getByText('总公司').first()).toBeVisible()
  await expect(page.getByRole('heading', { name: 'TokenJoy' })).toBeVisible()
})
