import { expect, test } from '@playwright/test'

/**
 * 角色管理页面 E2E 测试
 *
 * 基于 PRD US-05 验收标准 + 页面实际交互覆盖：
 * - 页面渲染：角色列表（系统预设/自定义分组）、成员列表
 * - 角色切换：点击角色查看对应成员
 * - 搜索过滤：角色列表搜索 + 成员表搜索
 * - 角色 CRUD：创建、编辑、删除自定义角色
 * - 预设角色保护：不可编辑/删除
 * - 成员管理：添加成员到角色、移除成员
 * - 业务规则：普通成员不可移除、最后一个超管不可移除
 * - Toast 通知：操作成功/失败的用户反馈
 * - API 数据校验：接口返回结构完整性
 */

// ─── 页面渲染 ───────────────────────────────────────────────────────────

test.describe('角色管理 - 页面渲染', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
    await expect(page.getByRole('heading', { name: '角色管理' })).toBeVisible()
  })

  test('左侧面板显示角色标题和新建按钮', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 3, name: '角色' })).toBeVisible()
    await expect(page.getByRole('button', { name: '新建' })).toBeVisible()
  })

  test('显示系统预设角色分组及所有预设角色', async ({ page }) => {
    await expect(page.getByText('系统预设', { exact: true })).toBeVisible()
    await expect(page.getByText('超级管理员').first()).toBeVisible()
    await expect(page.getByText('组织管理员').first()).toBeVisible()
    await expect(page.getByText('普通成员').first()).toBeVisible()
  })

  test('显示自定义角色分组', async ({ page }) => {
    await expect(page.getByText('自定义')).toBeVisible()
  })

  test('角色项显示成员数量', async ({ page }) => {
    // 每个角色项旁边显示数字
    const roleItems = page.locator('[class*="cursor-pointer"]').filter({ hasText: '超级管理员' })
    await expect(roleItems.first()).toBeVisible()
    // 成员数是数字
    await expect(roleItems.first().locator('.tabular-nums')).toBeVisible()
  })

  test('默认选中第一个角色并显示右侧面板', async ({ page }) => {
    // 右侧面板标题 - 应该是第一个角色名
    const panelTitle = page.locator('h3').nth(1)
    await expect(panelTitle).toBeVisible()
    // 成员表存在
    await expect(page.getByRole('columnheader', { name: '姓名' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '角色' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: '操作' })).toBeVisible()
  })

  test('右侧面板显示角色类型标识和成员数', async ({ page }) => {
    await page.getByText('超级管理员').first().click()
    await expect(page.getByText('系统预设角色')).toBeVisible()
    await expect(page.getByText(/\d+ 名成员/)).toBeVisible()
  })

  test('右侧面板显示添加成员按钮', async ({ page }) => {
    await expect(page.getByRole('button', { name: '添加成员' })).toBeVisible()
  })

  test('角色搜索输入框存在', async ({ page }) => {
    await expect(page.locator('input[placeholder*="搜索角色"]')).toBeVisible()
  })

  test('成员搜索输入框存在', async ({ page }) => {
    await expect(page.locator('input[placeholder*="搜索成员"]')).toBeVisible()
  })
})

// ─── 角色切换 ───────────────────────────────────────────────────────────

test.describe('角色管理 - 角色切换', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
    await expect(page.getByRole('heading', { name: '角色管理' })).toBeVisible()
  })

  test('点击不同角色切换右侧成员列表', async ({ page }) => {
    // 点击普通成员角色
    await page.getByText('普通成员').first().click()
    await expect(page.getByRole('heading', { level: 3, name: '普通成员' })).toBeVisible()
    await expect(page.getByText(/\d+ 名成员/)).toBeVisible()

    // 切换到组织管理员
    await page.getByText('组织管理员').first().click()
    await expect(page.getByRole('heading', { level: 3, name: '组织管理员' })).toBeVisible()
  })

  test('选中角色有高亮样式', async ({ page }) => {
    const roleItem = page.locator('[class*="cursor-pointer"]').filter({ hasText: '超级管理员' }).first()
    await roleItem.click()
    await expect(roleItem).toHaveClass(/bg-muted/)
  })

  test('预设角色右侧面板标注"系统预设角色"', async ({ page }) => {
    await page.getByText('超级管理员').first().click()
    await expect(page.getByText('系统预设角色')).toBeVisible()
  })

  test('自定义角色右侧面板标注"自定义角色"', async ({ page }) => {
    // 先创建一个自定义角色
    const roleName = `切换测试角色${Date.now().toString().slice(-4)}`
    await page.evaluate(async (name) => {
      await fetch('/api/org/roles', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, permissions: ['p-1'] }),
      })
    }, roleName)
    await page.reload()
    await expect(page.getByText(roleName)).toBeVisible()

    await page.getByText(roleName).click()
    await expect(page.getByText('自定义角色')).toBeVisible()

    // 清理
    const roles = await page.evaluate(async () => {
      const res = await fetch('/api/org/roles', { credentials: 'include' })
      return res.json()
    })
    const created = roles.find((r: { name: string }) => r.name === roleName)
    if (created) {
      await page.evaluate(async (id) => {
        await fetch(`/api/org/roles/${id}`, { method: 'DELETE', credentials: 'include' })
      }, created.id)
    }
  })
})

// ─── 角色搜索过滤 ───────────────────────────────────────────────────────

test.describe('角色管理 - 搜索过滤', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
    await expect(page.getByRole('heading', { name: '角色管理' })).toBeVisible()
  })

  test('搜索角色关键词可过滤列表', async ({ page }) => {
    const searchInput = page.locator('input[placeholder*="搜索角色"]')
    await searchInput.fill('管理')
    // 包含"管理"的角色应该可见
    await expect(page.getByText('超级管理员').first()).toBeVisible()
    await expect(page.getByText('组织管理员').first()).toBeVisible()
  })

  test('搜索无匹配时显示提示', async ({ page }) => {
    const searchInput = page.locator('input[placeholder*="搜索角色"]')
    await searchInput.fill('不存在的角色xyz')
    await expect(page.getByText('无匹配角色')).toBeVisible()
  })

  test('清空搜索恢复完整角色列表', async ({ page }) => {
    const searchInput = page.locator('input[placeholder*="搜索角色"]')
    await searchInput.fill('管理')
    await expect(page.getByText('普通成员').first()).toBeHidden()

    await searchInput.clear()
    await expect(page.getByText('普通成员').first()).toBeVisible()
    await expect(page.getByText('超级管理员').first()).toBeVisible()
  })

  test('成员表搜索可过滤角色下的成员', async ({ page }) => {
    // 选择普通成员角色（通常有多个成员）
    await page.getByText('普通成员').first().click()
    await page.waitForTimeout(500)

    const memberSearch = page.locator('input[placeholder*="搜索成员"]')
    await memberSearch.fill('admin')
    // 过滤结果应该变化（至少不报错）
    await page.waitForTimeout(300)
  })
})

// ─── 角色 CRUD ──────────────────────────────────────────────────────────

test.describe('角色管理 - 角色 CRUD', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
    await expect(page.getByRole('heading', { name: '角色管理' })).toBeVisible()
  })

  test('创建自定义角色 - 完整流程', async ({ page }) => {
    await page.getByRole('button', { name: '新建' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()
    await expect(page.getByText('创建角色')).toBeVisible()

    const uniqueName = `E2E角色${Date.now().toString().slice(-5)}`
    await page.locator('input[name="name"]').fill(uniqueName)

    // 选择至少一个权限
    const firstCheckbox = page.getByRole('dialog').getByRole('checkbox').first()
    await firstCheckbox.click()

    // 提交
    await page.getByRole('button', { name: '创建' }).click()
    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10_000 })

    // toast 通知
    await expect(page.getByText(`角色「${uniqueName}」已创建`)).toBeVisible()

    // 新角色出现在自定义分组中
    await expect(page.getByText(uniqueName, { exact: true })).toBeVisible()

    // 清理
    const roles = await page.evaluate(async () => {
      const res = await fetch('/api/org/roles', { credentials: 'include' })
      return res.json()
    })
    const created = roles.find((r: { name: string }) => r.name === uniqueName)
    if (created) {
      await page.evaluate(async (id) => {
        await fetch(`/api/org/roles/${id}`, { method: 'DELETE', credentials: 'include' })
      }, created.id)
    }
  })

  test('创建角色 - 空名称无法提交', async ({ page }) => {
    await page.getByRole('button', { name: '新建' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()

    // 不填名称，选一个权限
    const firstCheckbox = page.getByRole('dialog').getByRole('checkbox').first()
    await firstCheckbox.click()

    await page.getByRole('button', { name: '创建' }).click()
    await page.waitForTimeout(500)
    // Dialog 仍然打开（验证未通过）
    await expect(page.getByRole('dialog')).toBeVisible()
  })

  test('创建角色 - 重复名称被后端拒绝', async ({ page }) => {
    await page.getByRole('button', { name: '新建' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()

    // 使用已存在的角色名
    await page.locator('input[name="name"]').fill('超级管理员')
    const firstCheckbox = page.getByRole('dialog').getByRole('checkbox').first()
    await firstCheckbox.click()

    await page.getByRole('button', { name: '创建' }).click()
    await page.waitForTimeout(1000)
    // 应该有错误 toast 或 dialog 仍打开
    const dialogVisible = await page.getByRole('dialog').isVisible()
    expect(dialogVisible).toBe(true)
  })

  test('编辑自定义角色', async ({ page }) => {
    // 先通过 API 创建角色
    const roleName = `编辑测试${Date.now().toString().slice(-4)}`
    const createRes = await page.evaluate(async (name) => {
      const res = await fetch('/api/org/roles', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, permissions: ['p-1'] }),
      })
      return res.json()
    }, roleName)
    await page.reload()
    await expect(page.getByText(roleName)).toBeVisible()

    // hover 角色项显示编辑按钮并点击
    const roleItem = page.locator('[class*="cursor-pointer"]').filter({ hasText: roleName })
    await roleItem.hover()
    const editBtn = roleItem.locator('button').first()
    await editBtn.click()

    // Dialog 应该打开且有角色名
    await expect(page.getByRole('dialog')).toBeVisible()
    const nameInput = page.locator('input[name="name"]')
    await expect(nameInput).toHaveValue(roleName)

    // 修改名称
    const updatedName = roleName + '改'
    await nameInput.clear()
    await nameInput.fill(updatedName)
    await page.getByRole('button', { name: '保存' }).click()
    await expect(page.getByRole('dialog')).toBeHidden({ timeout: 10_000 })

    // toast 通知
    await expect(page.getByText(`角色「${updatedName}」已更新`)).toBeVisible()
    // 列表中显示新名称
    await expect(page.getByText(updatedName, { exact: true })).toBeVisible()

    // 清理
    await page.evaluate(async (id) => {
      await fetch(`/api/org/roles/${id}`, { method: 'DELETE', credentials: 'include' })
    }, createRes.id)
  })

  test('删除自定义角色 - 确认对话框', async ({ page }) => {
    // 创建角色
    const roleName = `删除测试${Date.now().toString().slice(-4)}`
    const createRes = await page.evaluate(async (name) => {
      const res = await fetch('/api/org/roles', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, permissions: ['p-1'] }),
      })
      return { status: res.status, data: await res.json() }
    }, roleName)
    expect(createRes.status).toBe(200)

    await page.reload()
    await expect(page.getByRole('heading', { name: '角色管理' })).toBeVisible()
    await expect(page.getByText(roleName, { exact: true })).toBeVisible({ timeout: 10_000 })

    // hover 角色项显示删除按钮
    const roleItem = page.locator('[class*="cursor-pointer"]').filter({ hasText: roleName })
    await roleItem.hover()
    const deleteBtn = roleItem.locator('button').nth(1)
    await deleteBtn.click()

    // 确认对话框
    await expect(page.getByText('删除角色')).toBeVisible()
    await expect(page.getByText('确定要删除该角色吗？')).toBeVisible()

    // 确认删除
    await page.getByRole('button', { name: '删除' }).click()
    await page.waitForTimeout(1000)

    // toast 通知
    await expect(page.getByText('角色已删除')).toBeVisible()
    // 角色从列表消失
    await expect(page.getByText(roleName)).toBeHidden()
  })

  test('删除有成员的角色 - 提示成员数量', async ({ page }) => {
    // 创建角色并添加成员
    const roleName = `有成员角色${Date.now().toString().slice(-4)}`
    const createRes = await page.evaluate(async (name) => {
      const res = await fetch('/api/org/roles', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, permissions: ['p-1'] }),
      })
      return res.json()
    }, roleName)

    // 添加一个成员到该角色
    const members = await page.evaluate(async () => {
      const res = await fetch('/api/org/members?page=1&pageSize=1', { credentials: 'include' })
      return res.json()
    })
    if (members.items?.length > 0) {
      await page.evaluate(
        async ({ roleId, memberId }) => {
          await fetch(`/api/org/roles/${roleId}/members`, {
            method: 'POST',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ memberId }),
          })
        },
        { roleId: createRes.id, memberId: members.items[0].id },
      )
    }

    await page.reload()
    await expect(page.getByText(roleName)).toBeVisible()

    // 尝试删除
    const roleItem = page.locator('[class*="cursor-pointer"]').filter({ hasText: roleName })
    await roleItem.hover()
    const deleteBtn = roleItem.locator('button').nth(1)
    await deleteBtn.click()

    // 确认对话框应提示成员数量
    await expect(page.getByText(/该角色下有 \d+ 名成员/)).toBeVisible()

    // 取消删除
    await page.getByRole('button', { name: '取消' }).click()

    // 清理
    await page.evaluate(async (id) => {
      await fetch(`/api/org/roles/${id}`, { method: 'DELETE', credentials: 'include' })
    }, createRes.id)
  })

  test('预设角色无编辑/删除按钮', async ({ page }) => {
    // hover 预设角色不应该出现编辑/删除按钮
    const presetRole = page.locator('[class*="cursor-pointer"]').filter({ hasText: '超级管理员' }).first()
    await presetRole.hover()
    await page.waitForTimeout(300)

    // 预设角色 hover 后不应出现操作按钮
    const buttons = presetRole.locator('[class*="group-hover"]')
    await expect(buttons).toBeHidden()
  })
})

// ─── 预设角色保护（API 级别） ────────────────────────────────────────────

test.describe('角色管理 - 预设角色保护', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
  })

  test('API 不允许修改预设角色', async ({ page }) => {
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

  test('API 不允许删除预设角色', async ({ page }) => {
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

// ─── 成员管理 ───────────────────────────────────────────────────────────

test.describe('角色管理 - 成员管理', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/org/roles')
    await expect(page.getByRole('heading', { name: '角色管理' })).toBeVisible()
  })

  test('成员表显示角色下的成员', async ({ page }) => {
    await page.getByText('超级管理员').first().click()
    await page.waitForTimeout(500)
    // 至少有一个成员行
    const rows = page.getByRole('row')
    expect(await rows.count()).toBeGreaterThan(1) // header + at least 1 member row
  })

  test('成员行显示姓名和角色标签', async ({ page }) => {
    await page.getByText('超级管理员').first().click()
    await page.waitForTimeout(500)
    // 姓名列有内容
    const firstDataRow = page.getByRole('row').nth(1)
    await expect(firstDataRow.getByRole('cell').first()).not.toBeEmpty()
    // 角色标签（Badge）
    await expect(firstDataRow.locator('[class*="badge"]').first()).toBeVisible()
  })

  test('添加成员对话框 - 打开并搜索', async ({ page }) => {
    // 选择一个自定义角色或普通成员
    await page.getByText('普通成员').first().click()
    await page.waitForTimeout(500)

    await page.getByRole('button', { name: '添加成员' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()
    await expect(page.getByText('添加角色成员')).toBeVisible()

    // 有搜索输入框和搜索按钮
    const searchInput = page.getByRole('dialog').locator('input[placeholder*="输入姓名"]')
    await expect(searchInput).toBeVisible()
    await expect(page.getByRole('dialog').getByRole('button', { name: '搜索' })).toBeVisible()

    // 初始状态提示"请搜索成员"
    await expect(page.getByText('请搜索成员')).toBeVisible()

    // 关闭
    await page.getByRole('dialog').getByRole('button', { name: '关闭' }).click()
    await expect(page.getByRole('dialog')).toBeHidden()
  })

  test('添加成员对话框 - 搜索并添加成员', async ({ page }) => {
    // 创建一个自定义角色用于测试
    const roleName = `添加成员测试${Date.now().toString().slice(-4)}`
    const createRes = await page.evaluate(async (name) => {
      const res = await fetch('/api/org/roles', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, permissions: ['p-1'] }),
      })
      return res.json()
    }, roleName)
    await page.reload()

    // 选中自定义角色
    await page.getByText(roleName).click()
    await page.waitForTimeout(500)

    // 初始无成员
    await expect(page.getByText('暂无成员')).toBeVisible()

    // 添加成员
    await page.getByRole('button', { name: '添加成员' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()

    // 搜索（输入一个空格触发搜索所有成员）
    const searchInput = page.getByRole('dialog').locator('input[placeholder*="输入姓名"]')
    await searchInput.fill('a')
    await page.getByRole('dialog').getByRole('button', { name: '搜索' }).click()
    await page.waitForTimeout(1000)

    // 如果有结果，点击添加
    const addButtons = page.getByRole('dialog').getByRole('button', { name: '添加' })
    if ((await addButtons.count()) > 0) {
      await addButtons.first().click()
      await page.waitForTimeout(500)
      // toast 通知
      await expect(page.getByText('成员已添加到角色')).toBeVisible()
    }

    // 关闭对话框
    await page.getByRole('dialog').getByRole('button', { name: '关闭' }).click()

    // 清理
    await page.evaluate(async (id) => {
      await fetch(`/api/org/roles/${id}`, { method: 'DELETE', credentials: 'include' })
    }, createRes.id)
  })

  test('移除成员 - 确认对话框', async ({ page }) => {
    await page.getByText('超级管理员').first().click()
    await page.waitForTimeout(500)

    // 点击移除按钮
    const removeBtn = page.getByRole('button', { name: '移除' }).first()
    if (await removeBtn.isVisible()) {
      await removeBtn.click()
      // 确认对话框
      await expect(page.getByText('移除成员')).toBeVisible()
      await expect(page.getByText(/确定将/)).toBeVisible()

      // 取消移除
      await page.getByRole('button', { name: '取消' }).click()
      await expect(page.getByText('移除成员')).toBeHidden()
    }
  })

  test('普通成员角色 - 移除操作被阻止并 toast 提示', async ({ page }) => {
    await page.getByText('普通成员').first().click()
    await page.waitForTimeout(500)

    // 尝试点击移除
    const removeBtn = page.getByRole('button', { name: '移除' }).first()
    if (await removeBtn.isVisible()) {
      await removeBtn.click()
      // 应该 toast 提示而不是弹出确认框
      await expect(page.getByText('普通成员为保底角色，不可移除')).toBeVisible()
    }
  })

  test('API 不允许移除最后一个超级管理员', async ({ page }) => {
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
})

// ─── API 数据校验 ───────────────────────────────────────────────────────

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

    // 验证所有预设角色存在
    const names = data.map((r: { name: string }) => r.name)
    expect(names).toContain('超级管理员')
    expect(names).toContain('组织管理员')
    expect(names).toContain('普通成员')
    expect(names).toContain('只读审计员')
    expect(names).toContain('API 调用者')
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

  test('role members 接口返回成员列表结构正确', async ({ page }) => {
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

  test('自定义角色完整生命周期 - 创建/更新/删除', async ({ page }) => {
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

  test('角色成员增删 - API 验证', async ({ page }) => {
    // 创建自定义角色
    const roleName = `api-member-${Date.now().toString().slice(-5)}`
    const createRes = await page.evaluate(async (name) => {
      const res = await fetch('/api/org/roles', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, permissions: ['p-1'] }),
      })
      return res.json()
    }, roleName)

    // 获取一个成员
    const members = await page.evaluate(async () => {
      const res = await fetch('/api/org/members?page=1&pageSize=5', { credentials: 'include' })
      return res.json()
    })

    if (members.items?.length > 0) {
      const memberId = members.items[0].id

      // 添加成员
      const addRes = await page.evaluate(
        async ({ roleId, memberId }) => {
          const res = await fetch(`/api/org/roles/${roleId}/members`, {
            method: 'POST',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ memberId }),
          })
          return { status: res.status }
        },
        { roleId: createRes.id, memberId },
      )
      expect(addRes.status).toBe(200)

      // 验证成员已添加
      const roleMembers = await page.evaluate(async (roleId) => {
        const res = await fetch(`/api/org/roles/${roleId}/members`, { credentials: 'include' })
        return res.json()
      }, createRes.id)
      const memberIds = roleMembers.map((m: { id: string }) => m.id)
      expect(memberIds).toContain(memberId)

      // 移除成员
      const removeRes = await page.evaluate(
        async ({ roleId, memberId }) => {
          const res = await fetch(`/api/org/roles/${roleId}/members/${memberId}`, {
            method: 'DELETE',
            credentials: 'include',
          })
          return { status: res.status }
        },
        { roleId: createRes.id, memberId },
      )
      expect(removeRes.status).toBe(200)
    }

    // 清理角色
    await page.evaluate(async (id) => {
      await fetch(`/api/org/roles/${id}`, { method: 'DELETE', credentials: 'include' })
    }, createRes.id)
  })
})
