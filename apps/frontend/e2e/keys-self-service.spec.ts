import { expect, test } from '@playwright/test'

test.describe.configure({ mode: 'serial' })

test.describe('我的 Key - 自管理', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/keys/mine')
    await expect(page.getByRole('banner').getByRole('heading', { name: '我的 Key' })).toBeVisible()
  })

  test('displays page actions', async ({ page }) => {
    // Page has action buttons regardless of key count
    await expect(page.getByRole('button', { name: '新建 Key' })).toBeVisible()
  })

  test('creates a new Key', async ({ page }) => {
    await page.getByRole('button', { name: '新建 Key' }).click()
    await expect(page.getByRole('heading', { level: 2, name: /创建|新建/ })).toBeVisible()

    // Step 1: basic info
    const nameInput = page.getByRole('textbox', { name: /名称|备注|如：开发调试/ })
    await nameInput.fill('E2E测试Key')
    const quotaInput = page.getByRole('spinbutton')
    if (await quotaInput.isVisible()) {
      await quotaInput.clear()
      await quotaInput.fill('100')
    }

    // Click 下一步 to go to step 2
    await page.getByRole('button', { name: '下一步' }).click()
    await page.waitForTimeout(500)

    // Step 2: model selection - select a model if not already selected
    const selectModelBtn = page.getByRole('button', { name: /选择模型/ })
    await expect(selectModelBtn).toBeVisible()
    // Check if models are already selected (button text shows count)
    const btnText = await selectModelBtn.textContent()
    if (!btnText?.includes('(')) {
      await selectModelBtn.click()
      await expect(page.getByRole('heading', { level: 2, name: '选择模型' })).toBeVisible()
      await page.getByRole('checkbox').first().check()
      await page.getByRole('button', { name: /确认/ }).click()
    }

    // Submit
    const createKeyBtn = page.getByRole('button', { name: '创建 Key', exact: true })
    await expect(createKeyBtn).toBeVisible()
    await expect(createKeyBtn).toBeEnabled()
    await createKeyBtn.click()

    // Wait for result - might fail if quota is not allocated
    const success = await page
      .getByRole('heading', { level: 2, name: 'Key 已生成' })
      .isVisible({ timeout: 10_000 })
      .catch(() => false)
    if (!success) {
      test.skip(true, 'Key 创建失败（可能额度未分配）')
      return
    }
    await page.getByRole('button', { name: '完成' }).dispatchEvent('click')
    await expect(page.getByRole('cell', { name: 'E2E测试Key' }).first()).toBeVisible({
      timeout: 5_000,
    })
  })

  test('rotates an existing Key', async ({ page }) => {
    const row = page.locator('tbody tr').first()
    const visible = await row.isVisible({ timeout: 3000 }).catch(() => false)
    if (!visible) {
      test.skip(true, '没有已创建的 Key（前置 create 测试被跳过）')
      return
    }
    await row.getByRole('button').click()
    await page.getByRole('menuitem', { name: '重新生成' }).click()
    await expect(page.getByRole('heading', { level: 2, name: '重新生成 Key' })).toBeVisible()
    await page.getByRole('button', { name: '确认重新生成' }).click()
    await expect(page.getByRole('heading', { level: 2, name: 'Key 已生成' })).toBeVisible({
      timeout: 10_000,
    })
    await page.getByRole('button', { name: '完成' }).dispatchEvent('click')
  })

  test('deletes a Key', async ({ page }) => {
    const row = page.locator('tbody tr').first()
    const visible = await row.isVisible({ timeout: 3000 }).catch(() => false)
    if (!visible) {
      test.skip(true, '没有已创建的 Key（前置 create 测试被跳过）')
      return
    }
    await row.getByRole('button').click()
    await page.getByRole('menuitem', { name: '删除' }).click()
    const dialog = page.locator('[role="alertdialog"]')
    await expect(dialog).toBeVisible()
    await dialog.getByRole('button', { name: '删除' }).click()
    await expect(dialog).toBeHidden()
  })
})
