import { create } from 'zustand'
import type { AppApis } from '@/api/app-apis'
import { defaultApis } from '@/api/app-apis'
import type { Department, Member, Paginated } from '@/api/types'
import type { RowSelectionState } from '@tanstack/react-table'

interface OrgStructureState {
  departments: Department[]
  selectedDept: Department | undefined
  expanded: Set<string>
  deptLoading: boolean

  members: Member[]
  total: number
  page: number
  pageSize: number
  keyword: string
  membersLoading: boolean
  rowSelection: RowSelectionState

  pendingCount: number
  selectedIds: string[]

  loadDepartments: () => Promise<void>
  selectDept: (dept: Department | undefined) => void
  toggleExpand: (id: string) => void
  createDept: (name: string, parentId: string) => Promise<void>
  updateDept: (id: string, name: string) => Promise<void>
  deleteDept: (id: string) => Promise<void>

  loadMembers: () => Promise<void>
  setPage: (page: number) => void
  setKeyword: (keyword: string) => void
  setRowSelection: (selection: RowSelectionState) => void
  createMember: (data: Omit<Member, 'id' | 'status' | 'roles' | 'source'>) => Promise<void>
  updateMember: (id: string, data: Partial<Member>) => Promise<void>
  deleteMember: (ids: string[]) => Promise<void>
  updateMemberStatus: (ids: string[], status: 'active' | 'inactive') => Promise<void>
  transferMembers: (ids: string[], departmentId: string) => Promise<void>
  inviteMember: (value: string) => Promise<void>
}

let structureApis: AppApis = defaultApis

export function initOrgStructureApis(apis: AppApis) {
  structureApis = apis
}

function getApis() {
  return structureApis
}

export const useOrgStructureStore = create<OrgStructureState>((set, get) => ({
  departments: [],
  selectedDept: undefined,
  expanded: new Set<string>(),
  deptLoading: false,

  members: [],
  total: 0,
  page: 1,
  pageSize: 10,
  keyword: '',
  membersLoading: false,
  rowSelection: {},

  get pendingCount() {
    return get().members.filter((m) => m.status === 'pending').length
  },
  get selectedIds() {
    return Object.keys(get().rowSelection)
  },

  loadDepartments: async () => {
    const apis = getApis()
    set({ deptLoading: true })
    try {
      const departments = await apis.departmentApi.getTree()
      set({ departments })
    } finally {
      set({ deptLoading: false })
    }
  },

  selectDept: (dept) => {
    set({ selectedDept: dept, page: 1, rowSelection: {} })
    void get().loadMembers()
  },

  toggleExpand: (id) => {
    set((state) => {
      const next = new Set(state.expanded)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return { expanded: next }
    })
  },

  createDept: async (name, parentId) => {
    const apis = getApis()
    await apis.departmentApi.create({ name, parentId })
    await get().loadDepartments()
  },

  updateDept: async (id, name) => {
    const apis = getApis()
    await apis.departmentApi.update(id, { name })
    await get().loadDepartments()
  },

  deleteDept: async (id) => {
    const apis = getApis()
    await apis.departmentApi.delete(id)
    const { selectedDept } = get()
    if (selectedDept?.id === id) {
      set({ selectedDept: undefined })
    }
    await get().loadDepartments()
    await get().loadMembers()
  },

  loadMembers: async () => {
    const apis = getApis()
    const { page, pageSize, keyword, selectedDept } = get()
    set({ membersLoading: true })
    try {
      const params: Parameters<typeof apis.memberApi.list>[0] = {
        page,
        pageSize,
        keyword: keyword || undefined,
      }
      if (selectedDept) {
        params.departmentId = selectedDept.id
      }
      const res: Paginated<Member> = await apis.memberApi.list(params)
      set({ members: res.items, total: res.total })
    } finally {
      set({ membersLoading: false })
    }
  },

  setPage: (page) => {
    set({ page })
    void get().loadMembers()
  },

  setKeyword: (keyword) => {
    set({ keyword, page: 1, rowSelection: {} })
    void get().loadMembers()
  },

  setRowSelection: (selection) => set({ rowSelection: selection }),

  createMember: async (data) => {
    const apis = getApis()
    await apis.memberApi.create(data)
    await get().loadMembers()
    await get().loadDepartments()
  },

  updateMember: async (id, data) => {
    const apis = getApis()
    await apis.memberApi.update(id, data)
    await get().loadMembers()
  },

  deleteMember: async (ids) => {
    const apis = getApis()
    await apis.memberApi.delete(ids)
    set({ rowSelection: {} })
    await get().loadMembers()
    await get().loadDepartments()
  },

  updateMemberStatus: async (ids, status) => {
    const apis = getApis()
    await apis.memberApi.updateStatus(ids, status)
    set({ rowSelection: {} })
    await get().loadMembers()
  },

  transferMembers: async (ids, departmentId) => {
    const apis = getApis()
    await apis.memberApi.transferDepartment(ids, departmentId)
    set({ rowSelection: {} })
    await get().loadMembers()
    await get().loadDepartments()
  },

  inviteMember: async (value) => {
    const apis = getApis()
    const isEmail = value.includes('@')
    await apis.memberApi.invite(isEmail ? { email: value } : { phone: value })
    await get().loadMembers()
  },
}))
