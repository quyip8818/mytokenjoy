import { useState } from 'react'
import type { BudgetProjectView, Member } from '@/api/types'
import { BudgetProjectHeader, BudgetProjectSummary } from './budget-project-summary'
import { BudgetProjectDeleteAction } from './budget-project-delete-action'
import { BudgetProjectMembersSection } from './budget-project-members-section'
import { BudgetProjectSettingsForm } from './budget-project-settings-form'

interface BudgetDetailProjectProps {
  project: BudgetProjectView
  members: Member[]
  departmentMembers: Member[]
  membersLoading?: boolean
  onUpdateGroup: (groupId: string, data: { budget?: number; memberIds?: string[] }) => Promise<void>
  onDeleteGroup: (groupId: string) => Promise<void>
  onUpdated: () => void
  onDeleted: () => void
}

export function BudgetDetailProject({
  project,
  members,
  departmentMembers,
  membersLoading = false,
  onUpdateGroup,
  onDeleteGroup,
  onUpdated,
  onDeleted,
}: BudgetDetailProjectProps) {
  const [deleting, setDeleting] = useState(false)

  async function handleDelete() {
    setDeleting(true)
    try {
      await onDeleteGroup(project.id)
      onDeleted()
    } catch {
      setDeleting(false)
    }
  }

  return (
    <div className="flex flex-1 flex-col gap-6 overflow-y-auto p-5">
      <div className="flex items-center justify-between gap-3">
        <BudgetProjectHeader project={project} />
        <BudgetProjectDeleteAction
          projectName={project.name}
          deleting={deleting}
          onDelete={handleDelete}
        />
      </div>

      <BudgetProjectSummary project={project} />

      <BudgetProjectMembersSection
        project={project}
        members={members}
        departmentMembers={departmentMembers}
        membersLoading={membersLoading}
        onUpdateGroup={onUpdateGroup}
        onUpdated={onUpdated}
      />

      <BudgetProjectSettingsForm
        project={project}
        onUpdateGroup={onUpdateGroup}
        onUpdated={onUpdated}
      />
    </div>
  )
}
