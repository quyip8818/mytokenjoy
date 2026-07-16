import { useCallback, useMemo, useState } from 'react'
import type { AppApis } from '@/api/app-apis'
import type { BudgetNode, Project, ProjectView } from '@/api/types'
import type { Member } from '@/api/types'
import { findBudgetNode, projectsForDepartment } from '../lib/mappers'
import { filterProjectMembers, useBudgetDepartmentMembers } from './use-budget-department-members'

interface UseBudgetSelectionOptions {
  injectedApis?: AppApis
  tree: BudgetNode[]
  projects: ProjectView[]
  projectsData: Project[]
}

export function useBudgetSelection({
  injectedApis,
  tree,
  projects,
  projectsData,
}: UseBudgetSelectionOptions) {
  const [selectedTeamId, setSelectedTeamId] = useState<string | undefined>()
  const [activeProjectId, setActiveProjectId] = useState<string | null>(null)
  const [approvalOpen, setApprovalOpen] = useState(false)

  const resolvedSelectedTeamId = selectedTeamId ?? tree[0]?.id

  const selectedNode = useMemo(
    () => (resolvedSelectedTeamId ? findBudgetNode(tree, resolvedSelectedTeamId) : null),
    [tree, resolvedSelectedTeamId],
  )

  const activeProject = useMemo(
    () =>
      activeProjectId ? projects.find((project) => project.id === activeProjectId) : undefined,
    [projects, activeProjectId],
  )

  const departmentIdForMembers = activeProject?.departmentId ?? selectedNode?.id
  const { departmentMembers, departmentMembersLoading } = useBudgetDepartmentMembers({
    injectedApis,
    departmentId: departmentIdForMembers,
  })

  const projectMembers: Member[] = useMemo(
    () => (activeProject ? filterProjectMembers(departmentMembers, activeProject.memberIds) : []),
    [activeProject, departmentMembers],
  )

  const handleSelectTeam = useCallback((nodeId: string) => {
    setSelectedTeamId(nodeId)
    setActiveProjectId(null)
  }, [])

  const projectsForNode = useCallback(
    (departmentId: string) => projectsForDepartment(projectsData, departmentId),
    [projectsData],
  )

  return {
    selectedTeamId: resolvedSelectedTeamId,
    selectedNode,
    activeProjectId,
    activeProject,
    approvalOpen,
    setApprovalOpen,
    departmentMembers,
    departmentMembersLoading,
    projectMembers,
    handleSelectTeam,
    setActiveProjectId,
    projectsForNode,
  }
}
