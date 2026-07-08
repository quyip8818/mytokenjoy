import { expect, test, type Page, type Response } from '@playwright/test'

/**
 * 组织架构页面 E2E 测试
 *
 * 覆盖范围：
 * - 页面渲染：部门树、成员列表、分页
 * - 部门操作：选择部门过滤、展开/收起子部门
 * - 成员搜索与分页
 * - 成员 CRUD：添加、编辑、删除
 * - 批量操作：停用、启用、转移部门
 * - API 数据校验：接口返回结构和字段完整性
 */

test.describe('组织架构 - 页面渲染与数据', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/structure')
    await expect(page.getByRole('heading', { name: '组织架构' })).toBeVisible()
  })

  test('部门树正确渲染根节点和子部门', async ({ page }) => {
    await expect(page.getByRole('treeitem', { name: /全部成员/ })).toBeVisible()
    await expect(page.getByRole('treeitem', { name: /总公司/ })).toBeVisible()
    // 子部门可见（总公司下的直接子部门）
    await expect(page.getByRole('treeitem', { name: /技术部/ })).toBeVisible()
    await expect(page.getByRole('treeitem', { name: /产品部/ })).toBeVisible()
    await expect(page.getByRole('treeitem', { name: /市场部/ })).toBeVisible()
    await expect(page.getByRole('treeitem', { name: /行政部/ })).toBeVisible()
  })

  test('成员列表表头和数据列完整', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: '姓名' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '部门' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '手机号' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '状态' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '操作' })).toBeVisible()
    // 至少有数据行
    const dataRows = page.getByRole('row').filter({ hasNot: page.getByRole('columnheader') })
    await expect(dataRows.first()).toBeVisible()
  })

  test('成员总数显示且大于 0', async ({ page }) => {
    const countText = page.getByText(/共 \d+ 人/)
    await expect(countText).toBeVisible()
    const text = await countText.textContent()
    const count = parseInt(text?.match(/\d+/)?.[0] ?? '0')
    expect(count).toBeGreaterThan(0)
  })

  test('分页控件显示且可操作', async ({ page }) => {
    const pageInfo = page.getByText(/\d+ \/ \d+/)
    await expect(pageInfo).toBeVisible()

    const text = await pageInfo.textContent()
    expect(text).toMatch(/1 \/ \d+/)

    // 下一页按钮可点击
    const nextBtn = page.getByRole('button', { name: '下一页' })
    await expect(nextBtn).toBeEnabled()
    await nextBtn.click()
    await expect(pageInfo).toHaveText(/2 \/ \d+/)
  })
})

test.describe('组织架构 - 部门选择与过滤', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/structure')
    await expect(page.getByRole('heading', { name: '组织架构' })).toBeVisible()
  })

  test('点击部门过滤成员列表', async ({ page }) => {
    // 点击技术部
    await page.getByRole('treeitem', { name: /技术部/ }).click()
    await expect(page.getByRole('heading', { level: 3, name: '技术部' })).toBeVisible()

    // 成员数应比全部少
    const countText = await page.getByText(/共 \d+ 人/).textContent()
    const count = parseInt(countText?.match(/\d+/)?.[0] ?? '0')
    expect(count).toBeGreaterThan(0)
    expect(count).toBeLessThan(50)
  })

  test('点击全部成员显示所有人', async ({ page }) => {
    // 先选择一个部门
    await page.getByRole('treeitem', { name: /技术部/ }).click()
    await page.waitForTimeout(300)

    // 切回全部成员
    await page.getByRole('treeitem', { name: /全部成员/ }).click()
    await expect(page.getByRole('heading', { level: 3, name: '全部成员' })).toBeVisible()
  })

  test('展开子部门显示更细分部门', async ({ page }) => {
    // 技术部下应有子部门（后端组、前端组、测试组）
    const techItem = page.getByRole('treeitem', { name: /技术部/ })
    await techItem.getByRole('button', { name: '展开' }).click()
    await expect(page.getByRole('treeitem', { name: /后端组/ })).toBeVisible()
    await expect(page.getByRole('treeitem', { name: /前端组/ })).toBeVisible()
    await expect(page.getByRole('treeitem', { name: /测试组/ })).toBeVisible()
  })
})

test.describe('组织架构 - 成员搜索', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/structure')
    await expect(page.getByRole('heading', { name: '组织架构' })).toBeVisible()
  })

  test('输入关键字过滤成员', async ({ page }) => {
    const allCountText = await page.getByText(/共 \d+ 人/).textContent()
    const allCount = parseInt(allCountText?.match(/\d+/)?.[0] ?? '0')

    const searchInput = page.locator('input[placeholder*="搜索成员"]')
    await searchInput.fill('伟')
    await page.waitForTimeout(800)

    const filteredText = await page.getByText(/共 \d+ 人/).textContent()
    const filteredCount = parseInt(filteredText?.match(/\d+/)?.[0] ?? '0')
    expect(filteredCount).toBeGreaterThan(0)
    expect(filteredCount).toBeLessThan(allCount)
  })

  test('清空搜索恢复全部成员', async ({ page }) => {
    const searchInput = page.locator('input[placeholder*="搜索成员"]')
    await searchInput.fill('伟')
    await page.waitForTimeout(800)

    await searchInput.clear()
    await page.waitForTimeout(800)

    const countText = await page.getByText(/共 \d+ 人/).textContent()
    const count = parseInt(countText?.match(/\d+/)?.[0] ?? '0')
    expect(count).toBeGreaterThan(5) // 恢复到全量
  })
})

test.describe('组织架构 - 成员 CRUD', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/structure')
    await expect(page.getByRole('heading', { name: '组织架构' })).toBeVisible()
    await page.getByRole('treeitem', { name: /总公司/ }).click()
    await expect(page.getByRole('heading', { level: 3, name: '总公司' })).toBeVisible()
  })

  test('添加成员：表单提交后列表更新', async ({ page }) => {
    const countText = await page.getByText(/共 \d+ 人/).textContent()
    const countBefore = parseInt(countText?.match(/\d+/)?.[0] ?? '0')

    await page.getByRole('button', { name: '添加成员' }).click()
    await expect(page.getByRole('dialog', { name: '添加成员' })).toBeVisible()

    const uniqueName = `自动化${Date.now().toString().slice(-6)}`
    await page.locator('input[name="name"]').fill(uniqueName)
    await page.locator('input[name="phone"]').fill('13700001111')
    await page.locator('input[name="email"]').fill(`auto-${Date.now()}@test.com`)
    await page.getByRole('combobox').click()
    await page.getByRole('option', { name: '总公司' }).click()
    await page.getByRole('button', { name: '添加' }).click()

    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10_000 })
    await expect(page.getByRole('cell', { name: uniqueName })).toBeVisible()
    await expect(page.getByText(`共 ${countBefore + 1} 人`)).toBeVisible()
  })

  test('编辑成员：修改姓名后列表更新', async ({ page }) => {
    const activeRow = page.getByRole('row').filter({ hasText: '已激活' }).first()
    const originalName = await activeRow.getByRole('cell').nth(1).textContent()

    await activeRow.getByRole('button', { name: '编辑' }).click()
    await expect(page.getByRole('dialog', { name: '编辑成员' })).toBeVisible()

    const newName = `改名${Date.now().toString().slice(-4)}`
    const nameInput = page.locator('input[name="name"]')
    await nameInput.clear()
    await nameInput.fill(newName)
    await page.getByRole('button', { name: '保存' }).click()

    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10_000 })
    await expect(page.getByRole('cell', { name: newName })).toBeVisible()
  })

  test('删除成员：确认后从列表消失', async ({ page }) => {
    const countText = await page.getByText(/共 \d+ 人/).textContent()
    const countBefore = parseInt(countText?.match(/\d+/)?.[0] ?? '0')

    const activeRow = page.getByRole('row').filter({ hasText: '已激活' }).first()
    await activeRow.getByRole('checkbox').click()
    await page.getByRole('button', { name: /删除/ }).click()

    await expect(page.getByRole('alertdialog', { name: '删除成员' })).toBeVisible()
    await expect(page.getByText('删除后不可恢复')).toBeVisible()
    await page.getByRole('button', { name: '确认' }).click()

    await expect(page.getByRole('alertdialog')).toBeHidden()
    await expect(page.getByText(`共 ${countBefore - 1} 人`)).toBeVisible({ timeout: 10_000 })
  })
})

test.describe('组织架构 - 批量操作', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/structure')
    await expect(page.getByRole('heading', { name: '组织架构' })).toBeVisible()
  })

  test('选中成员后显示批量操作工具栏', async ({ page }) => {
    const activeRow = page.getByRole('row').filter({ hasText: '已激活' }).first()
    await activeRow.getByRole('checkbox').click()

    await expect(page.getByText('已选 1 人')).toBeVisible()
    await expect(page.getByRole('button', { name: '转移部门' })).toBeVisible()
    await expect(page.getByRole('button', { name: '停用' })).toBeVisible()
    await expect(page.getByRole('button', { name: '删除' })).toBeVisible()
  })

  test('取消选择隐藏工具栏', async ({ page }) => {
    const activeRow = page.getByRole('row').filter({ hasText: '已激活' }).first()
    await activeRow.getByRole('checkbox').click()
    await expect(page.getByText('已选 1 人')).toBeVisible()

    await page.getByRole('button', { name: '取消选择' }).click()
    await expect(page.getByText('已选 1 人')).toBeHidden()
  })

  test('停用成员后状态变更', async ({ page }) => {
    const activeRow = page.getByRole('row').filter({ hasText: '已激活' }).first()
    const memberName = await activeRow.getByRole('cell').nth(1).textContent()
    await activeRow.getByRole('checkbox').click()

    await page.getByRole('button', { name: '停用' }).click()
    await expect(page.getByRole('alertdialog')).toBeVisible()
    await page.getByRole('button', { name: '确认' }).click()
    await expect(page.getByRole('alertdialog')).toBeHidden()

    // 成员状态应变为 已停用
    await page.waitForTimeout(1000)
    if (memberName) {
      const row = page.getByRole('row').filter({ hasText: memberName })
      await expect(row.getByText(/已停用|停用/)).toBeVisible({ timeout: 5_000 })
    }
  })
})

test.describe('组织架构 - API 数据校验', () => {
  test('departments/tree 接口返回正确结构', async ({ page }) => {
    const responsePromise = page.waitForResponse((r) =>
      r.url().includes('/api/org/departments/tree') && r.status() === 200,
    )
    await page.goto('/org/structure')
    const response = await responsePromise
    const data = await response.json()

    // 应为数组，根节点是总公司
    expect(Array.isArray(data)).toBe(true)
    expect(data.length).toBeGreaterThan(0)

    const root = data[0]
    expect(root).toHaveProperty('id')
    expect(root).toHaveProperty('name')
    expect(root).toHaveProperty('children')
    expect(root).toHaveProperty('memberCount')
    expect(root.name).toBe('总公司')
    expect(root.memberCount).toBeGreaterThan(0)
    expect(Array.isArray(root.children)).toBe(true)
    expect(root.children.length).toBeGreaterThan(0)

    // 子部门结构
    const child = root.children[0]
    expect(child).toHaveProperty('id')
    expect(child).toHaveProperty('name')
    expect(child).toHaveProperty('memberCount')
  })

  test('members 接口返回分页结构和完整字段', async ({ page }) => {
    const responsePromise = page.waitForResponse((r) =>
      r.url().includes('/api/org/members') && r.status() === 200,
    )
    await page.goto('/org/structure')
    const response = await responsePromise
    const data = await response.json()

    // 分页结构
    expect(data).toHaveProperty('items')
    expect(data).toHaveProperty('total')
    expect(data).toHaveProperty('page')
    expect(data).toHaveProperty('pageSize')
    expect(data.total).toBeGreaterThan(0)
    expect(data.page).toBe(1)
    expect(data.pageSize).toBe(10)
    expect(Array.isArray(data.items)).toBe(true)
    expect(data.items.length).toBeLessThanOrEqual(10)

    // 成员字段完整性
    const member = data.items[0]
    expect(member).toHaveProperty('id')
    expect(member).toHaveProperty('name')
    expect(member).toHaveProperty('departmentId')
    expect(member).toHaveProperty('departmentName')
    expect(member).toHaveProperty('status')
    expect(member).toHaveProperty('roles')
    expect(member).toHaveProperty('source')
    expect(typeof member.id).toBe('string')
    expect(typeof member.name).toBe('string')
    expect(Array.isArray(member.roles)).toBe(true)
    expect(['active', 'inactive', 'pending']).toContain(member.status)
  })

  test('删除成员 API 返回 200 且成员从列表消失', async ({ page }) => {
    await page.goto('/org/structure')
    await expect(page.getByRole('heading', { name: '组织架构' })).toBeVisible()
    await page.waitForTimeout(500)

    // 获取当前成员列表
    const membersResponse = await page.evaluate(async () => {
      const res = await fetch('/api/org/members?page=1&pageSize=10', { credentials: 'include' })
      return { status: res.status, data: await res.json() }
    })
    expect(membersResponse.status).toBe(200)
    const targetMember = membersResponse.data.items.find(
      (m: { status: string }) => m.status === 'active',
    )
    if (!targetMember) return

    // 执行删除
    const deleteResponse = await page.evaluate(async (id) => {
      const res = await fetch('/api/org/members', {
        method: 'DELETE',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ids: [id] }),
      })
      return { status: res.status }
    }, targetMember.id)
    expect(deleteResponse.status).toBe(200)

    // 验证成员不再在列表中
    const afterResponse = await page.evaluate(async () => {
      const res = await fetch('/api/org/members?page=1&pageSize=100', { credentials: 'include' })
      return res.json()
    })
    const ids = afterResponse.items.map((m: { id: string }) => m.id)
    expect(ids).not.toContain(targetMember.id)
  })

  test('编辑成员 API 保留 roles 和 status（merge 语义）', async ({ page }) => {
    await page.goto('/org/structure')
    await page.waitForTimeout(500)

    // 获取一个 active 成员
    const members = await page.evaluate(async () => {
      const res = await fetch('/api/org/members?page=1&pageSize=50', { credentials: 'include' })
      return res.json()
    })
    const target = members.items.find((m: { status: string }) => m.status === 'active')
    if (!target) return

    const originalRoles = target.roles
    const originalStatus = target.status

    // 只修改 name
    const updateResponse = await page.evaluate(
      async ({ id, name }) => {
        const res = await fetch(`/api/org/members/${id}`, {
          method: 'PUT',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name: name + '改' }),
        })
        return { status: res.status, data: await res.json() }
      },
      { id: target.id, name: target.name },
    )

    expect(updateResponse.status).toBe(200)
    // 验证 merge：roles 和 status 被保留
    expect(updateResponse.data.roles).toEqual(originalRoles)
    expect(updateResponse.data.status).toBe(originalStatus)
    expect(updateResponse.data.name).toBe(target.name + '改')
  })
})
