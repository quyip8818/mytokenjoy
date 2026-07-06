import type { Department } from '@/api/types'

export const mockDepartments: Department[] = [
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
        children: [],
      },
    ],
  },
]

export const mockDepartmentTree: Department[] = [
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
