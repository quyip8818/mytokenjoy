import { useCallback, useEffect, useMemo, useState } from 'react'
import type { Department, Member } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { ChevronRight, Loader2, Search, Users } from 'lucide-react'

interface BudgetOrgMemberPickerProps {
  selectedIds: string[]
  onChange: (ids: string[]) => void
  defaultExpandDepartmentId?: string
  disabled?: boolean
  getDepartmentTree: () => Promise<Department[]>
  /** Returns direct members only (directOnly: true) */
  getMembers: (departmentId: string) => Promise<Member[]>
  /** Returns all recursive members (no directOnly) */
  getAllDeptMembers: (departmentId: string) => Promise<Member[]>
  searchMembers: (keyword: string) => Promise<Member[]>
}

export function BudgetOrgMemberPicker({
  selectedIds,
  onChange,
  defaultExpandDepartmentId,
  disabled = false,
  getDepartmentTree,
  getMembers,
  getAllDeptMembers,
  searchMembers,
}: BudgetOrgMemberPickerProps) {
  const [open, setOpen] = useState(false)
  const [tree, setTree] = useState<Department[]>([])
  const [treeLoading, setTreeLoading] = useState(false)
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set())
  // Direct members per department (shown in expanded list)
  const [directMembers, setDirectMembers] = useState<Record<string, Member[]>>({})
  // All recursive members per department (used for dept checkbox toggle)
  const [allMembers, setAllMembers] = useState<Record<string, Member[]>>({})
  const [loadingDepts, setLoadingDepts] = useState<Set<string>>(new Set())
  const [search, setSearch] = useState('')
  const [searchResults, setSearchResults] = useState<Member[] | null>(null)
  const [lastFetchedSearch, setLastFetchedSearch] = useState('')
  const [selectedNames, setSelectedNames] = useState<Map<string, string>>(new Map())

  const trimmedSearch = search.trim()
  const isSearching = trimmedSearch.length > 0
  const searchLoading = isSearching && lastFetchedSearch !== trimmedSearch

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      setOpen(nextOpen)
      if (!nextOpen) return

      setTreeLoading(true)
      getDepartmentTree()
        .then((data) => {
          setTree(data ?? [])
          if (defaultExpandDepartmentId) {
            const pathIds = findAncestorPath(data, defaultExpandDepartmentId)
            setExpandedIds(new Set(pathIds))
          }
        })
        .finally(() => setTreeLoading(false))
    },
    [getDepartmentTree, defaultExpandDepartmentId],
  )

  // Load direct members when expanding a department
  const loadDirectMembers = useCallback(
    async (deptId: string) => {
      if (directMembers[deptId]) return
      setLoadingDepts((prev) => new Set([...prev, deptId]))
      try {
        const members = await getMembers(deptId)
        setDirectMembers((prev) => ({ ...prev, [deptId]: members ?? [] }))
      } finally {
        setLoadingDepts((prev) => {
          const next = new Set(prev)
          next.delete(deptId)
          return next
        })
      }
    },
    [directMembers, getMembers],
  )

  // Load all recursive members for department checkbox
  const loadAllMembers = useCallback(
    async (deptId: string): Promise<Member[]> => {
      if (allMembers[deptId]) return allMembers[deptId]
      const members = await getAllDeptMembers(deptId)
      const result = members ?? []
      setAllMembers((prev) => ({ ...prev, [deptId]: result }))
      return result
    },
    [allMembers, getAllDeptMembers],
  )

  const toggleExpand = useCallback(
    (deptId: string) => {
      setExpandedIds((prev) => {
        const next = new Set(prev)
        if (next.has(deptId)) {
          next.delete(deptId)
        } else {
          next.add(deptId)
          loadDirectMembers(deptId)
        }
        return next
      })
    },
    [loadDirectMembers],
  )

  const toggleDepartment = useCallback(
    async (deptId: string) => {
      setLoadingDepts((prev) => new Set([...prev, deptId]))
      try {
        const members = await loadAllMembers(deptId)
        if (!members || members.length === 0) return

        const memberIds = members.map((m) => m.id)
        const allSelected = memberIds.every((id) => selectedIds.includes(id))
        const newNames = new Map(selectedNames)

        if (allSelected) {
          const remaining = selectedIds.filter((id) => !memberIds.includes(id))
          for (const id of memberIds) newNames.delete(id)
          onChange(remaining)
        } else {
          const toAdd = memberIds.filter((id) => !selectedIds.includes(id))
          for (const m of members) newNames.set(m.id, m.alias)
          onChange([...selectedIds, ...toAdd])
        }
        setSelectedNames(newNames)
      } finally {
        setLoadingDepts((prev) => {
          const next = new Set(prev)
          next.delete(deptId)
          return next
        })
      }
    },
    [loadAllMembers, selectedIds, onChange, selectedNames],
  )

  const toggleMember = useCallback(
    (member: Member) => {
      const newNames = new Map(selectedNames)
      if (selectedIds.includes(member.id)) {
        onChange(selectedIds.filter((id) => id !== member.id))
        newNames.delete(member.id)
      } else {
        onChange([...selectedIds, member.id])
        newNames.set(member.id, member.alias)
      }
      setSelectedNames(newNames)
    },
    [selectedIds, onChange, selectedNames],
  )

  // Search debounce
  useEffect(() => {
    if (!isSearching) return
    let cancelled = false
    const timer = setTimeout(async () => {
      try {
        const members = await searchMembers(trimmedSearch)
        if (!cancelled) {
          setSearchResults(members ?? [])
          setLastFetchedSearch(trimmedSearch)
        }
      } catch {
        if (!cancelled) setLastFetchedSearch(trimmedSearch)
      }
    }, 300)
    return () => {
      cancelled = true
      clearTimeout(timer)
    }
  }, [isSearching, trimmedSearch, searchMembers])

  const selectedLabels = useMemo(() => {
    return selectedIds.slice(0, 3).map((id) => ({ id, name: selectedNames.get(id) ?? id }))
  }, [selectedIds, selectedNames])

  // Compute dept checkbox state from allMembers cache
  const getDeptCheckState = useCallback(
    (deptId: string): boolean | 'indeterminate' => {
      const members = allMembers[deptId]
      if (!members || members.length === 0) return false
      const memberIds = members.map((m) => m.id)
      const allSelected = memberIds.every((id) => selectedIds.includes(id))
      if (allSelected) return true
      const someSelected = memberIds.some((id) => selectedIds.includes(id))
      return someSelected ? 'indeterminate' : false
    },
    [allMembers, selectedIds],
  )

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={cn(
            'h-8 w-full justify-start gap-2 font-normal',
            !selectedIds.length && 'text-muted-foreground',
          )}
          disabled={disabled}
          aria-label="选择关联成员"
        >
          <Users className="size-4 shrink-0" />
          {selectedIds.length === 0 ? (
            <span>选择成员…</span>
          ) : (
            <span className="flex flex-wrap gap-1">
              {selectedLabels.map((item) => (
                <Badge key={item.id} variant="outline" className="h-5 px-1.5 text-xs font-normal">
                  {item.name}
                </Badge>
              ))}
              {selectedIds.length > 3 && (
                <Badge variant="outline" className="h-5 px-1.5 text-xs font-normal">
                  +{selectedIds.length - 3}
                </Badge>
              )}
            </span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className="w-72 p-0"
        align="start"
        onOpenAutoFocus={(e) => e.preventDefault()}
      >
        <div className="border-b border-border p-2">
          <div className="relative">
            <Search className="absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="搜索成员..."
              className="h-7 pl-7 text-xs"
            />
          </div>
        </div>

        <div
          className="max-h-64 overflow-y-auto overscroll-contain p-1"
          onWheel={(e) => e.stopPropagation()}
        >
          {treeLoading ? (
            <div className="flex items-center justify-center py-6">
              <Loader2 className="size-4 animate-spin text-muted-foreground" />
            </div>
          ) : isSearching ? (
            <SearchResultList
              results={searchResults}
              loading={searchLoading}
              selectedIds={selectedIds}
              onToggle={toggleMember}
            />
          ) : (
            tree.map((dept) => (
              <DeptTreeNode
                key={dept.id}
                dept={dept}
                level={0}
                expandedIds={expandedIds}
                selectedIds={selectedIds}
                directMembers={directMembers}
                loadingDepts={loadingDepts}
                getDeptCheckState={getDeptCheckState}
                onToggleExpand={toggleExpand}
                onToggleDepartment={toggleDepartment}
                onToggleMember={toggleMember}
              />
            ))
          )}
        </div>

        {selectedIds.length > 0 && (
          <div className="border-t border-border px-3 py-1.5">
            <span className="text-xs text-muted-foreground">已选 {selectedIds.length} 人</span>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}

// --- Helper components ---

function SearchResultList({
  results,
  loading,
  selectedIds,
  onToggle,
}: {
  results: Member[] | null
  loading: boolean
  selectedIds: string[]
  onToggle: (member: Member) => void
}) {
  if (loading) {
    return (
      <div className="flex items-center justify-center py-6">
        <Loader2 className="size-4 animate-spin text-muted-foreground" />
      </div>
    )
  }
  if (!results || results.length === 0) {
    return <p className="px-2 py-4 text-center text-xs text-muted-foreground">未找到匹配成员</p>
  }
  return (
    <ul className="space-y-0.5">
      {results.map((member) => (
        <MemberRow
          key={member.id}
          member={member}
          checked={selectedIds.includes(member.id)}
          onToggle={() => onToggle(member)}
          indent={0}
        />
      ))}
    </ul>
  )
}

function MemberRow({
  member,
  checked,
  onToggle,
  indent,
}: {
  member: Member
  checked: boolean
  onToggle: () => void
  indent: number
}) {
  return (
    <label
      className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1 text-xs hover:bg-muted"
      style={{ paddingLeft: `${indent * 14 + 24}px` }}
    >
      <Checkbox
        checked={checked}
        onCheckedChange={onToggle}
        className="size-3.5"
        aria-label={member.alias}
      />
      <span className="flex-1 truncate">{member.alias}</span>
      <span className="text-[11px] text-muted-foreground">{member.departmentName}</span>
    </label>
  )
}

function DeptTreeNode({
  dept,
  level,
  expandedIds,
  selectedIds,
  directMembers,
  loadingDepts,
  getDeptCheckState,
  onToggleExpand,
  onToggleDepartment,
  onToggleMember,
}: {
  dept: Department
  level: number
  expandedIds: Set<string>
  selectedIds: string[]
  directMembers: Record<string, Member[]>
  loadingDepts: Set<string>
  getDeptCheckState: (id: string) => boolean | 'indeterminate'
  onToggleExpand: (id: string) => void
  onToggleDepartment: (id: string) => void
  onToggleMember: (member: Member) => void
}) {
  const hasChildren = dept.children && dept.children.length > 0
  const isExpanded = expandedIds.has(dept.id)
  const isLoading = loadingDepts.has(dept.id)
  const members = directMembers[dept.id]
  const checkState = getDeptCheckState(dept.id)

  return (
    <div>
      <div
        className="flex items-center gap-1 rounded-md px-1.5 py-1 text-xs hover:bg-muted"
        style={{ paddingLeft: `${level * 14 + 6}px` }}
      >
        {/* Expand/collapse arrow */}
        <span
          role="button"
          tabIndex={-1}
          aria-label={isExpanded ? '收起' : '展开'}
          className="flex size-4 shrink-0 cursor-pointer items-center justify-center"
          onClick={() => onToggleExpand(dept.id)}
        >
          <ChevronRight
            className={cn(
              'size-3 text-muted-foreground transition-transform duration-150',
              isExpanded && 'rotate-90',
            )}
          />
        </span>

        {/* Checkbox */}
        <Checkbox
          checked={checkState}
          onCheckedChange={() => onToggleDepartment(dept.id)}
          className="size-3.5"
          aria-label={`选择${dept.name}`}
        />

        {/* Department name (click also toggles) */}
        <span
          className="flex-1 cursor-pointer truncate font-medium text-foreground"
          onClick={() => onToggleDepartment(dept.id)}
        >
          {dept.name}
        </span>

        {isLoading && <Loader2 className="size-3 animate-spin text-muted-foreground" />}
      </div>

      {/* Expanded content: sub-departments + direct members */}
      {isExpanded && (
        <div>
          {hasChildren &&
            dept.children!.map((child) => (
              <DeptTreeNode
                key={child.id}
                dept={child}
                level={level + 1}
                expandedIds={expandedIds}
                selectedIds={selectedIds}
                directMembers={directMembers}
                loadingDepts={loadingDepts}
                getDeptCheckState={getDeptCheckState}
                onToggleExpand={onToggleExpand}
                onToggleDepartment={onToggleDepartment}
                onToggleMember={onToggleMember}
              />
            ))}

          {/* Direct members of this department */}
          {members &&
            members.map((member) => (
              <MemberRow
                key={member.id}
                member={member}
                checked={selectedIds.includes(member.id)}
                onToggle={() => onToggleMember(member)}
                indent={level + 1}
              />
            ))}
        </div>
      )}
    </div>
  )
}

// Find the path of ancestor IDs from root to the target department
function findAncestorPath(tree: Department[], targetId: string): string[] {
  function dfs(nodes: Department[], path: string[]): string[] | null {
    for (const node of nodes) {
      const currentPath = [...path, node.id]
      if (node.id === targetId) return currentPath
      if (node.children) {
        const result = dfs(node.children, currentPath)
        if (result) return result
      }
    }
    return null
  }
  return dfs(tree, []) ?? []
}
