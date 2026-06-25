import { useState, useMemo } from 'react'
import { Plus } from 'lucide-react'
import type { Department } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
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
import { filterDepartmentTree, getDeptDeleteError } from '@/lib/org'
import { DepartmentTreeNode } from '@/routes/org/components/department-tree-node'

interface DepartmentTreeProps {
  departments: Department[]
  selectedId: string | undefined
  onSelect: (dept: Department | undefined) => void
  onAddChildDept: (parentId: string, parentName: string) => void
  onUpdateDeptName: (id: string, name: string) => void | Promise<void>
  onDeleteDept: (dept: Department) => void | Promise<void>
  readOnly?: boolean
}

export function DepartmentTree({
  departments,
  selectedId,
  onSelect,
  onAddChildDept,
  onUpdateDeptName,
  onDeleteDept,
  readOnly = false,
}: DepartmentTreeProps) {
  const [search, setSearch] = useState('')
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [editId, setEditId] = useState<string | null>(null)
  const [editName, setEditName] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<Department | null>(null)
  const [deleteError, setDeleteError] = useState('')

  const toggleExpand = (id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const filteredTree = useMemo(
    () => filterDepartmentTree(departments, search),
    [departments, search],
  )

  const handleCommitEdit = async (id: string) => {
    if (!editName.trim()) return
    await onUpdateDeptName(id, editName.trim())
    setEditId(null)
    setEditName('')
  }

  const handleDelete = (dept: Department) => {
    const error = getDeptDeleteError(dept)
    if (error) {
      setDeleteError(error)
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
    await onDeleteDept(deleteTarget)
    setDeleteTarget(null)
  }

  const rootDept = departments[0]
  const nodeProps = {
    search,
    expanded,
    selectedId,
    editId,
    editName,
    readOnly,
    onSelect,
    onToggleExpand: toggleExpand,
    onAddChildDept,
    onStartEdit: (id: string, name: string) => {
      setEditId(id)
      setEditName(name)
    },
    onEditNameChange: setEditName,
    onCommitEdit: handleCommitEdit,
    onCancelEdit: () => setEditId(null),
    onDelete: handleDelete,
  }

  return (
    <TooltipProvider>
      <div className="flex h-full w-[220px] shrink-0 flex-col rounded-lg border border-border/50 bg-card shadow-card">
        <div className="p-3 border-b flex items-center gap-2">
          <Input
            type="text"
            placeholder="搜索部门"
            className="flex-1 h-7 text-sm"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          {!readOnly && (
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    size="xs"
                    onClick={() => {
                      if (rootDept) onAddChildDept(rootDept.id, rootDept.name)
                    }}
                  />
                }
              >
                <Plus className="h-3.5 w-3.5" />
              </TooltipTrigger>
              <TooltipContent>添加子部门</TooltipContent>
            </Tooltip>
          )}
        </div>
        <div className="flex-1 overflow-y-auto p-2">
          {filteredTree.map((dept) => (
            <DepartmentTreeNode key={dept.id} dept={dept} level={0} {...nodeProps} />
          ))}
        </div>

        <AlertDialog
          open={!!deleteTarget}
          onOpenChange={(isOpen) => {
            if (!isOpen) {
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
                  className="bg-destructive text-destructive-foreground hover:bg-destructive/80"
                >
                  删除
                </AlertDialogAction>
              )}
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </TooltipProvider>
  )
}
