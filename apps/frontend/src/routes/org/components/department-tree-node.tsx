import { Plus, Pencil, X } from 'lucide-react'
import type { Department } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'

function highlightText(text: string, keyword: string) {
  if (!keyword) return text
  const idx = text.indexOf(keyword)
  if (idx === -1) return text
  return (
    <>
      {text.slice(0, idx)}
      <span className="rounded-sm bg-amber-200/80 px-0.5">{keyword}</span>
      {text.slice(idx + keyword.length)}
    </>
  )
}

export interface DepartmentTreeNodeProps {
  dept: Department
  level: number
  search: string
  expanded: Set<string>
  selectedId: string | undefined
  editId: string | null
  editName: string
  readOnly: boolean
  onSelect: (dept: Department) => void
  onToggleExpand: (id: string) => void
  onAddChildDept: (parentId: string, parentName: string) => void
  onStartEdit: (id: string, name: string) => void
  onEditNameChange: (name: string) => void
  onCommitEdit: (id: string) => void
  onCancelEdit: () => void
  onDelete: (dept: Department) => void
}

export function DepartmentTreeNode({
  dept,
  level,
  search,
  expanded,
  selectedId,
  editId,
  editName,
  readOnly,
  onSelect,
  onToggleExpand,
  onAddChildDept,
  onStartEdit,
  onEditNameChange,
  onCommitEdit,
  onCancelEdit,
  onDelete,
}: DepartmentTreeNodeProps) {
  const isExpanded = expanded.has(dept.id)
  const hasChildren = dept.children && dept.children.length > 0
  const isSelected = selectedId === dept.id
  const isEditing = editId === dept.id

  return (
    <div>
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
          onClick={(e) => {
            e.stopPropagation()
            onToggleExpand(dept.id)
          }}
        >
          {hasChildren ? (isExpanded ? '▾' : '▸') : ''}
        </button>
        {isEditing ? (
          <Input
            className="flex-1 h-6 text-sm"
            value={editName}
            autoFocus
            onChange={(e) => onEditNameChange(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') void onCommitEdit(dept.id)
              if (e.key === 'Escape') onCancelEdit()
            }}
            onBlur={() => void onCommitEdit(dept.id)}
            onClick={(e) => e.stopPropagation()}
          />
        ) : (
          <span className="flex-1 truncate">{highlightText(dept.name, search)}</span>
        )}
        <span className="text-xs text-muted-foreground mr-1">{dept.memberCount}</span>
        {!readOnly && (
          <div
            className="hidden group-hover:flex items-center gap-0.5"
            onClick={(e) => e.stopPropagation()}
          >
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => onAddChildDept(dept.id, dept.name)}
                  />
                }
              >
                <Plus className="h-3.5 w-3.5" />
              </TooltipTrigger>
              <TooltipContent>添加子部门</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    onClick={() => onStartEdit(dept.id, dept.name)}
                  />
                }
              >
                <Pencil className="h-3.5 w-3.5" />
              </TooltipTrigger>
              <TooltipContent>编辑</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    className="hover:text-destructive"
                    onClick={() => onDelete(dept)}
                  />
                }
              >
                <X className="h-3.5 w-3.5" />
              </TooltipTrigger>
              <TooltipContent>删除</TooltipContent>
            </Tooltip>
          </div>
        )}
      </div>
      {hasChildren &&
        isExpanded &&
        dept.children!.map((child) => (
          <DepartmentTreeNode
            key={child.id}
            dept={child}
            level={level + 1}
            search={search}
            expanded={expanded}
            selectedId={selectedId}
            editId={editId}
            editName={editName}
            readOnly={readOnly}
            onSelect={onSelect}
            onToggleExpand={onToggleExpand}
            onAddChildDept={onAddChildDept}
            onStartEdit={onStartEdit}
            onEditNameChange={onEditNameChange}
            onCommitEdit={onCommitEdit}
            onCancelEdit={onCancelEdit}
            onDelete={onDelete}
          />
        ))}
    </div>
  )
}
