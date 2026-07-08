import { expect, test } from '@playwright/test'

/**
 * 角色管理页面 E2E 测试
 *
 * 覆盖范围：
 * - 页面渲染：角色列表（预设/自定义分组）、成员列表
 * - 角色选择：切换角色查看成员
 * - 角色 CRUD：创建、编辑（自定义角色）、删除
 * - 预设角色保护：不可编辑/删除预设角色
 * - 成员管理：添加成员到角色、移除成员
 * - API 数据校验：接口返回结构和字段完整性
 */

test.describe('角色管理 - 页面渲染', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
    await expect(page.getByRole('heading', { name: '角色管理' })).toBeVisible()
  })

  test('显示角色列表标题和新建按钮', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 3, name: '角色' })).toBeVisible()
    await expect(page.getByRole('button', { name: '新建' })).toBeVisible()
  })

  test('显示系统预设角色分组', async ({ page }) => {
    await expect(page.getByText('系统预设')).toBeVisible()
    await expect(page.getByText('超级管理员').first()).toBeVisible()
    await expect(page.getByText('组织管理员').first()).toBeVisible()
    await expect(page.getByText('普通成员').first()).toBeVisible()
  })

  test('显示自定义角色分组', async ({ page }) => {
    await expect(page.getByText('自定义')).toBeVisible()
  })

  test('角色项显示成员数量', async ({ page }) => {
    // 超级管理员显示成员数
    const adminItem = page.locator('text=超级管理员').first().locator('..')
    await expect(adminItem).toContainText(/\d+/)
  })

  test('默认选中第一个角色并显示成员表', async ({ page }) => {
    // 右侧面板标题
    await expect(page.getByRole('heading', { level: 3 }).nth(1)).toBeVisible()
    // 成员表表头
    await expect(page.getByRole('columnheader', { name: '姓名' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '角色' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '操作' })).toBeVisible()
  })
})

test.describe('角色管理 - 角色切换', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
    await expect(page.getByRole('heading', { name: '角色管理' })).toBeVisible()
  })

  test('点击不同角色切换右侧成员列表', async ({ page }) => {
    // 点击普通成员
    await page.locator('text=普通成员').first().click()
    await page.waitForTimeout(500)
    await expect(page.getByRole('heading', { level: 3, name: '普通成员' })).toBeVisible()
    await expect(page.getByText(/名成员/)).toBeVisible()

    // 切换到超级管理员
    await page.locator('text=超级管理员').first().click()
    await page.waitForTimeout(500)
    await expect(page.getByRole('heading', { level: 3, name: '超级管理员' })).toBeVisible()
  })

  test('预设角色显示"系统预设角色"标识', async ({ page }) => {
    await page.locator('text=超级管理员').first().click()
    await page.waitForTimeout(300)
    await expect(page.getByText('系统预设角色')).toBeVisible()
  })

  test('搜索角色可过滤列表', async ({ page }) => {
    const searchInput = page.locator('input[placeholder*="搜索角色"]')
    await searchInput.fill('管理')
    await page.waitForTimeout(500)
    // 应该能看到包含"管理"的角色
    await expect(page.getByText('超级管理员').first()).toBeVisible()
    await expect(page.getByText('组织管理员').first()).toBeVisible()
  })
})

test.describe('角色管理 - 角色 CRUD', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
    await expect(page.getByRole('heading', { name: '角色管理' })).toBeVisible()
  })

  test('创建自定义角色', async ({ page }) => {
    await page.getByRole('button', { name: '新建' }).click()
    await expect(page.getByRole('dialog', { name: '创建角色' })).toBeVisible()

    // 填写角色名
    const uniqueName = `测试角色${Date.now().toString().slice(-5)}`
    await page.locator('input[name="name"]').fill(uniqueName)

    // 选择至少一个权限
    const firstCheckbox = page.getByRole('dialog').getByRole('checkbox').first()
    await firstCheckbox.click()

    // 提交
    await page.getByRole('button', { name: '确定' }).click()
    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10_000 })

    // 新角色出现在自定义分组中
    await expect(page.getByText(uniqueName)).toBeVisible()
  })

  test('创建角色 - 空名称被拒绝', async ({ page }) => {
    await page.getByRole('button', { name: '新建' }).click()
    await expect(page.getByRole('dialog', { name: '创建角色' })).toBeVisible()

    // 不填名称，选一个权限
    const firstCheckbox = page.getByRole('dialog').getByRole('checkbox').first()
    await firstCheckbox.click()

    // 提交 - 应该有验证提示或 dialog 不关闭
    await page.getByRole('button', { name: '确定' }).click()
    await page.waitForTimeout(500)
    // Dialog 应该仍然打开（因为验证失败）
    await expect(page.getByRole('dialog')).toBeVisible()
  })

  test('创建角色 - 重复名称被拒绝', async ({ page }) => {
    await page.getByRole('button', { name: '新建' }).click()
    await expect(page.getByRole('dialog', { name: '创建角色' })).toBeVisible()

    // 使用已存在的角色名
    await page.locator('input[name="name"]').fill('超级管理员')
    const firstCheckbox = page.getByRole('dialog').getByRole('checkbox').first()
    await firstCheckbox.click()

    await page.getByRole('button', { name: '确定' }).click()
    await page.waitForTimeout(1000)
    // Should show error or dialog stays open
    const dialogStillOpen = await page.getByRole('dialog').isVisible()
    expect(dialogStillOpen).toBe(true)
  })

  test('预设角色不可修改', async ({ page }) => {
    // 验证通过 API 直接尝试修改预设角色返回 400
    const result = await page.evaluate(async () => {
      const res = await fetch('/api/org/roles/role-1', {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: 'hacked', permissions: ['p-1'] }),
      })
      return { status: res.status, data: await res.json() }
    })
    expect(result.status).toBe(400)
    expect(result.data.message).toContain('preset')
  })

  test('预设角色不可删除', async ({ page }) => {
    const result = await page.evaluate(async () => {
      const res = await fetch('/api/org/roles/role-1', {
        method: 'DELETE',
        credentials: 'include',
      })
      return { status: res.status, data: await res.json() }
    })
    expect(result.status).toBe(400)
    expect(result.data.message).toContain('preset')
  })
})

test.describe('角色管理 - 成员管理', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
    await expect(page.getByRole('heading', { name: '角色管理' })).toBeVisible()
  })

  test('成员表显示角色下的成员', async ({ page }) => {
    await page.locator('text=超级管理员').first().click()
    await page.waitForTimeout(500)
    // 至少有一个成员行
    const memberRow = page.getByRole('row').filter({ hasText: '管理员' })
    await expect(memberRow).toBeVisible()
  })

  test('添加成员对话框可打开并搜索', async ({ page }) => {
    // 选择一个普通成员角色
    await page.locator('text=普通成员').first().click()
    await page.waitForTimeout(500)

    await page.getByRole('button', { name: '添加成员' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()
    await expect(page.getByText('添加角色成员')).toBeVisible()

    // 有搜索输入框
    const searchInput = page.getByRole('dialog').locator('input')
    await expect(searchInput).toBeVisible()

    // 关闭
    await page.getByRole('button', { name: /关闭|Close/ }).click()
    await expect(page.getByRole('dialog')).toBeHidden()
  })

  test('移除成员按钮可见', async ({ page }) => {
    await page.locator('text=超级管理员').first().click()
    await page.waitForTimeout(500)
    await expect(page.getByRole('button', { name: '移除' })).toBeVisible()
  })

  test('不可移除最后一个超级管理员', async ({ page }) => {
    // 通过 API 验证
    const members = await page.evaluate(async () => {
      const res = await fetch('/api/org/roles/role-1/members', { credentials: 'include' })
      return res.json()
    })

    if (members.length === 1) {
      const result = await page.evaluate(async (memberId) => {
        const res = await fetch(`/api/org/roles/role-1/members/${memberId}`, {
          method: 'DELETE',
          credentials: 'include',
        })
        return { status: res.status, data: await res.json() }
      }, members[0].id)
      expect(result.status).toBe(400)
      expect(result.data.message).toContain('last super admin')
    }
  })

  test('不可通过 API 添加成员到受保护预设角色', async ({ page }) => {
    // 尝试给超级管理员角色添加成员应该被拒绝
    const result = await page.evaluate(async () => {
      // 获取一个普通成员ID
      const membersRes = await fetch('/api/org/members?page=1&pageSize=5', {
        credentials: 'include',
      })
      const members = await membersRes.json()
      const target = members.items.find((m: { id: string }) => m.id !== 'm-admin')
      if (!target) return { status: 0, skipped: true }

      const res = await fetch(`/api/org/roles/role-1/members`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ memberId: target.id }),
      })
      return { status: res.status, data: await res.json() }
    })
    if (!('skipped' in result)) {
      expect(result.status).toBe(403)
      expect(result.data.message).toContain('protected')
    }
  })
})

test.describe('角色管理 - API 数据校验', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
    await page.waitForTimeout(500)
  })

  test('roles 接口返回完整角色列表', async ({ page }) => {
    const data = await page.evaluate(async () => {
      const res = await fetch('/api/org/roles', { credentials: 'include' })
      return res.json()
    })

    expect(Array.isArray(data)).toBe(true)
    expect(data.length).toBeGreaterThanOrEqual(5) // 至少有预设角色

    // 验证角色结构
    const role = data[0]
    expect(role).toHaveProperty('id')
    expect(role).toHaveProperty('name')
    expect(role).toHaveProperty('type')
    expect(role).toHaveProperty('permissions')
    expect(role).toHaveProperty('memberCount')
    expect(typeof role.id).toBe('string')
    expect(typeof role.name).toBe('string')
    expect(['preset', 'custom']).toContain(role.type)
    expect(Array.isArray(role.permissions)).toBe(true)
    expect(typeof role.memberCount).toBe('number')

    // 验证预设角色存在
    const names = data.map((r: { name: string }) => r.name)
    expect(names).toContain('超级管理员')
    expect(names).toContain('组织管理员')
    expect(names).toContain('普通成员')
  })

  test('permissions 接口返回权限定义列表', async ({ page }) => {
    const data = await page.evaluate(async () => {
      const res = await fetch('/api/org/permissions', { credentials: 'include' })
      return res.json()
    })

    expect(Array.isArray(data)).toBe(true)
    expect(data.length).toBeGreaterThan(0)

    // 验证权限结构
    const perm = data[0]
    expect(perm).toHaveProperty('id')
    expect(perm).toHaveProperty('name')
    expect(perm).toHaveProperty('group')
    expect(typeof perm.id).toBe('string')
    expect(typeof perm.name).toBe('string')
    expect(typeof perm.group).toBe('string')
  })

  test('role members 接口返回该角色下的成员列表', async ({ page }) => {
    const data = await page.evaluate(async () => {
      const res = await fetch('/api/org/roles/role-1/members', { credentials: 'include' })
      return { status: res.status, data: await res.json() }
    })

    expect(data.status).toBe(200)
    expect(Array.isArray(data.data)).toBe(true)

    if (data.data.length > 0) {
      const member = data.data[0]
      expect(member).toHaveProperty('id')
      expect(member).toHaveProperty('name')
      expect(member).toHaveProperty('roles')
      expect(member).toHaveProperty('status')
      expect(member).toHaveProperty('departmentName')
      expect(Array.isArray(member.roles)).toBe(true)
      expect(member.roles).toContain('超级管理员')
    }
  })

  test('创建并删除自定义角色 - 完整生命周期', async ({ page }) => {
    const roleName = `lifecycle-${Date.now()}`

    // 创建
    const createRes = await page.evaluate(async (name) => {
      const res = await fetch('/api/org/roles', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, permissions: ['p-1', 'p-2'] }),
      })
      return { status: res.status, data: await res.json() }
    }, roleName)

    expect(createRes.status).toBe(200)
    expect(createRes.data.name).toBe(roleName)
    expect(createRes.data.type).toBe('custom')
    expect(createRes.data.permissions).toContain('p-1')
    expect(createRes.data.permissions).toContain('p-2')
    const roleId = createRes.data.id

    // 更新
    const updateRes = await page.evaluate(
      async ({ id, name }) => {
        const res = await fetch(`/api/org/roles/${id}`, {
          method: 'PUT',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name: name + '-updated', permissions: ['p-1', 'p-3'] }),
        })
        return { status: res.status, data: await res.json() }
      },
      { id: roleId, name: roleName },
    )

    expect(updateRes.status).toBe(200)
    expect(updateRes.data.name).toBe(roleName + '-updated')
    expect(updateRes.data.permissions).toContain('p-3')
    expect(updateRes.data.permissions).not.toContain('p-2')

    // 删除
    const deleteRes = await page.evaluate(async (id) => {
      const res = await fetch(`/api/org/roles/${id}`, {
        method: 'DELETE',
        credentials: 'include',
      })
      return { status: res.status }
    }, roleId)

    expect(deleteRes.status).toBe(200)

    // 验证不再出现在列表中
    const listRes = await page.evaluate(async () => {
      const res = await fetch('/api/org/roles', { credentials: 'include' })
      return res.json()
    })
    const ids = listRes.map((r: { id: string }) => r.id)
    expect(ids).not.toContain(roleId)
  })
})
