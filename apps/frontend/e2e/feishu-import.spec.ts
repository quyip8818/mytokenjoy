import { expect, test } from '@playwright/test'

test.describe('飞书数据导入', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/structure')
    await expect(page.getByRole('heading', { name: '组织架构' })).toBeVisible()
  })

  test('飞书导入成员正确映射到总公司', async ({ page }) => {
    // Trigger import via API
    const importResult = await page.evaluate(async () => {
      const res = await fetch('/api/org/data-source/import', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      })
      return { status: res.status, data: await res.json() }
    })

    expect(importResult.status).toBe(200)
    expect(importResult.data.successMembers).toBeGreaterThanOrEqual(7)
    expect(importResult.data.failures).toHaveLength(0)

    // Verify members appear in the org structure
    const members = await page.evaluate(async () => {
      const res = await fetch('/api/org/members?page=1&pageSize=100', { credentials: 'include' })
      return res.json()
    })

    const feishuMembers = members.items.filter(
      (m: { id: string }) => m.id.startsWith('m-feishu-'),
    )
    expect(feishuMembers.length).toBeGreaterThanOrEqual(7)

    // All feishu members should be in 总公司 (root dept)
    for (const m of feishuMembers) {
      expect(m.departmentName).toBe('总公司')
      expect(m.status).toBe('active')
      expect(m.source).toBe('imported')
      expect(m.name).toBeTruthy()
      expect(m.phone).toBeTruthy()
    }
  })

  test('飞书导入幂等：重复导入不产生重复成员', async ({ page }) => {
    // Import twice
    const import1 = await page.evaluate(async () => {
      const res = await fetch('/api/org/data-source/import', {
        method: 'POST',
        credentials: 'include',
      })
      return res.json()
    })

    const import2 = await page.evaluate(async () => {
      const res = await fetch('/api/org/data-source/import', {
        method: 'POST',
        credentials: 'include',
      })
      return res.json()
    })

    // Second import should not add new members (idempotent)
    expect(import2.successMembers).toBeLessThanOrEqual(import1.successMembers)

    // Verify no duplicate feishu members
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

  test('导入后部门树正确显示成员计数', async ({ page }) => {
    // Get the department tree
    const tree = await page.evaluate(async () => {
      const res = await fetch('/api/org/departments/tree', { credentials: 'include' })
      return res.json()
    })

    // Root dept (总公司) should include feishu members in count
    const root = tree[0]
    expect(root.name).toBe('总公司')
    expect(root.memberCount).toBeGreaterThanOrEqual(7)
  })

  test('导入的成员在组织架构页面可见', async ({ page }) => {
    // Select 总公司 and search for a feishu member name
    await page.getByRole('treeitem', { name: /总公司/ }).click()
    await page.waitForTimeout(500)

    // Search for a known feishu member
    const searchInput = page.locator('input[placeholder*="搜索成员"]')
    await searchInput.fill('闰土')
    await page.waitForTimeout(1000)

    // Should find the member
    await expect(page.getByRole('cell', { name: '闰土' })).toBeVisible()
  })

  test('数据源状态显示已导入', async ({ page }) => {
    const status = await page.evaluate(async () => {
      const res = await fetch('/api/org/data-source/status', { credentials: 'include' })
      return res.json()
    })

    expect(status.platform).toBe('feishu')
    expect(status.connected).toBe(true)
    expect(status.lastImport).toBeTruthy()
    expect(status.lastImportResult).toBeTruthy()
    expect(status.lastImportResult.successMembers).toBeGreaterThanOrEqual(7)
  })
})
