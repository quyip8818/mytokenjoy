import { describe, expect, it, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { OrgTreeSidebar } from '@/features/dashboard/components/org-tree-sidebar'
import type { Department } from '@/api/types'

const tree: Department[] = [
  {
    id: 'd1',
    name: '工程部',
    parentId: null,
    memberCount: 10,
    children: [{ id: 'd1-1', name: '后端组', parentId: 'd1', memberCount: 5, children: [] }],
  },
  { id: 'd2', name: '产品部', parentId: null, memberCount: 8, children: [] },
]

describe('OrgTreeSidebar', () => {
  it('renders root node and top-level departments', () => {
    render(
      <OrgTreeSidebar
        departments={tree}
        selectedDeptId={null}
        onSelect={() => {}}
        loading={false}
      />,
    )
    expect(screen.getByText('全公司')).toBeInTheDocument()
    expect(screen.getByText('工程部')).toBeInTheDocument()
    expect(screen.getByText('产品部')).toBeInTheDocument()
  })

  it('highlights the selected node', () => {
    render(
      <OrgTreeSidebar departments={tree} selectedDeptId="d1" onSelect={() => {}} loading={false} />,
    )
    const node = screen.getByText('工程部').closest('[role="treeitem"]')
    expect(node).toHaveAttribute('aria-selected', 'true')
  })

  it('calls onSelect when a node is clicked', () => {
    const onSelect = vi.fn()
    render(
      <OrgTreeSidebar
        departments={tree}
        selectedDeptId={null}
        onSelect={onSelect}
        loading={false}
      />,
    )
    fireEvent.click(screen.getByText('产品部'))
    expect(onSelect).toHaveBeenCalledWith('d2')
  })

  it('calls onSelect with null when root is clicked', () => {
    const onSelect = vi.fn()
    render(
      <OrgTreeSidebar departments={tree} selectedDeptId="d1" onSelect={onSelect} loading={false} />,
    )
    fireEvent.click(screen.getByText('全公司'))
    expect(onSelect).toHaveBeenCalledWith(null)
  })
})
