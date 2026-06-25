import { describe, expect, it } from 'vitest'
import type { Department } from '@/api/types'
import { findParentDeptId, flattenDepartments, flattenDepartmentTree, getDeptPath, filterDepartmentTree, getDeptDeleteError } from './org'

const departments: Department[] = [
  {
    id: 'd1',
    name: '总部',
    parentId: null,
    memberCount: 2,
    children: [
      {
        id: 'd2',
        name: '研发部',
        parentId: 'd1',
        memberCount: 1,
        children: [
          {
            id: 'd3',
            name: '前端组',
            parentId: 'd2',
            memberCount: 1,
            children: [],
          },
        ],
      },
    ],
  },
]

describe('flattenDepartments', () => {
  it('returns flat list with depth levels', () => {
    const flat = flattenDepartments(departments)
    expect(flat).toEqual([
      { id: 'd1', name: '总部', level: 0 },
      { id: 'd2', name: '研发部', level: 1 },
      { id: 'd3', name: '前端组', level: 2 },
    ])
  })
})

describe('flattenDepartmentTree', () => {
  it('returns all department nodes in preorder', () => {
    const flat = flattenDepartmentTree(departments)
    expect(flat.map((d) => d.id)).toEqual(['d1', 'd2', 'd3'])
  })
})

describe('getDeptPath', () => {
  it('builds breadcrumb path for nested department', () => {
    expect(getDeptPath(departments, 'd3')).toBe('总部 / 研发部 / 前端组')
  })

  it('returns null when department id is not found', () => {
    expect(getDeptPath(departments, 'missing')).toBeNull()
  })
})

describe('findParentDeptId', () => {
  it('finds direct parent id', () => {
    expect(findParentDeptId(departments, 'd2')).toBe('d1')
  })

  it('finds ancestor parent for nested department', () => {
    expect(findParentDeptId(departments, 'd3')).toBe('d2')
  })

  it('returns null for root department', () => {
    expect(findParentDeptId(departments, 'd1')).toBeNull()
  })
})

describe('filterDepartmentTree', () => {
  it('returns full tree when keyword is empty', () => {
    expect(filterDepartmentTree(departments, '')).toEqual(departments)
  })

  it('filters by department name', () => {
    const filtered = filterDepartmentTree(departments, '前端')
    expect(filtered).toHaveLength(1)
    expect(filtered[0]?.children?.[0]?.children?.[0]?.name).toBe('前端组')
  })
})

describe('getDeptDeleteError', () => {
  it('returns error when department has children', () => {
    expect(getDeptDeleteError(departments[0]!)).not.toBeNull()
  })

  it('returns null for leaf department without members', () => {
    const leaf = { id: 'x', name: '空部门', parentId: 'd1', memberCount: 0, children: [] }
    expect(getDeptDeleteError(leaf)).toBeNull()
  })
})
