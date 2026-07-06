import { useState, useEffect, useMemo } from 'react'
import type { Department } from '@/api/types'
import { departmentApi } from '@/api/org'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
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

interface DepartmentTreeProps {
  selectedId: string | undefined
  onSelect: (dept: Department | undefined) => void
  onTreeChange: () => void
}

export function DepartmentTree({ selectedId, onSelect, onTreeChange }: DepartmentTreeProps) {
  const [tree, setTree] = useState<Department[]>([])
  const [search, setSearch] = useState('')
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [addOpen, setAddOpen] = useState(false)
  const [addParentId, setAddParentId] = useState<string | null>(null)
  const [addName, setAddName] = useState('')
  const [editId, setEditId] = useState<string | null>(null)
  const [editName, setEditName] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<Department | null>(null)
  const [deleteError, setDeleteError] = useState('')

  const loadTree = async () => {
    const data = await departmentApi.getTree()
    setTree(data)
  }

  useEffect(() => {
    departmentApi.getTree().then(setTree)
  }, [])

  const toggleExpand = (id: string) => {
    setExpanded(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const filteredTree = useMemo(() => {
    function filter(departments: Department[], keyword: string): Department[] {
      if (!keyword) return departments
      return departments.reduce<Department[]>((acc, dept) => {
        const childMatches = dept.children ? filter(dept.children, keyword) : []
        if (dept.name.includes(keyword) || childMatches.length > 0) {
          acc.push({ ...dept, children: childMatches.length > 0 ? childMatches : dept.children })
        }
        return acc
      }, [])
    }
    return filter(tree, search)
  }, [tree, search])

  const handleAdd = async () => {
    if (!addName.trim() || !addParentId) return
    await departmentApi.create({ name: addName.trim(), parentId: addParentId })
    setAddName('')
    setAddParentId(null)
    setAddOpen(false)
    loadTree()
    onTreeChange()
  }

  const handleEdit = async (id: string) => {
    if (!editName.trim()) return
    await departmentApi.update(id, { name: editName.trim() })
    setEditId(null)
    setEditName('')
    loadTree()
    onTreeChange()
  }

  const handleDelete = async (dept: Department) => {
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
    await departmentApi.delete(deleteTarget.id)
    if (selectedId === deleteTarget.id) onSelect(undefined)
    setDeleteTarget(null)
    loadTree()
    onTreeChange()
  }

  const highlightText = (text: string, keyword: string) => {
    if (!keyword) return text
    const idx = text.indexOf(keyword)
    if (idx === -1) return text
    return (
      <>
        {text.slice(0, idx)}
        <span className="bg-yellow-200">{keyword}</span>
        {text.slice(idx + keyword.length)}
      </>
    )
  }

  const renderNode = (dept: Department, level: number) => {
    const isExpanded = expanded.has(dept.id)
    const hasChildren = dept.children && dept.children.length > 0
    const isSelected = selectedId === dept.id
    const isEditing = editId === dept.id

    return (
      <div key={dept.id}>
        <div
          className={`flex items-center gap-1 px-2 py-1.5 rounded cursor-pointer group text-sm hover:bg-muted ${
            isSelected ? 'bg-primary/10 text-primary' : 'text-foreground'
          }`}
          style={{ paddingLeft: `${level * 16 + 8}px` }}
          onClick={() => onSelect(dept)}
        >
          <button
            type="button"
            className="w-4 h-4 flex items-center justify-center shrink-0"
            onClick={(e) => { e.stopPropagation(); toggleExpand(dept.id) }}
          >
            {hasChildren ? (isExpanded ? '▾' : '▸') : ''}
          </button>
          {isEditing ? (
            <Input
              className="flex-1 h-6 text-sm"
              value={editName}
              autoFocus
              onChange={e => setEditName(e.target.value)}
              onKeyDown={e => { if (e.key === 'Enter') handleEdit(dept.id); if (e.key === 'Escape') setEditId(null) }}
              onBlur={() => handleEdit(dept.id)}
              onClick={e => e.stopPropagation()}
            />
          ) : (
            <span className="flex-1 truncate">{highlightText(dept.name, search)}</span>
          )}
          <span className="text-xs text-muted-foreground mr-1">{dept.memberCount}</span>
          <div className="hidden group-hover:flex items-center gap-0.5" onClick={e => e.stopPropagation()}>
            <Button
              variant="ghost"
              size="icon-xs"
              title="添加子部门"
              onClick={() => { setAddParentId(dept.id); setAddName(''); setAddOpen(true) }}
            >+</Button>
            <Button
              variant="ghost"
              size="icon-xs"
              title="编辑"
              onClick={() => { setEditId(dept.id); setEditName(dept.name) }}
            >✎</Button>
            <Button
              variant="ghost"
              size="icon-xs"
              title="删除"
              className="hover:text-destructive"
              onClick={() => handleDelete(dept)}
            >✕</Button>
          </div>
        </div>
        {hasChildren && isExpanded && dept.children!.map(child => renderNode(child, level + 1))}
      </div>
    )
  }

  return (
    <div className="w-[220px] shrink-0 border rounded-lg bg-card flex flex-col h-full">
      <div className="p-3 border-b flex items-center gap-2">
        <Input
          type="text"
          placeholder="搜索部门"
          className="flex-1 h-7 text-sm"
          value={search}
          onChange={e => setSearch(e.target.value)}
        />
        <Button
          size="xs"
          onClick={() => { setAddParentId(tree[0]?.id ?? ''); setAddName(''); setAddOpen(true) }}
        >+</Button>
      </div>
      <div className="flex-1 overflow-y-auto p-2">
        {filteredTree.map(dept => renderNode(dept, 0))}
      </div>

      {/* Add department dialog */}
      <Dialog open={addOpen} onOpenChange={(isOpen) => { if (!isOpen) { setAddOpen(false); setAddParentId(null) } }}>
        <DialogContent className="sm:max-w-xs">
          <DialogHeader>
            <DialogTitle>添加子部门</DialogTitle>
          </DialogHeader>
          <Input
            type="text"
            placeholder="部门名称"
            value={addName}
            onChange={e => setAddName(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') handleAdd() }}
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setAddOpen(false)}>取消</Button>
            <Button onClick={handleAdd}>确定</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirm dialog */}
      <AlertDialog open={!!deleteTarget} onOpenChange={(isOpen) => { if (!isOpen) { setDeleteTarget(null); setDeleteError('') } }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{deleteError ? '无法删除' : '删除部门'}</AlertDialogTitle>
            <AlertDialogDescription>
              {deleteError || `确定删除部门「${deleteTarget?.name ?? ''}」？`}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => { setDeleteTarget(null); setDeleteError('') }}>
              {deleteError ? '知道了' : '取消'}
            </AlertDialogCancel>
            {!deleteError && (
              <AlertDialogAction onClick={confirmDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/80">
                删除
              </AlertDialogAction>
            )}
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
