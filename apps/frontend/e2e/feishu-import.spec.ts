import { expect, test } from '@playwright/test'
import { E2E_BASE_URL } from './e2e-db'

// Feishu import tests require data source to be connected.
// In a fresh demo without feishu credentials, these tests will be skipped.
test.describe('飞书数据导入', () => {
  let isConnected = false

  test.beforeAll(async ({ browser }) => {
    const context = await browser.newContext({
      storageState: '.auth/admin.json',
      baseURL: E2E_BASE_URL,
    })
    const page = await context.newPage()
    await page.goto(E2E_BASE_URL + '/')
    const status = await page.evaluate(async () => {
      const res = await fetch('/api/org/data-source/status', { credentials: 'include' })
      return res.json()
    })
    isConnected = status.connected === true
    await context.close()
  })

  test.beforeEach(async ({ page }) => {
    test.skip(!isConnected, '数据源未连接，跳过飞书导入测试')
    await page.goto('/org/structure')
    await expect(page.getByRole('banner').getByRole('heading', { name: '组织架构' })).toBeVisible()
  })

  test('飞书导入成员和部门结构', async ({ page }) => {
    const importResult = await page.evaluate(async () => {
      const res = await fetch('/api/org/data-source/import', {
        method: 'POST',
        credentials: 'include',
      })
      return { status: res.status, data: await res.json() }
    })

    expect(importResult.status).toBe(200)
    expect(importResult.data.successMembers).toBeGreaterThanOrEqual(9)
    expect(importResult.data.successDepartments).toBeGreaterThanOrEqual(2)
    expect(importResult.data.failures).toHaveLength(0)
  })

  test('导入的部门正确出现在部门树中', async ({ page }) => {
    const tree = await page.evaluate(async () => {
      const res = await fetch('/api/org/departments/tree', { credentials: 'include' })
      return res.json()
    })

    const allDepts: { name: string; id: string; memberCount: number }[] = []
    function walk(nodes: typeof tree) {
      for (const n of nodes) {
        allDepts.push({ name: n.name, id: n.id, memberCount: n.memberCount })
        walk(n.children || [])
      }
    }
    walk(tree)

    const feishuDepts = allDepts.filter((d) => d.id.includes('feishu'))
    expect(feishuDepts.length).toBeGreaterThanOrEqual(2)

    const deptNames = feishuDepts.map((d) => d.name)
    expect(deptNames).toContain('软件研发')
    expect(deptNames).toContain('市场部')
  })

  test('成员正确归属到对应部门', async ({ page }) => {
    const members = await page.evaluate(async () => {
      const res = await fetch('/api/org/members?page=1&pageSize=100', { credentials: 'include' })
      return res.json()
    })

    const feishuMembers = members.items.filter((m: { id: string }) => m.id.startsWith('m-feishu-'))
    expect(feishuMembers.length).toBeGreaterThanOrEqual(9)

    const yangYuhan = feishuMembers.find((m: { alias: string }) => m.alias === '杨雨涵')
    expect(yangYuhan).toBeTruthy()
    expect(yangYuhan.departmentName).toBe('软件研发')

    const zhangShufeng = feishuMembers.find((m: { alias: string }) => m.alias === '张淑峰')
    expect(zhangShufeng).toBeTruthy()
    expect(zhangShufeng.departmentName).toBe('市场部')
  })

  test('导入幂等：重复导入不产生重复', async ({ page }) => {
    await page.evaluate(async () => {
      await fetch('/api/org/data-source/import', { method: 'POST', credentials: 'include' })
    })
    await page.evaluate(async () => {
      await fetch('/api/org/data-source/import', { method: 'POST', credentials: 'include' })
    })

    const members = await page.evaluate(async () => {
      const res = await fetch('/api/org/members?page=1&pageSize=100', { credentials: 'include' })
      return res.json()
    })
    const feishuIds = members.items
      .filter((m: { id: string }) => m.id.startsWith('m-feishu-'))
      .map((m: { id: string }) => m.id)
    const uniqueIds = new Set(feishuIds)
    expect(feishuIds.length).toBe(uniqueIds.size)
  })

  test('数据源状态反映导入结果', async ({ page }) => {
    const status = await page.evaluate(async () => {
      const res = await fetch('/api/org/data-source/status', { credentials: 'include' })
      return res.json()
    })

    expect(status.platform).toBe('feishu')
    expect(status.connected).toBe(true)
    expect(status.lastImport).toBeTruthy()
    expect(status.lastImportResult.successMembers).toBeGreaterThanOrEqual(9)
    expect(status.lastImportResult.successDepartments).toBeGreaterThanOrEqual(2)
  })

  test('导入的成员在页面可搜索', async ({ page }) => {
    await page.getByRole('treeitem', { name: /总公司/ }).click()
    await page.waitForTimeout(500)

    const searchInput = page.locator('input[placeholder*="搜索成员"]')
    await searchInput.fill('杨雨涵')
    await searchInput.press('Enter')
    await page.waitForTimeout(1000)
    await expect(page.getByRole('cell', { name: '杨雨涵' })).toBeVisible()
  })
})
