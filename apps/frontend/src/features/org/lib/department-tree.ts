import type { Department } from '@/api/types'

export type FlatDepartment = {
  id: string
  name: string
  level: number
}

export function flattenDepts(departments: Department[], level = 0): FlatDepartment[] {
  const result: FlatDepartment[] = []
  for (const dept of departments) {
    result.push({ id: dept.id, name: dept.name, level })
    if (dept.children) result.push(...flattenDepts(dept.children, level + 1))
  }
  return result
}
