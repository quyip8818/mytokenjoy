import { describe, expect, it } from 'vitest'
import {
  findParentDeptId,
  flattenDepartments,
  flattenDepartmentTree,
  getDeptPath,
  filterDepartmentTree,
  getDeptDeleteError,
} from '@/lib/org'
import { mockDepartmentTree } from '@tests/fixtures/departments'

describe('flattenDepartments', () => {
  it('returns flat list with depth levels', () => {
    const flat = flattenDepartments(mockDepartmentTree)
    expect(flat).toEqual([
      { id: 'd1', name: '总部', level: 0 },
      { id: 'd2', name: '研发部', level: 1 },
      { id: 'd3', name: '前端组', level: 2 },
    ])
  })
})

describe('flattenDepartmentTree', () => {
  it('returns all department nodes in preorder', () => {
    const flat = flattenDepartmentTree(mockDepartmentTree)
    expect(flat.map((d) => d.id)).toEqual(['d1', 'd2', 'd3'])
  })
})

describe('getDeptPath', () => {
  it('builds breadcrumb path for nested department', () => {
    expect(getDeptPath(mockDepartmentTree, 'd3')).toBe('总部 / 研发部 / 前端组')
  })

  it('returns null when department id is not found', () => {
    expect(getDeptPath(mockDepartmentTree, 'missing')).toBeNull()
  })
})

describe('findParentDeptId', () => {
  it('finds direct parent id', () => {
    expect(findParentDeptId(mockDepartmentTree, 'd2')).toBe('d1')
  })

  it('finds ancestor parent for nested department', () => {
    expect(findParentDeptId(mockDepartmentTree, 'd3')).toBe('d2')
  })

  it('returns null for root department', () => {
    expect(findParentDeptId(mockDepartmentTree, 'd1')).toBeNull()
  })
})

describe('filterDepartmentTree', () => {
  it('returns full tree when keyword is empty', () => {
    expect(filterDepartmentTree(mockDepartmentTree, '')).toEqual(mockDepartmentTree)
  })

  it('filters by department name', () => {
    const filtered = filterDepartmentTree(mockDepartmentTree, '前端')
    expect(filtered).toHaveLength(1)
    expect(filtered[0]?.children?.[0]?.children?.[0]?.name).toBe('前端组')
  })
})

describe('getDeptDeleteError', () => {
  it('returns error when department has children', () => {
    expect(getDeptDeleteError(mockDepartmentTree[0]!)).not.toBeNull()
  })

  it('returns null for leaf department without members', () => {
    const leaf = { id: 'x', name: '空部门', parentId: 'd1', memberCount: 0, children: [] }
    expect(getDeptDeleteError(leaf)).toBeNull()
  })
})
