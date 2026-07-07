import { expect, test } from '@playwright/test'

const routes = [
  { path: '/org/data-source', heading: '数据源' },
  { path: '/org/structure', heading: '组织架构' },
  { path: '/org/roles', heading: '角色管理' },
  { path: '/budget', heading: '预算管理' },
  { path: '/budget/alerts', heading: '预警规则' },
  { path: '/models/list', heading: '模型列表' },
  { path: '/models/routing', heading: '模型白名单' },
  { path: '/keys/mine', heading: '我的 Key' },
  { path: '/keys/approval', heading: '审批中心' },
  { path: '/keys/platform', heading: 'Key 管理' },
  { path: '/keys/provider', heading: '供应商 Key' },
  { path: '/dashboard/cost', heading: '成本看板' },
  { path: '/dashboard/usage', heading: '用量分析' },
  { path: '/wallet', heading: '钱包管理' },
  { path: '/audit/operations', heading: '操作审计' },
  { path: '/audit/calls', heading: '调用日志' },
]

for (const { path, heading } of routes) {
  test(`${path} renders heading "${heading}"`, async ({ page }) => {
    await page.goto(path)
    await expect(
      page.getByRole('banner').getByRole('heading', { name: heading }),
    ).toBeVisible()
  })
}
