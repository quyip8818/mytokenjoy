import { describe, expect, it } from 'vitest'
import type { BudgetNode } from '@/api/types'
import {
  BUDGET_DANGER_THRESHOLD,
  BUDGET_WARNING_THRESHOLD,
  computeUnallocated,
  findBudgetNode,
  formatBudgetPeriodLabel,
  getBudgetProgressTone,
  updateBudgetNodeInTree,
} from '@/features/budget/lib/mappers'

function makeNode(overrides: Partial<BudgetNode> & Pick<BudgetNode, 'id' | 'name'>): BudgetNode {
  return {
    parentId: null,
    budget: 1000,
    consumed: 100,
    period: '2026-01',
    ...overrides,
  }
}

describe('getBudgetProgressTone', () => {
  it('returns danger at or above danger threshold', () => {
    expect(getBudgetProgressTone(BUDGET_DANGER_THRESHOLD)).toBe('danger')
    expect(getBudgetProgressTone(100)).toBe('danger')
  })

  it('returns warning between warning and danger thresholds', () => {
    expect(getBudgetProgressTone(BUDGET_WARNING_THRESHOLD)).toBe('warning')
    expect(getBudgetProgressTone(BUDGET_DANGER_THRESHOLD - 1)).toBe('warning')
  })

  it('returns default below warning threshold', () => {
    expect(getBudgetProgressTone(BUDGET_WARNING_THRESHOLD - 1)).toBe('default')
  })
})

describe('computeUnallocated', () => {
  it('subtracts consumed, reserved pool, and children budgets', () => {
    const node = makeNode({
      id: 'n1',
      name: 'Root',
      budget: 1000,
      consumed: 200,
      reservedPool: 100,
      children: [makeNode({ id: 'c1', name: 'Child', budget: 300 })],
    })
    expect(computeUnallocated(node)).toBe(400)
  })

  it('never returns negative unallocated amount', () => {
    const node = makeNode({
      id: 'n1',
      name: 'Root',
      budget: 100,
      consumed: 500,
    })
    expect(computeUnallocated(node)).toBe(0)
  })
})

describe('findBudgetNode', () => {
  const tree: BudgetNode[] = [
    makeNode({
      id: 'root',
      name: 'Root',
      children: [makeNode({ id: 'child', name: 'Child' })],
    }),
  ]

  it('finds nested node by id', () => {
    expect(findBudgetNode(tree, 'child')?.name).toBe('Child')
  })

  it('returns null when id is missing', () => {
    expect(findBudgetNode(tree, 'missing')).toBeNull()
  })
})

describe('updateBudgetNodeInTree', () => {
  it('updates matching node in place', () => {
    const tree: BudgetNode[] = [makeNode({ id: 'n1', name: 'Before', budget: 500 })]
    expect(updateBudgetNodeInTree(tree, 'n1', { name: 'After', budget: 800 })).toBe(true)
    expect(tree[0].name).toBe('After')
    expect(tree[0].budget).toBe(800)
  })

  it('returns false when node is not found', () => {
    const tree: BudgetNode[] = [makeNode({ id: 'n1', name: 'Root' })]
    expect(updateBudgetNodeInTree(tree, 'missing', { name: 'X' })).toBe(false)
  })
})

describe('formatBudgetPeriodLabel', () => {
  it('formats YYYY-MM period from API', () => {
    expect(formatBudgetPeriodLabel('2026-06')).toBe('2026 年 6 月')
    expect(formatBudgetPeriodLabel(undefined)).toBe('-')
  })
})
