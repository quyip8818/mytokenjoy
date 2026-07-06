import type { Department } from '@/api/types'

export interface FlatDepartment {
  id: string
  name: string
  level: number
}

export function flattenDepartments(departments: Department[], level = 0): FlatDepartment[] {
  const result: FlatDepartment[] = []
  for (const dept of departments) {
    result.push({ id: dept.id, name: dept.name, level })
    if (dept.children) result.push(...flattenDepartments(dept.children, level + 1))
  }
  return result
}

export function flattenDepartmentTree(departments: Department[]): Department[] {
  const result: Department[] = []
  for (const dept of departments) {
    result.push(dept)
    if (dept.children) result.push(...flattenDepartmentTree(dept.children))
  }
  return result
}

export function getDeptPath(departments: Department[], targetId: string): string | null {
  function walk(nodes: Department[], path: string[]): string | null {
    for (const node of nodes) {
      const current = [...path, node.name]
      if (node.id === targetId) return current.join(' / ')
      if (node.children) {
        const found = walk(node.children, current)
        if (found) return found
      }
    }
    return null
  }
  return walk(departments, [])
}

export function findParentDeptId(departments: Department[], deptId: string): string | null {
  for (const dept of departments) {
    if (dept.children?.some((c) => c.id === deptId)) return dept.id
    if (dept.children) {
      const found = findParentDeptId(dept.children, deptId)
      if (found) return found
    }
  }
  return null
}

export function filterDepartmentTree(departments: Department[], keyword: string): Department[] {
  if (!keyword) return departments
  return departments.reduce<Department[]>((acc, dept) => {
    const childMatches = dept.children ? filterDepartmentTree(dept.children, keyword) : []
    if (dept.name.includes(keyword) || childMatches.length > 0) {
      acc.push({ ...dept, children: childMatches.length > 0 ? childMatches : dept.children })
    }
    return acc
  }, [])
}

export function getDeptDeleteError(dept: Department): string | null {
  if ((dept.children && dept.children.length > 0) || dept.memberCount > 0) {
    return '请先移动或删除该部门下的子部门和成员'
  }
  return null
}

export function buildDeptParentMap(
  departments: Department[],
  map = new Map<string, string | null>(),
): Map<string, string | null> {
  for (const dept of departments) {
    map.set(dept.id, dept.parentId)
    if (dept.children) buildDeptParentMap(dept.children, map)
  }
  return map
}
