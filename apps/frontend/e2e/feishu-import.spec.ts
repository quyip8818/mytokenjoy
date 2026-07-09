import { expect, test } from '@playwright/test'

test.describe('飞书数据导入', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/structure')
    await expect(page.getByRole('heading', { name: '组织架构' })).toBeVisible()
  })

  test('飞书导入成员和部门结构', async ({ page }) => {
    // Trigger import
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

    // Flatten tree
    const allDepts: { name: string; id: string; memberCount: number }[] = []
    function walk(nodes: typeof tree) {
      for (const n of nodes) {
        allDepts.push({ name: n.name, id: n.id, memberCount: n.memberCount })
        walk(n.children || [])
      }
    }
    walk(tree)

    // Feishu departments should exist
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

    const feishuMembers = members.items.filter(
      (m: { id: string }) => m.id.startsWith('m-feishu-'),
    )
    expect(feishuMembers.length).toBeGreaterThanOrEqual(9)

    // 杨雨涵 should be in 软件研发
    const yangYuhan = feishuMembers.find((m: { name: string }) => m.name === '杨雨涵')
    expect(yangYuhan).toBeTruthy()
    expect(yangYuhan.departmentName).toBe('软件研发')

    // 张淑峰 should be in 市场部
    const zhangShufeng = feishuMembers.find((m: { name: string }) => m.name === '张淑峰')
    expect(zhangShufeng).toBeTruthy()
    expect(zhangShufeng.departmentName).toBe('市场部')

    // Others in 总公司
    const inRoot = feishuMembers.filter(
      (m: { departmentName: string }) => m.departmentName === '总公司',
    )
    expect(inRoot.length).toBeGreaterThanOrEqual(7)
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
    await page.waitForTimeout(1000)
    await expect(page.getByRole('cell', { name: '杨雨涵' })).toBeVisible()
  })
})
