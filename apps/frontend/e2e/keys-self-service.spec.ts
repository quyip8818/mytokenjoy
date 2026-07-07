import { expect, test } from '@playwright/test'

// These tests must run in order: create → rotate → delete
test.describe.configure({ mode: 'serial' })

test.describe('我的 Key - 自管理', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/keys/mine')
    await expect(page.getByRole('banner').getByRole('heading', { name: '我的 Key' })).toBeVisible()
  })

  test('displays quota stats', async ({ page }) => {
    await expect(page.getByText('总额度')).toBeVisible()
    await expect(page.getByText('已使用')).toBeVisible()
    await expect(page.getByText('剩余')).toBeVisible()
  })

  test('creates a new Key via 2-step workflow', async ({ page }) => {
    // Click create button (scope to main to avoid panel footer)
    await page.getByRole('main').getByRole('button', { name: '创建 Key' }).click()

    // Step 1: Basic Info - panel appears
    await expect(page.getByRole('heading', { level: 2, name: '创建 Key' })).toBeVisible()

    // Fill name using placeholder-based locator
    await page.getByRole('textbox', { name: '如：开发调试' }).fill('E2E测试Key')

    // Clear and set quota
    const quotaInput = page.getByRole('spinbutton')
    await quotaInput.clear()
    await quotaInput.fill('100')

    // Next step
    await page.getByRole('button', { name: '下一步' }).click()

    // Step 2: Model Whitelist - select models
    await page.getByRole('button', { name: /选择模型/ }).click()

    // Model picker panel
    await expect(page.getByRole('heading', { level: 2, name: '选择模型' })).toBeVisible()
    await page.getByRole('checkbox').first().check()
    await page.getByRole('button', { name: /确认/ }).click()

    // Submit creation (button in panel footer = contentinfo)
    await page.getByRole('contentinfo').getByRole('button', { name: '创建 Key' }).click()

    // Key reveal panel
    await expect(page.getByRole('heading', { level: 2, name: 'Key 已生成' })).toBeVisible({ timeout: 10_000 })
    await expect(page.getByText('请立即复制保存')).toBeVisible()
    // Panel may be outside viewport due to stacking, use JS click
    await page.getByRole('button', { name: '完成' }).dispatchEvent('click')

    // Verify key appears in table
    await expect(page.getByRole('cell', { name: 'E2E测试Key' }).first()).toBeVisible()
  })

  test('rotates an existing Key', async ({ page }) => {
    // Wait for table to have rows
    await expect(page.locator('tbody tr').first()).toBeVisible()

    // Open action menu on first key row
    await page.locator('tbody tr').first().getByRole('button').click()
    await page.getByRole('menuitem', { name: '重新生成' }).click()

    // Confirm rotation
    await expect(page.getByRole('heading', { level: 2, name: '重新生成 Key' })).toBeVisible()
    await expect(page.getByText('旧 Key 将立即失效')).toBeVisible()
    await page.getByRole('button', { name: '确认重新生成' }).click()

    // Key reveal panel shows new key
    await expect(page.getByRole('heading', { level: 2, name: 'Key 已生成' })).toBeVisible({ timeout: 10_000 })
    // Panel may be outside viewport due to stacking, use JS click
    await page.getByRole('button', { name: '完成' }).dispatchEvent('click')
  })

  test('deletes a Key via confirmation dialog', async ({ page }) => {
    // Wait for table to have rows
    await expect(page.locator('tbody tr').first()).toBeVisible()

    // Open action menu on first key row
    await page.locator('tbody tr').first().getByRole('button').click()
    await page.getByRole('menuitem', { name: '删除' }).click()

    // Confirm in alert dialog
    const dialog = page.locator('[role="alertdialog"]')
    await expect(dialog).toBeVisible()
    await expect(dialog.getByText('此操作不可撤销')).toBeVisible()
    await dialog.getByRole('button', { name: '删除' }).click()

    // Dialog closes after action
    await expect(dialog).toBeHidden()
  })
})
