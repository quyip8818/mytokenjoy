import { describe, expect, it, vi } from 'vitest'
import { act } from '@testing-library/react'
import type { Department, RoutingRule } from '@/api/types'
import { useModelRoutingPage } from '@/features/models/hooks/use-model-routing-page'
import { createMockApis, renderHookWithProviders } from '@tests/utils'
import { waitForLoaded } from '@tests/helpers/wait-for-loaded'
import { mockModels } from '@tests/fixtures/models'

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
    id: 'dept-1',
    nodeId: 'dept-1',
    nodeName: '总公司',
    inherited: false,
    allowedModelIds: [1, 2],
    defaultModelId: 1,
    fallbackModelId: null,
  },
  {
    id: 'dept-2',
    nodeId: 'dept-2',
    nodeName: '技术部',
    inherited: false,
    allowedModelIds: [1, 2],
    defaultModelId: 1,
    fallbackModelId: null,
  },
  {
    id: 'dept-3',
    nodeId: 'dept-3',
    nodeName: '后端组',
    inherited: true,
    allowedModelIds: [1],
    defaultModelId: 1,
    fallbackModelId: null,
  },
]

function createRoutingApis() {
  return createMockApis({
    routingApi: {
      getRules: vi.fn().mockResolvedValue(rules),
      updateRule: vi.fn().mockResolvedValue(rules[0]),
    },
    departmentApi: {
      getTree: vi.fn().mockResolvedValue(departments),
    },
    modelApi: {
      list: vi.fn().mockResolvedValue(mockModels),
    },
  })
}

describe('useModelRoutingPage', () => {
  it('loads rules, departments, and models on mount', async () => {
    const apis = createRoutingApis()

    const { result } = renderHookWithProviders(() => useModelRoutingPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(apis.routingApi.getRules).toHaveBeenCalled()
    expect(apis.departmentApi.getTree).toHaveBeenCalled()
    expect(apis.modelApi.list).toHaveBeenCalled()
    expect(result.current.rules).toEqual(rules)
    expect(result.current.departments).toEqual(departments)
    expect(result.current.models.length).toBeGreaterThan(0)
  })

  it('selects first department by default', async () => {
    const apis = createRoutingApis()

    const { result } = renderHookWithProviders(() => useModelRoutingPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    expect(result.current.selectedNodeId).toBe('dept-1')
    expect(result.current.selectedDepartment?.name).toBe('总公司')
  })

  it('finds selected rule for node', async () => {
    const apis = createRoutingApis()

    const { result } = renderHookWithProviders(() => useModelRoutingPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    act(() => {
      result.current.setSelectedNodeId('dept-2')
    })

    expect(result.current.selectedRule?.nodeId).toBe('dept-2')
    expect(result.current.selectedDepartment?.name).toBe('技术部')
  })

  it('resolves parent rule', async () => {
    const apis = createRoutingApis()

    const { result } = renderHookWithProviders(() => useModelRoutingPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    act(() => {
      result.current.setSelectedNodeId('dept-2')
    })

    expect(result.current.parentRule?.nodeId).toBe('dept-1')
  })

  it('handleSave calls routingApi.updateRule', async () => {
    const apis = createRoutingApis()

    const { result } = renderHookWithProviders(() => useModelRoutingPage(apis), { apis })

    await waitForLoaded(result, 'loading')

    await act(async () => {
      await result.current.handleSave({
        inherited: false,
        allowedModelIds: [1, 2],
        defaultModelId: 1,
        fallbackModelId: null,
      })
    })

    expect(apis.routingApi.updateRule).toHaveBeenCalledWith('dept-1', {
      inherited: false,
      allowedModelIds: [1, 2],
      defaultModelId: 1,
      fallbackModelId: null,
    })
  })
})
