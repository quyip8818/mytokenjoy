import { create } from 'zustand'
import type { Department, Member, Paginated } from '@/api/types'
import { departmentApi, memberApi } from '@/api/org'
import type { RowSelectionState } from '@tanstack/react-table'

interface OrgStructureState {
  // Department state
  departments: Department[]
  selectedDept: Department | undefined
  expanded: Set<string>
  deptLoading: boolean

  // Member state
  members: Member[]
  total: number
  page: number
  pageSize: number
  keyword: string
  membersLoading: boolean
  rowSelection: RowSelectionState

  // Computed
  pendingCount: number
  selectedIds: string[]

  // Department actions
  loadDepartments: () => Promise<void>
  selectDept: (dept: Department | undefined) => void
  toggleExpand: (id: string) => void
  createDept: (name: string, parentId: string) => Promise<void>
  updateDept: (id: string, name: string) => Promise<void>
  deleteDept: (id: string) => Promise<void>

  // Member actions
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

export const useOrgStructureStore = create<OrgStructureState>((set, get) => ({
  // Initial state
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

  // Department actions
  loadDepartments: async () => {
    set({ deptLoading: true })
    try {
      const departments = await departmentApi.getTree()
      set({ departments })
    } finally {
      set({ deptLoading: false })
    }
  },

  selectDept: (dept) => {
    set({ selectedDept: dept, page: 1, rowSelection: {} })
    // Auto-load members when department changes
    get().loadMembers()
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
    await departmentApi.create({ name, parentId })
    await get().loadDepartments()
  },

  updateDept: async (id, name) => {
    await departmentApi.update(id, { name })
    await get().loadDepartments()
  },

  deleteDept: async (id) => {
    await departmentApi.delete(id)
    const { selectedDept } = get()
    if (selectedDept?.id === id) {
      set({ selectedDept: undefined })
    }
    await get().loadDepartments()
    await get().loadMembers()
  },

  // Member actions
  loadMembers: async () => {
    const { page, pageSize, keyword, selectedDept } = get()
    set({ membersLoading: true })
    try {
      const params: Parameters<typeof memberApi.list>[0] = {
        page,
        pageSize,
        keyword: keyword || undefined,
      }
      if (selectedDept) {
        params.departmentId = selectedDept.id
      }
      const res: Paginated<Member> = await memberApi.list(params)
      set({ members: res.items, total: res.total })
    } finally {
      set({ membersLoading: false })
    }
  },

  setPage: (page) => {
    set({ page })
    get().loadMembers()
  },

  setKeyword: (keyword) => {
    set({ keyword, page: 1, rowSelection: {} })
    get().loadMembers()
  },

  setRowSelection: (selection) => set({ rowSelection: selection }),

  createMember: async (data) => {
    await memberApi.create(data)
    await get().loadMembers()
    await get().loadDepartments()
  },

  updateMember: async (id, data) => {
    await memberApi.update(id, data)
    await get().loadMembers()
  },

  deleteMember: async (ids) => {
    await memberApi.delete(ids)
    set({ rowSelection: {} })
    await get().loadMembers()
    await get().loadDepartments()
  },

  updateMemberStatus: async (ids, status) => {
    await memberApi.updateStatus(ids, status)
    set({ rowSelection: {} })
    await get().loadMembers()
  },

  transferMembers: async (ids, departmentId) => {
    await memberApi.transferDepartment(ids, departmentId)
    set({ rowSelection: {} })
    await get().loadMembers()
    await get().loadDepartments()
  },

  inviteMember: async (value) => {
    const isEmail = value.includes('@')
    await memberApi.invite(isEmail ? { email: value } : { phone: value })
    await get().loadMembers()
  },
}))
