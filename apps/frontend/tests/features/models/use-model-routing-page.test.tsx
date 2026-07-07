import { describe, expect, it, vi } from 'vitest'
import type { Department, RoutingRule } from '@/api/types'
import { useModelRoutingPage } from '@/features/models/hooks/use-model-routing-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'

const departments: Department[] = [
  {
    id: 'dept-1',
    name: '总公司',
    parentId: null,
    memberCount: 0,
    children: [
      {
        id: 'dept-2',
        name: '技术部',
        parentId: 'dept-1',
        memberCount: 0,
        children: [{ id: 'dept-3', name: '后端组', parentId: 'dept-2', memberCount: 0 }],
      },
    ],
  },
]

const rules: RoutingRule[] = [
  {
    id: 'rule-2',
    nodeId: 'dept-2',
    nodeName: '技术部',
    inherited: false,
    allowedModels: ['gpt-4o', 'claude-3'],
    defaultModel: 'gpt-4o',
    fallbackModel: null,
  },
  {
    id: 'rule-3',
    nodeId: 'dept-3',
    nodeName: '后端组',
    inherited: true,
    allowedModels: ['gpt-4o'],
    defaultModel: 'gpt-4o',
    fallbackModel: null,
  },
]

describe('useModelRoutingPage', () => {
  it('derives parent model count from department tree', async () => {
    const apis = createMockApis({
      routingApi: {
        getRules: vi.fn().mockResolvedValue(rules),
      },
      departmentApi: {
        getTree: vi.fn().mockResolvedValue(departments),
      },
    })

    const { result } = renderHookWithProviders(() => useModelRoutingPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    const childRule = rules[1]
    expect(result.current.getParentCount(childRule)).toBe(2)
  })
})
