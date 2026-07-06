import type { Department } from '@/api/types'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Separator } from '@/components/ui/separator'
import { ChevronRight, MoreHorizontal, FolderPlus, Pencil, Trash2, Info, Folder, FolderOpen, Users } from 'lucide-react'

interface DepartmentNodeProps {
  department: Department
  level: number
  selectedId: string | undefined
  expandedIds: Set<string>
  searchKeyword: string
  onSelect: (dept: Department) => void
  onToggle: (id: string) => void
  onAdd: (parentId: string) => void
  onEdit: (dept: Department) => void
  onDelete: (dept: Department) => void
}

export function DepartmentNode({
  department, level, selectedId, expandedIds, searchKeyword,
  onSelect, onToggle, onAdd, onEdit, onDelete,
}: DepartmentNodeProps) {
  const hasChildren = department.children && department.children.length > 0
  const isSelected = selectedId === department.id
  const isExpanded = expandedIds.has(department.id)

  const highlightText = (text: string) => {
    if (!searchKeyword) return text
    const idx = text.indexOf(searchKeyword)
    if (idx === -1) return text
    return (
      <>
        {text.slice(0, idx)}
        <mark className="bg-amber-100 text-inherit rounded-sm px-0.5">{searchKeyword}</mark>
        {text.slice(idx + searchKeyword.length)}
      </>
    )
  }

  return (
    <div>
      <div
        role="treeitem"
        tabIndex={0}
        aria-selected={isSelected}
        aria-expanded={hasChildren ? isExpanded : undefined}
        className={cn(
          'group relative flex items-center gap-2 rounded-md px-2 py-1.5 text-sm cursor-pointer',
          isSelected
            ? 'bg-primary/8 text-primary'
            : 'text-foreground hover:bg-muted'
        )}
        style={{ paddingLeft: `${level * 16 + 8}px` }}
        onClick={() => onSelect(department)}
        onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(department) } }}
      >
        {/* Expand toggle */}
        {hasChildren ? (
          <span
            role="button"
            tabIndex={-1}
            aria-label={isExpanded ? '收起' : '展开'}
            onClick={(e) => { e.stopPropagation(); onToggle(department.id) }}
            onKeyDown={(e) => { if (e.key === 'Enter') { e.stopPropagation(); onToggle(department.id) } }}
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

        {/* Icon */}
        {hasChildren ? (
          isExpanded
            ? <FolderOpen className="size-4 shrink-0 text-muted-foreground" />
            : <Folder className="size-4 shrink-0 text-muted-foreground" />
        ) : (
          <Users className="size-4 shrink-0 text-muted-foreground" />
        )}

        {/* Name */}
        <span className="flex-1 truncate font-medium">{highlightText(department.name)}</span>

        {/* Actions popover */}
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="ghost"
              size="icon-xs"
              aria-label="部门操作"
              className="opacity-0 group-hover:opacity-100 transition-opacity duration-100"
              onClick={(e) => e.stopPropagation()}
            >
              <MoreHorizontal className="size-3.5" />
            </Button>
          </PopoverTrigger>
          <PopoverContent align="start" side="right" sideOffset={4} className="w-40 p-1" onClick={(e) => e.stopPropagation()}>
            <button type="button" className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-muted" onClick={() => onSelect(department)}>
              <Info className="size-3.5 text-muted-foreground" />查看详情
            </button>
            <button type="button" className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-muted" onClick={() => onAdd(department.id)}>
              <FolderPlus className="size-3.5 text-muted-foreground" />添加子部门
            </button>
            <button type="button" className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-muted" onClick={() => onEdit(department)}>
              <Pencil className="size-3.5 text-muted-foreground" />编辑部门
            </button>
            <Separator className="my-1" />
            <button type="button" className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-destructive hover:bg-red-50" onClick={() => onDelete(department)}>
              <Trash2 className="size-3.5" />删除部门
            </button>
          </PopoverContent>
        </Popover>
      </div>

      {/* Children */}
      {hasChildren && isExpanded && (
        <div>
          {department.children!.map((child) => (
            <DepartmentNode
              key={child.id} department={child} level={level + 1}
              selectedId={selectedId} expandedIds={expandedIds} searchKeyword={searchKeyword}
              onSelect={onSelect} onToggle={onToggle} onAdd={onAdd} onEdit={onEdit} onDelete={onDelete}
            />
          ))}
        </div>
      )}
    </div>
  )
}
