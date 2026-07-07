import { useMemo, useState } from 'react'
import type { Department } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { DepartmentNode } from './department-node'
import { cn } from '@/lib/utils'
import { Plus, Search, Users } from 'lucide-react'

interface DepartmentPanelProps {
  tree: Department[]
  selectedId: string | undefined
  onSelect: (dept: Department | undefined) => void
  onCreateDept: (name: string, parentId: string) => Promise<void>
  onUpdateDept: (id: string, name: string) => Promise<void>
  onDeleteDept: (id: string) => Promise<void>
}

export function DepartmentPanel({
  tree,
  selectedId,
  onSelect,
  onCreateDept,
  onUpdateDept,
  onDeleteDept,
}: DepartmentPanelProps) {
  const [search, setSearch] = useState('')
  const [userExpanded, setUserExpanded] = useState<Set<string> | null>(null)
  const [dialogState, setDialogState] = useState<
    { type: 'add'; parentId: string } | { type: 'edit'; dept: Department } | null
  >(null)
  const [inputValue, setInputValue] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<Department | null>(null)
  const [deleteError, setDeleteError] = useState('')

  const defaultExpanded = useMemo(
    () => (tree.length > 0 ? new Set(tree.map((dept) => dept.id)) : new Set<string>()),
    [tree],
  )
  const expanded = userExpanded ?? defaultExpanded

  const toggleExpand = (id: string) => {
    setUserExpanded((prev) => {
      const current = prev ?? defaultExpanded
      const next = new Set(current)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const filteredTree = useMemo(() => {
    function filter(depts: Department[], kw: string): Department[] {
      if (!kw) return depts
      return depts.reduce<Department[]>((acc, dept) => {
        const children = dept.children ? filter(dept.children, kw) : []
        if (dept.name.includes(kw) || children.length > 0) {
          acc.push({ ...dept, children: children.length > 0 ? children : dept.children })
        }
        return acc
      }, [])
    }
    return filter(tree, search)
  }, [tree, search])

  const handleAdd = async () => {
    if (!inputValue.trim() || dialogState?.type !== 'add') return
    await onCreateDept(inputValue.trim(), dialogState.parentId)
    setDialogState(null)
    setInputValue('')
  }

  const handleEdit = async () => {
    if (!inputValue.trim() || dialogState?.type !== 'edit') return
    await onUpdateDept(dialogState.dept.id, inputValue.trim())
    setDialogState(null)
    setInputValue('')
  }

  const handleDelete = (dept: Department) => {
    if ((dept.children && dept.children.length > 0) || dept.memberCount > 0) {
      setDeleteError('请先移动或删除该部门下的子部门和成员')
      setDeleteTarget(dept)
      return
    }
    setDeleteError('')
    setDeleteTarget(dept)
  }

  const confirmDelete = async () => {
    if (!deleteTarget || deleteError) {
      setDeleteTarget(null)
      setDeleteError('')
      return
    }
    await onDeleteDept(deleteTarget.id)
    if (selectedId === deleteTarget.id) {
      onSelect(undefined)
    }
    setDeleteTarget(null)
  }

  return (
    <div className="flex w-64 shrink-0 flex-col border-r border-border">
      <div className="flex items-center gap-2 border-b border-border p-3">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="text"
            placeholder="搜索部门..."
            className="h-8 pl-8 text-sm"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <Button
          size="icon"
          variant="outline"
          aria-label="添加部门"
          className="size-8"
          onClick={() => {
            setDialogState({ type: 'add', parentId: tree[0]?.id ?? '' })
            setInputValue('')
          }}
        >
          <Plus className="size-3.5" />
        </Button>
      </div>

      <div
        role="treeitem"
        tabIndex={0}
        aria-selected={!selectedId}
        className={cn(
          'flex cursor-pointer items-center gap-2 border-b border-border px-3 py-2.5 text-sm',
          !selectedId ? 'bg-primary/8 text-primary' : 'text-foreground hover:bg-muted',
        )}
        onClick={() => onSelect(undefined)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            onSelect(undefined)
          }
        }}
      >
        <Users className="size-4 shrink-0 text-muted-foreground" />
        <span className="font-medium">全部成员</span>
      </div>

      <div className="flex-1 overflow-y-auto p-2">
        {filteredTree.map((dept) => (
          <DepartmentNode
            key={dept.id}
            department={dept}
            level={0}
            selectedId={selectedId}
            expandedIds={expanded}
            searchKeyword={search}
            onSelect={onSelect}
            onToggle={toggleExpand}
            onAdd={(parentId) => {
              setDialogState({ type: 'add', parentId })
              setInputValue('')
            }}
            onEdit={(department) => {
              setDialogState({ type: 'edit', dept: department })
              setInputValue(department.name)
            }}
            onDelete={handleDelete}
          />
        ))}
      </div>

      <Dialog
        open={dialogState !== null}
        onOpenChange={(open) => {
          if (!open) setDialogState(null)
        }}
      >
        <DialogContent className="sm:max-w-xs">
          <DialogHeader>
            <DialogTitle>{dialogState?.type === 'add' ? '添加子部门' : '重命名部门'}</DialogTitle>
          </DialogHeader>
          <Input
            placeholder="部门名称"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                if (dialogState?.type === 'add') {
                  void handleAdd()
                } else {
                  void handleEdit()
                }
              }
            }}
            autoFocus
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogState(null)}>
              取消
            </Button>
            <Button onClick={() => void (dialogState?.type === 'add' ? handleAdd() : handleEdit())}>
              确定
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) {
            setDeleteTarget(null)
            setDeleteError('')
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{deleteError ? '无法删除' : '删除部门'}</AlertDialogTitle>
            <AlertDialogDescription>
              {deleteError || `确定删除部门「${deleteTarget?.name ?? ''}」？`}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel
              onClick={() => {
                setDeleteTarget(null)
                setDeleteError('')
              }}
            >
              {deleteError ? '知道了' : '取消'}
            </AlertDialogCancel>
            {!deleteError && (
              <AlertDialogAction
                onClick={() => void confirmDelete()}
                className="bg-destructive text-white hover:bg-destructive/90"
              >
                删除
              </AlertDialogAction>
            )}
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
