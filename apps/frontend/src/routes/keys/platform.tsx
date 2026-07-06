import { useState, useEffect, useMemo } from 'react'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Input } from '@/components/ui/input'
import { platformKeyApi } from '@/api/keys'
import { departmentApi } from '@/api/org'
import type { Department, PlatformKey } from '@/api/types'
import { cn } from '@/lib/utils'
import { ChevronRight, Folder, FolderOpen, Key, Plus, Search, Users } from 'lucide-react'

// ─── Department Tree Panel ───────────────────────────────────────────────────

function TreeNode({
  node,
  depth,
  selectedId,
  onSelect,
  expanded,
  onToggle,
}: {
  node: Department
  depth: number
  selectedId: string | undefined
  onSelect: (id: string) => void
  expanded: Set<string>
  onToggle: (id: string) => void
}) {
  const hasChildren = node.children && node.children.length > 0
  const isExpanded = expanded.has(node.id)
  const isSelected = selectedId === node.id

  return (
    <>
      <div
        role="treeitem"
        tabIndex={0}
        aria-selected={isSelected}
        aria-expanded={hasChildren ? isExpanded : undefined}
        onClick={() => onSelect(node.id)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            onSelect(node.id)
          }
        }}
        className={cn(
          'group flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm',
          isSelected
            ? 'bg-primary/8 text-primary'
            : 'text-foreground hover:bg-muted'
        )}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
      >
        {hasChildren ? (
          <span
            role="button"
            tabIndex={-1}
            aria-label={isExpanded ? '收起' : '展开'}
            onClick={(e) => {
              e.stopPropagation()
              onToggle(node.id)
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.stopPropagation()
                onToggle(node.id)
              }
            }}
            className="flex size-4 shrink-0 items-center justify-center"
          >
            <ChevronRight
              className={cn(
                'size-3.5 text-muted-foreground transition-transform duration-150',
                isExpanded && 'rotate-90'
              )}
            />
          </span>
        ) : (
          <span className="size-4" />
        )}

        {hasChildren ? (
          isExpanded ? (
            <FolderOpen className="size-4 shrink-0 text-muted-foreground" />
          ) : (
            <Folder className="size-4 shrink-0 text-muted-foreground" />
          )
        ) : (
          <Users className="size-4 shrink-0 text-muted-foreground" />
        )}
        <span className="flex-1 truncate font-medium">{node.name}</span>
      </div>

      {isExpanded && hasChildren &&
        node.children!.map((child) => (
          <TreeNode
            key={child.id}
            node={child}
            depth={depth + 1}
            selectedId={selectedId}
            onSelect={onSelect}
            expanded={expanded}
            onToggle={onToggle}
          />
        ))}
    </>
  )
}

// ─── Key Table Component ─────────────────────────────────────────────────────

function KeyTable({
  keys,
  type,
  onRevoke,
}: {
  keys: PlatformKey[]
  type: 'member' | 'project'
  onRevoke: (id: string) => void
}) {
  const ownerLabel = type === 'member' ? '成员' : '项目'

  return (
    <Table>
      <TableHeader>
        <TableRow className="border-border/50 hover:bg-transparent">
          <TableHead className="text-xs font-medium text-muted-foreground">{ownerLabel}</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">Key 名称</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">所属部门</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">Key 前缀</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">状态</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground w-40">额度</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">模型白名单</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">到期时间</TableHead>
          <TableHead className="text-xs font-medium text-muted-foreground">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {keys.length === 0 && (
          <TableRow>
            <TableCell colSpan={9} className="h-32 text-center text-sm text-muted-foreground">
              <div className="flex flex-col items-center gap-2">
                <Key className="size-8 text-muted-foreground/50" />
                <span>暂无{ownerLabel} Key</span>
              </div>
            </TableCell>
          </TableRow>
        )}
        {keys.map((key) => {
          const pct = Math.round((key.used / key.quota) * 100)
          return (
            <TableRow key={key.id} className="border-border-subtle hover:bg-muted/50">
              <TableCell className="font-medium text-sm">
                {type === 'member' ? key.memberName : key.projectName}
              </TableCell>
              <TableCell className="text-sm">{key.name}</TableCell>
              <TableCell className="text-sm text-muted-foreground">{key.departmentName}</TableCell>
              <TableCell>
                <span className="font-mono text-xs px-1.5 py-0.5 rounded bg-muted text-muted-foreground">{key.keyPrefix}</span>
              </TableCell>
              <TableCell>{getStatusBadge(key.status)}</TableCell>
              <TableCell>
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <Progress value={pct} className="flex-1 h-1.5" />
                    <span className="text-xs tabular-nums text-muted-foreground w-8 text-right">{pct}%</span>
                  </div>
                  <span className="text-xs text-muted-foreground">
                    {key.quotaMode === 'periodic' ? '月度' : '固定'} · ¥{key.used.toLocaleString()}/{key.quota.toLocaleString()}
                  </span>
                </div>
              </TableCell>
              <TableCell>
                <div className="flex flex-wrap gap-1">
                  {key.modelWhitelist.slice(0, 2).map(m => (
                    <Badge key={m} variant="outline" className="text-xs">{m}</Badge>
                  ))}
                  {key.modelWhitelist.length > 2 && (
                    <Badge variant="outline" className="text-xs">+{key.modelWhitelist.length - 2}</Badge>
                  )}
                </div>
              </TableCell>
              <TableCell className="text-sm text-muted-foreground">{key.expiresAt ?? '永不'}</TableCell>
              <TableCell>
                {key.status === 'active' && (
                  <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={() => onRevoke(key.id)}>
                    吊销
                  </Button>
                )}
              </TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

function getStatusBadge(status: string) {
  switch (status) {
    case 'active':
      return <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700">正常</span>
    case 'disabled':
      return <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-600">已禁用</span>
    case 'expired':
      return <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-red-50 text-red-700">已过期</span>
    default:
      return <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-600">{status}</span>
  }
}

// ─── Main Page ───────────────────────────────────────────────────────────────
export default function PlatformKeysPage() {
  const [departments, setDepartments] = useState<Department[]>([])
  const [selectedDeptId, setSelectedDeptId] = useState<string | undefined>()
  const [keys, setKeys] = useState<PlatformKey[]>([])
  const [activeTab, setActiveTab] = useState<'member' | 'project'>('member')
  const [search, setSearch] = useState('')
  const [treeSearch, setTreeSearch] = useState('')
  const [expanded, setExpanded] = useState<Set<string>>(new Set())

  // Load department tree
  useEffect(() => {
    departmentApi.getTree().then(tree => {
      setDepartments(tree)
      // Auto-expand all & select root
      const ids = new Set<string>()
      function collect(nodes: Department[]) {
        for (const n of nodes) {
          ids.add(n.id)
          if (n.children) collect(n.children)
        }
      }
      collect(tree)
      setExpanded(ids)
      if (tree.length > 0) setSelectedDeptId(tree[0].id)
    })
  }, [])

  // Load keys when department or tab changes
  useEffect(() => {
    if (!selectedDeptId) return
    platformKeyApi.list({ departmentId: selectedDeptId, type: activeTab }).then(res => setKeys(res.items))
  }, [selectedDeptId, activeTab])

  const handleRevoke = async (id: string) => {
    await platformKeyApi.revoke(id)
    setKeys(keys.map(k => k.id === id ? { ...k, status: 'disabled' } : k))
  }

  const toggleExpand = (id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  // Filter tree by search
  const filteredTree = useMemo(() => {
    if (!treeSearch) return departments
    const lower = treeSearch.toLowerCase()
    function filter(nodes: Department[]): Department[] {
      return nodes.reduce<Department[]>((acc, n) => {
        const children = n.children ? filter(n.children) : []
        if (n.name.toLowerCase().includes(lower) || children.length > 0) {
          acc.push({ ...n, children: children.length > 0 ? children : n.children })
        }
        return acc
      }, [])
    }
    return filter(departments)
  }, [departments, treeSearch])

  const effectiveExpanded = useMemo(() => {
    if (!treeSearch) return expanded
    const ids = new Set<string>()
    function collectAll(nodes: Department[]) {
      for (const n of nodes) {
        ids.add(n.id)
        if (n.children) collectAll(n.children)
      }
    }
    collectAll(filteredTree)
    return ids
  }, [treeSearch, filteredTree, expanded])

  // Filter keys by search text
  const filteredKeys = useMemo(() => {
    if (!search) return keys
    const lower = search.toLowerCase()
    return keys.filter(k => {
      const owner = k.type === 'member' ? k.memberName : k.projectName
      return (
        k.name.toLowerCase().includes(lower) ||
        (owner?.toLowerCase().includes(lower) ?? false) ||
        k.keyPrefix.toLowerCase().includes(lower)
      )
    })
  }, [keys, search])

  return (
    <div className="flex h-[calc(100dvh-7.5rem)] rounded-lg border border-border bg-card shadow-xs overflow-hidden">
      {/* Left: Department Tree */}
      <div className="flex w-64 shrink-0 flex-col border-r border-border">
        <div className="border-b border-border p-3">
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={treeSearch}
              onChange={(e) => setTreeSearch(e.target.value)}
              placeholder="搜索部门..."
              className="h-8 pl-8 text-sm"
            />
          </div>
        </div>
        <div className="flex-1 overflow-y-auto p-2" role="tree">
          {filteredTree.map((node) => (
            <TreeNode
              key={node.id}
              node={node}
              depth={0}
              selectedId={selectedDeptId}
              onSelect={setSelectedDeptId}
              expanded={effectiveExpanded}
              onToggle={toggleExpand}
            />
          ))}
        </div>
      </div>

      {/* Right: Key Content */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Toolbar */}
        <div className="flex items-center justify-between border-b border-border px-5 py-3">
          <div className="flex items-center gap-1">
            <button
              onClick={() => setActiveTab('member')}
              className={cn(
                'px-3 py-1.5 rounded-md text-sm font-medium transition-colors duration-100',
                activeTab === 'member'
                  ? 'bg-muted text-foreground'
                  : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
              )}
            >
              成员 Key
            </button>
            <button
              onClick={() => setActiveTab('project')}
              className={cn(
                'px-3 py-1.5 rounded-md text-sm font-medium transition-colors duration-100',
                activeTab === 'project'
                  ? 'bg-muted text-foreground'
                  : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
              )}
            >
              项目 Key
            </button>
          </div>

          <div className="flex items-center gap-3">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="搜索 Key..."
                className="h-8 w-52 pl-8 text-sm"
              />
            </div>
            <Button size="sm" className="h-8 gap-1.5">
              <Plus className="size-3.5" />
              签发 Key
            </Button>
          </div>
        </div>

        {/* Table */}
        <div className="flex-1 overflow-auto px-5 py-4">
          <KeyTable keys={filteredKeys} type={activeTab} onRevoke={handleRevoke} />
        </div>
      </div>
    </div>
  )
}
