import { useCallback, useEffect, useMemo, useState } from 'react'
import type { Department, Member } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { ChevronRight, Folder, FolderOpen, Loader2, Search, Users } from 'lucide-react'

interface BudgetOrgMemberPickerProps {
  selectedIds: string[]
  onChange: (ids: string[]) => void
  defaultExpandDepartmentId?: string
  disabled?: boolean
  getDepartmentTree: () => Promise<Department[]>
  getMembers: (departmentId: string) => Promise<Member[]>
  searchMembers: (keyword: string) => Promise<Member[]>
}

export function BudgetOrgMemberPicker({
  selectedIds,
  onChange,
  defaultExpandDepartmentId,
  disabled = false,
  getDepartmentTree,
  getMembers,
  searchMembers,
}: BudgetOrgMemberPickerProps) {
  const [open, setOpen] = useState(false)
  const [tree, setTree] = useState<Department[]>([])
  const [treeLoading, setTreeLoading] = useState(false)
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set())
  const [loadedMembers, setLoadedMembers] = useState<Record<string, Member[]>>({})
  const [loadingDepts, setLoadingDepts] = useState<Set<string>>(new Set())
  const [search, setSearch] = useState('')
  const [searchResults, setSearchResults] = useState<Member[] | null>(null)
  const [searchLoading, setSearchLoading] = useState(false)
  const [selectedNames, setSelectedNames] = useState<Map<string, string>>(new Map())

  // Fetch tree when popover opens
  useEffect(() => {
    if (!open) return
    setTreeLoading(true)
    getDepartmentTree()
      .then((data) => {
        setTree(data ?? [])
        // Default expand the path to the target department
        if (defaultExpandDepartmentId) {
          const pathIds = findAncestorPath(data, defaultExpandDepartmentId)
          setExpandedIds(new Set(pathIds))
          // Auto-load members for the default department
          loadDeptMembers(defaultExpandDepartmentId)
        }
      })
      .finally(() => setTreeLoading(false))
  }, [open, getDepartmentTree, defaultExpandDepartmentId])

  const loadDeptMembers = useCallback(
    async (deptId: string) => {
      if (loadedMembers[deptId]) return
      setLoadingDepts((prev) => new Set([...prev, deptId]))
      try {
        const members = await getMembers(deptId)
        setLoadedMembers((prev) => ({ ...prev, [deptId]: members ?? [] }))
      } finally {
        setLoadingDepts((prev) => {
          const next = new Set(prev)
          next.delete(deptId)
          return next
        })
      }
    },
    [loadedMembers, getMembers],
  )

  const toggleExpand = useCallback(
    (deptId: string) => {
      setExpandedIds((prev) => {
        const next = new Set(prev)
        if (next.has(deptId)) {
          next.delete(deptId)
        } else {
          next.add(deptId)
          // Load members on first expand
          loadDeptMembers(deptId)
        }
        return next
      })
    },
    [loadDeptMembers],
  )

  const toggleMember = useCallback(
    (member: Member) => {
      const newNames = new Map(selectedNames)
      if (selectedIds.includes(member.id)) {
        onChange(selectedIds.filter((id) => id !== member.id))
        newNames.delete(member.id)
      } else {
        onChange([...selectedIds, member.id])
        newNames.set(member.id, member.name)
      }
      setSelectedNames(newNames)
    },
    [selectedIds, onChange, selectedNames],
  )

  // Search debounce
  useEffect(() => {
    if (!search.trim()) {
      setSearchResults(null)
      return
    }
    setSearchLoading(true)
    const timer = setTimeout(async () => {
      try {
        const members = await searchMembers(search.trim())
        setSearchResults(members ?? [])
      } finally {
        setSearchLoading(false)
      }
    }, 300)
    return () => clearTimeout(timer)
  }, [search, searchMembers])

  const selectedLabels = useMemo(() => {
    return selectedIds
      .map((id) => selectedNames.get(id) ?? id)
      .slice(0, 3)
  }, [selectedIds, selectedNames])

  return (
    <Popover open={open} onOpenChange={setOpen}>
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
              {selectedLabels.map((name) => (
                <Badge key={name} variant="outline" className="h-5 px-1.5 text-xs font-normal">
                  {name}
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
      <PopoverContent className="w-72 p-0" align="start">
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

        <div className="max-h-64 overflow-y-auto p-1">
          {treeLoading ? (
            <div className="flex items-center justify-center py-6">
              <Loader2 className="size-4 animate-spin text-muted-foreground" />
            </div>
          ) : search.trim() ? (
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
                onToggle={toggleExpand}
                loadedMembers={loadedMembers}
                loadingDepts={loadingDepts}
                selectedIds={selectedIds}
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
        <li key={member.id}>
          <MemberCheckRow
            member={member}
            checked={selectedIds.includes(member.id)}
            onToggle={() => onToggle(member)}
          />
        </li>
      ))}
    </ul>
  )
}

function MemberCheckRow({
  member,
  checked,
  onToggle,
}: {
  member: Member
  checked: boolean
  onToggle: () => void
}) {
  return (
    <label className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-muted">
      <Checkbox checked={checked} onCheckedChange={onToggle} aria-label={member.name} />
      <span className="flex-1 truncate text-xs">{member.name}</span>
      <span className="text-[11px] text-muted-foreground">{member.departmentName}</span>
    </label>
  )
}

function DeptTreeNode({
  dept,
  level,
  expandedIds,
  onToggle,
  loadedMembers,
  loadingDepts,
  selectedIds,
  onToggleMember,
}: {
  dept: Department
  level: number
  expandedIds: Set<string>
  onToggle: (id: string) => void
  loadedMembers: Record<string, Member[]>
  loadingDepts: Set<string>
  selectedIds: string[]
  onToggleMember: (member: Member) => void
}) {
  const hasChildren = dept.children && dept.children.length > 0
  const isExpanded = expandedIds.has(dept.id)
  const members = loadedMembers[dept.id]
  const isLoadingMembers = loadingDepts.has(dept.id)

  return (
    <div>
      <div
        role="treeitem"
        tabIndex={0}
        aria-expanded={isExpanded}
        className="flex cursor-pointer items-center gap-1.5 rounded-md px-1.5 py-1 text-xs hover:bg-muted"
        style={{ paddingLeft: `${level * 14 + 6}px` }}
        onClick={() => onToggle(dept.id)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            onToggle(dept.id)
          }
        }}
      >
        <span className="flex size-3.5 shrink-0 items-center justify-center">
          <ChevronRight
            className={cn(
              'size-3 text-muted-foreground transition-transform duration-150',
              isExpanded && 'rotate-90',
            )}
          />
        </span>
        {isExpanded ? (
          <FolderOpen className="size-3.5 shrink-0 text-muted-foreground" />
        ) : (
          <Folder className="size-3.5 shrink-0 text-muted-foreground" />
        )}
        <span className="flex-1 truncate font-medium text-foreground">{dept.name}</span>
      </div>

      {isExpanded && (
        <div>
          {/* Child departments */}
          {hasChildren &&
            dept.children!.map((child) => (
              <DeptTreeNode
                key={child.id}
                dept={child}
                level={level + 1}
                expandedIds={expandedIds}
                onToggle={onToggle}
                loadedMembers={loadedMembers}
                loadingDepts={loadingDepts}
                selectedIds={selectedIds}
                onToggleMember={onToggleMember}
              />
            ))}

          {/* Members of this department */}
          {isLoadingMembers && (
            <div
              className="flex items-center gap-2 py-1 text-xs text-muted-foreground"
              style={{ paddingLeft: `${(level + 1) * 14 + 6}px` }}
            >
              <Loader2 className="size-3 animate-spin" />
              加载中…
            </div>
          )}
          {members &&
            members.map((member) => (
              <div
                key={member.id}
                style={{ paddingLeft: `${(level + 1) * 14 + 6}px` }}
              >
                <MemberCheckRow
                  member={member}
                  checked={selectedIds.includes(member.id)}
                  onToggle={() => onToggleMember(member)}
                />
              </div>
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
