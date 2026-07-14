import { useState } from 'react'
import type { PlatformKeyScope, ProjectView, Member } from '@/api/types'
import { ProjectHeader, ProjectSummary } from './project-summary'
import { ProjectDeleteAction } from './project-delete-action'
import { ProjectMembersSection } from './project-members-section'
import { ProjectSettingsForm } from './project-settings-form'
import { Button } from '@/components/ui/button'
import { Plus } from 'lucide-react'

interface ProjectDetailProps {
  project: ProjectView
  members: Member[]
  departmentMembers: Member[]
  membersLoading?: boolean
  onUpdateProject: (
    projectId: string,
    data: { budget?: number; memberIds?: string[]; memberBudgets?: Record<string, number> },
  ) => Promise<void>
  onDeleteProject: (projectId: string) => Promise<void>
  onCreateProjectKey: (
    project: ProjectView,
    scope: PlatformKeyScope,
    memberId?: string,
    memberName?: string,
  ) => void
  onUpdated: () => void
  onDeleted: () => void
}

export function ProjectDetail({
  project,
  members,
  departmentMembers,
  membersLoading = false,
  onUpdateProject,
  onDeleteProject,
  onCreateProjectKey,
  onUpdated,
  onDeleted,
}: ProjectDetailProps) {
  const [deleting, setDeleting] = useState(false)

  async function handleDelete() {
    setDeleting(true)
    try {
      await onDeleteProject(project.id)
      onDeleted()
    } catch {
      setDeleting(false)
    }
  }

  return (
    <div className="flex flex-1 flex-col gap-6 overflow-y-auto p-5">
      <div className="flex items-center justify-between gap-3">
        <ProjectHeader project={project} />
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            className="h-8 gap-1.5"
            onClick={() => onCreateProjectKey(project, 'project')}
          >
            <Plus className="size-3.5" />
            签发项目 Key
          </Button>
          <ProjectDeleteAction
            projectName={project.name}
            deleting={deleting}
            onDelete={handleDelete}
          />
        </div>
      </div>

      <ProjectSummary project={project} />

      <ProjectMembersSection
        project={project}
        members={members}
        departmentMembers={departmentMembers}
        membersLoading={membersLoading}
        onUpdateProject={onUpdateProject}
        onUpdated={onUpdated}
      />

      <ProjectSettingsForm
        project={project}
        onUpdateProject={onUpdateProject}
        onUpdated={onUpdated}
      />
    </div>
  )
}
