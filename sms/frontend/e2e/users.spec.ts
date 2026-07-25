import { test, expect } from './fixtures'

test.describe('用户管理页面', () => {
  test('页面正常加载并显示用户列表', async ({ authedPage: page }) => {
    await page.goto('/system/users')
    await expect(page.locator('table')).toBeVisible()
    await expect(page.getByRole('cell', { name: 'admin', exact: true })).toBeVisible()
  })

  test('显示用户角色标签', async ({ authedPage: page }) => {
    await page.goto('/system/users')
    await expect(page.locator('table')).toBeVisible()
    // admin 用户行应该显示管理员角色 badge
    await expect(page.getByText('管理员', { exact: true })).toBeVisible()
  })

  test('搜索用户', async ({ authedPage: page }) => {
    await page.goto('/system/users')
    await page.getByPlaceholder('用户名 / 姓名').fill('admin')
    await expect(page.getByRole('cell', { name: 'admin', exact: true })).toBeVisible()
    // 搜索不存在的用户
    await page.getByPlaceholder('用户名 / 姓名').fill('nonexistent_xyz')
    await expect(page.locator('text=暂无数据')).toBeVisible()
  })

  test('新建用户对话框', async ({ authedPage: page }) => {
    await page.goto('/system/users')
    await page.getByRole('button', { name: /新建用户/ }).click()
    await expect(page.locator('h2', { hasText: '新建用户' })).toBeVisible()
    // 验证表单字段存在
    await expect(page.locator('.input').nth(0)).toBeVisible()
    await expect(page.locator('select')).toHaveCount(2)
  })

  test('新建用户 - 空字段校验', async ({ authedPage: page }) => {
    await page.goto('/system/users')
    await page.getByRole('button', { name: /新建用户/ }).click()
    await page.getByRole('button', { name: '保存' }).click()
    await expect(page.locator('text=用户名和姓名不能为空')).toBeVisible({ timeout: 3000 })
  })

  test('新建用户成功', async ({ authedPage: page }) => {
    await page.goto('/system/users')
    await page.getByRole('button', { name: /新建用户/ }).click()

    const username = `test_${Date.now()}`
    await page.locator('.input').nth(0).fill(username)
    await page.locator('.input').nth(1).fill('测试用户')
    await page.locator('input[type="password"]').fill('password123')

    await page.getByRole('button', { name: '保存' }).click()
    await expect(page.locator('text=创建成功')).toBeVisible({ timeout: 5000 })
    await expect(page.getByRole('cell', { name: username })).toBeVisible()
  })

  test('编辑用户', async ({ authedPage: page }) => {
    await page.goto('/system/users')
    // 点击 admin 行的编辑按钮
    const adminRow = page.locator('tr', { hasText: 'admin' }).first()
    await adminRow.locator('button').first().click()
    await expect(page.locator('h2', { hasText: '编辑用户（admin）' })).toBeVisible()
    // 用户名字段应禁用
    await expect(page.locator('.input').nth(0)).toBeDisabled()
  })

  test('不能删除管理员账户', async ({ authedPage: page }) => {
    await page.goto('/system/users')
    const adminRow = page.locator('tr', { hasText: 'admin' }).first()
    const deleteBtn = adminRow.locator('button').nth(1)
    await expect(deleteBtn).toBeDisabled()
  })

  test('取消对话框', async ({ authedPage: page }) => {
    await page.goto('/system/users')
    await page.getByRole('button', { name: /新建用户/ }).click()
    await expect(page.locator('h2', { hasText: '新建用户' })).toBeVisible()
    await page.getByRole('button', { name: '取消' }).click()
    await expect(page.locator('h2', { hasText: '新建用户' })).toBeHidden()
  })
})
