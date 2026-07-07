import { expect, test } from '@playwright/test'

test.describe.configure({ mode: 'serial' })

test.describe('我的 Key - 自管理', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/keys/mine')
    await expect(page.getByRole('heading', { name: '我的 Key' })).toBeVisible()
  })

  test('displays quota stats', async ({ page }) => {
    await expect(page.getByText('总额度')).toBeVisible()
    await expect(page.getByText('已使用')).toBeVisible()
    await expect(page.getByText('剩余')).toBeVisible()
  })

  test('creates a new Key', async ({ page }) => {
    await page.getByRole('main').getByRole('button', { name: '创建 Key' }).click()
    await expect(page.getByRole('heading', { level: 2, name: '创建 Key' })).toBeVisible()

    await page.getByRole('textbox', { name: '如：开发调试' }).fill('E2E测试Key')
    const quotaInput = page.getByRole('spinbutton')
    await quotaInput.clear()
    await quotaInput.fill('100')
    await page.getByRole('button', { name: '下一步' }).click()

    // Model selection
    await page.getByRole('button', { name: /选择模型/ }).click()
    await expect(page.getByRole('heading', { level: 2, name: '选择模型' })).toBeVisible()
    await page.getByRole('checkbox').first().check()
    await page.getByRole('button', { name: /确认/ }).click()

    // Submit
    await page.getByRole('contentinfo').getByRole('button', { name: '创建 Key' }).click()
    await expect(page.getByRole('heading', { level: 2, name: 'Key 已生成' })).toBeVisible({ timeout: 10_000 })
    await page.getByRole('button', { name: '完成' }).dispatchEvent('click')
    await expect(page.getByRole('cell', { name: 'E2E测试Key' }).first()).toBeVisible()
  })

  test('rotates an existing Key', async ({ page }) => {
    await expect(page.locator('tbody tr').first()).toBeVisible()
    await page.locator('tbody tr').first().getByRole('button').click()
    await page.getByRole('menuitem', { name: '重新生成' }).click()
    await expect(page.getByRole('heading', { level: 2, name: '重新生成 Key' })).toBeVisible()
    await page.getByRole('button', { name: '确认重新生成' }).click()
    await expect(page.getByRole('heading', { level: 2, name: 'Key 已生成' })).toBeVisible({ timeout: 10_000 })
    await page.getByRole('button', { name: '完成' }).dispatchEvent('click')
  })

  test('deletes a Key', async ({ page }) => {
    await expect(page.locator('tbody tr').first()).toBeVisible()
    await page.locator('tbody tr').first().getByRole('button').click()
    await page.getByRole('menuitem', { name: '删除' }).click()
    const dialog = page.locator('[role="alertdialog"]')
    await expect(dialog).toBeVisible()
    await dialog.getByRole('button', { name: '删除' }).click()
    await expect(dialog).toBeHidden()
  })
})
