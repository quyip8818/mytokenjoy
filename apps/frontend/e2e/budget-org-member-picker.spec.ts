import { expect, test } from '@playwright/test'

test.describe('创建项目 - 组织树成员选择', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/budget')
    await expect(page.getByRole('heading', { name: '预算管理' })).toBeVisible()
    // Open the create project dialog
    await page.getByRole('button', { name: '创建项目' }).click()
    await expect(page.getByRole('dialog', { name: '创建项目' })).toBeVisible()
  })

  test('opens org tree picker on click', async ({ page }) => {
    await page.getByRole('button', { name: '选择关联成员' }).click()
    // Popover should show org tree with department checkboxes
    await expect(page.getByRole('checkbox', { name: '选择总公司' })).toBeVisible()
  })

  test('selecting a root department selects all recursive members', async ({ page }) => {
    await page.getByRole('button', { name: '选择关联成员' }).click()
    // Click 总公司 text to select all members
    const popover = page.locator('[data-slot="popover-content"]')
    await popover.getByText('总公司').click()
    // Should show "已选 N 人" with N > 0
    await expect(popover.getByText(/已选 \d+ 人/)).toBeVisible()
    // Checkbox should be checked
    await expect(page.getByRole('checkbox', { name: '选择总公司' })).toBeChecked()
  })

  test('expanding a department shows sub-departments', async ({ page }) => {
    await page.getByRole('button', { name: '选择关联成员' }).click()
    const popover = page.locator('[data-slot="popover-content"]')
    // 总公司 is already expanded by default (defaultExpandDepartmentId)
    await expect(popover.getByText('技术部')).toBeVisible()
    await expect(popover.getByText('产品部')).toBeVisible()
  })

  test('expanding a leaf department shows direct members', async ({ page }) => {
    await page.getByRole('button', { name: '选择关联成员' }).click()
    const popover = page.locator('[data-slot="popover-content"]')
    // Expand 技术部
    await popover.getByRole('button', { name: '展开' }).first().click()
    await expect(popover.getByText('后端组')).toBeVisible()
    // Expand 后端组
    await popover.getByRole('button', { name: '展开' }).first().click()
    // Should see member checkboxes
    await expect(popover.getByRole('checkbox', { name: '张三' })).toBeVisible()
  })

  test('deselecting a single member updates count and shows indeterminate', async ({ page }) => {
    await page.getByRole('button', { name: '选择关联成员' }).click()
    const popover = page.locator('[data-slot="popover-content"]')

    // Select all via 总公司
    await popover.getByText('总公司').click()
    const countText = popover.getByText(/已选 \d+ 人/)
    await expect(countText).toBeVisible()
    const initialCount = await countText.textContent()
    const initialNum = parseInt(initialCount!.match(/\d+/)![0])

    // Expand 技术部 → 后端组
    await popover.getByRole('button', { name: '展开' }).first().click()
    await popover.getByRole('button', { name: '展开' }).first().click()

    // Deselect 张三
    await page.getByRole('checkbox', { name: '张三' }).click()

    // Count should decrease by 1
    await expect(popover.getByText(`已选 ${initialNum - 1} 人`)).toBeVisible()

    // 总公司 checkbox should be indeterminate (mixed)
    const rootCheckbox = page.getByRole('checkbox', { name: '选择总公司' })
    await expect(rootCheckbox).not.toBeChecked()
    // Verify aria-checked is "mixed" for indeterminate
    await expect(rootCheckbox).toHaveAttribute('aria-checked', 'mixed')
  })

  test('selecting a mid-level department selects its recursive members', async ({ page }) => {
    await page.getByRole('button', { name: '选择关联成员' }).click()
    const popover = page.locator('[data-slot="popover-content"]')

    // Click 技术部 text to select all tech members
    await popover.getByText('技术部').click()

    // Should show selected count
    await expect(popover.getByText(/已选 \d+ 人/)).toBeVisible()
    await expect(page.getByRole('checkbox', { name: '选择技术部' })).toBeChecked()
  })

  test('deselecting a department removes all its members', async ({ page }) => {
    await page.getByRole('button', { name: '选择关联成员' }).click()
    const popover = page.locator('[data-slot="popover-content"]')

    // Select 技术部
    await popover.getByText('技术部').click()
    await expect(popover.getByText(/已选 \d+ 人/)).toBeVisible()

    // Deselect 技术部 (click again)
    await popover.getByText('技术部').click()

    // Should have 0 selected — footer should disappear
    await expect(popover.getByText(/已选 \d+ 人/)).not.toBeVisible()
  })

  test('search shows matching members with checkboxes', async ({ page }) => {
    await page.getByRole('button', { name: '选择关联成员' }).click()
    const popover = page.locator('[data-slot="popover-content"]')

    await popover.getByPlaceholder('搜索成员...').fill('张三')
    // Wait for search results
    await expect(popover.getByRole('checkbox', { name: '张三' })).toBeVisible({ timeout: 3000 })
  })

  test('mouse wheel scrolls the member list', async ({ page }) => {
    await page.getByRole('button', { name: '选择关联成员' }).click()
    const popover = page.locator('[data-slot="popover-content"]')
    const scrollContainer = popover.locator('.overflow-y-auto')

    // Expand enough nodes to make the list scrollable
    await popover.getByText('总公司').click() // select all first
    // Expand 技术部
    await popover.getByRole('button', { name: '展开' }).first().click()
    // Expand 后端组
    await popover.getByRole('button', { name: '展开' }).first().click()

    // Get initial scroll position
    const scrollBefore = await scrollContainer.evaluate((el) => el.scrollTop)

    // Simulate mouse wheel
    await scrollContainer.hover()
    await page.mouse.wheel(0, 100)
    await page.waitForTimeout(200)

    const scrollAfter = await scrollContainer.evaluate((el) => el.scrollTop)
    expect(scrollAfter).toBeGreaterThan(scrollBefore)
  })
})
