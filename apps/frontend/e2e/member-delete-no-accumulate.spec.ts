import { expect, test } from '@playwright/test'

test.describe('成员删除 - 请求不累积', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/structure')
    await expect(page.getByRole('heading', { name: '组织架构' })).toBeVisible()
    // Select a department with multiple members
    await page.getByRole('treeitem', { name: /总公司/ }).click()
    await expect(page.getByRole('heading', { level: 3, name: '总公司' })).toBeVisible()
  })

  test('deleting members sequentially does not accumulate extra API requests', async ({
    page,
  }) => {
    // Helper to delete the first active member and count resulting API calls
    async function deleteFirstMemberAndCountRequests(): Promise<{
      sessionCalls: number
      memberCalls: number
      treeCalls: number
    }> {
      // Wait for table to stabilize
      await page.waitForTimeout(500)

      const activeRows = page.getByRole('row').filter({ hasText: '已激活' })
      const rowCount = await activeRows.count()
      if (rowCount === 0) {
        throw new Error('No active members to delete')
      }

      // Start counting API requests BEFORE triggering delete
      let sessionCalls = 0
      let memberCalls = 0
      let treeCalls = 0

      const requestHandler = (request: { url: () => string; method: () => string }) => {
        const url = request.url()
        if (url.includes('/api/session') && request.method() === 'GET') sessionCalls++
        if (url.includes('/api/org/members') && request.method() === 'GET') memberCalls++
        if (url.includes('/api/org/departments/tree')) treeCalls++
      }
      page.on('request', requestHandler)

      // Select first active member
      await activeRows.first().getByRole('checkbox').click()

      // Click delete button
      await page.getByRole('button', { name: /删除/ }).click()

      // Confirm
      await expect(page.getByRole('alertdialog')).toBeVisible()
      await page.getByRole('button', { name: '确认' }).click()

      // Wait for dialog to close and data to reload
      await expect(page.getByRole('alertdialog')).toBeHidden()
      await page.waitForTimeout(2000)

      page.off('request', requestHandler)
      return { sessionCalls, memberCalls, treeCalls }
    }

    // Delete first member
    const first = await deleteFirstMemberAndCountRequests()

    // Delete second member
    const second = await deleteFirstMemberAndCountRequests()

    // Delete third member
    const third = await deleteFirstMemberAndCountRequests()

    // The key assertion: request counts should NOT accumulate.
    // Each delete should produce roughly the same number of requests (±1 tolerance).
    // If accumulating, second would be ~2x first, third ~3x first.
    const maxAllowed = Math.max(first.sessionCalls, 1) + 1

    expect(second.sessionCalls).toBeLessThanOrEqual(maxAllowed)
    expect(third.sessionCalls).toBeLessThanOrEqual(maxAllowed)

    expect(second.memberCalls).toBeLessThanOrEqual(Math.max(first.memberCalls, 1) + 1)
    expect(third.memberCalls).toBeLessThanOrEqual(Math.max(first.memberCalls, 1) + 1)

    expect(second.treeCalls).toBeLessThanOrEqual(Math.max(first.treeCalls, 1) + 1)
    expect(third.treeCalls).toBeLessThanOrEqual(Math.max(first.treeCalls, 1) + 1)
  })
})
