import { useCallback, useMemo, useState } from 'react'
import { NavLink, useLocation } from 'react-router'
import { ChevronDown, PanelLeftClose, PanelLeftOpen } from 'lucide-react'
import { cn } from '@/lib/utils'
import { getVisibleNavGroups, type NavGroup, type NavItem } from '@/config/nav'
import { useApprovalPendingCountQuery } from '@/features/approval'
import { usePermissions } from '@/features/session'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import {
  SIDEBAR_COLLAPSED_WIDTH_CLASS,
  SIDEBAR_EXPANDED_WIDTH_CLASS,
} from './sidebar-layout-constants'
import { useSidebarLayout } from './use-sidebar-layout'

// ─── Persistence ───

const STORAGE_KEY = 'nav-collapsed-groups'

function loadCollapsedGroups(): Set<string> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return new Set(JSON.parse(raw) as string[])
  } catch {
    // ignore
  }
  return new Set()
}

function saveCollapsedGroups(groups: Set<string>) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify([...groups]))
}

// ─── Hook ───

function useGroupCollapse(navGroups: NavGroup[], currentPath: string) {
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(() => {
    const persisted = loadCollapsedGroups()
    const base =
      persisted.size === 0
        ? new Set(navGroups.filter((g) => g.collapsed).map((g) => g.group))
        : new Set(persisted)
    // Ensure the group containing the current page is expanded on initial load
    for (const group of navGroups) {
      if (group.items.some((item) => currentPath === item.path)) {
        base.delete(group.group)
      }
    }
    return base
  })

  const toggle = useCallback((groupName: string) => {
    setCollapsedGroups((prev) => {
      const next = new Set(prev)
      if (next.has(groupName)) {
        next.delete(groupName)
      } else {
        next.add(groupName)
      }
      saveCollapsedGroups(next)
      return next
    })
  }, [])

  const isGroupCollapsed = useCallback(
    (group: NavGroup) => collapsedGroups.has(group.group),
    [collapsedGroups],
  )

  return { isGroupCollapsed, toggle }
}

// ─── Nav Item ───

interface SidebarNavItemProps {
  item: NavItem
  sidebarCollapsed: boolean
  badge: number
}

function SidebarNavItem({ item, sidebarCollapsed, badge }: SidebarNavItemProps) {
  const location = useLocation()
  const isActive = location.pathname === item.path
  const Icon = item.icon

  if (sidebarCollapsed) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <NavLink
            to={item.path}
            aria-label={item.label}
            className={cn(
              'group/nav relative flex items-center justify-center rounded-lg p-1.5 transition-all duration-150',
              isActive
                ? 'bg-primary/8 text-primary'
                : 'text-muted-foreground hover:scale-105 hover:bg-accent hover:text-accent-foreground',
            )}
          >
            <Icon className="size-[18px]" strokeWidth={1.75} />
            {badge > 0 && (
              <span className="absolute top-0.5 right-0.5 size-2 rounded-full bg-primary ring-2 ring-sidebar" />
            )}
          </NavLink>
        </TooltipTrigger>
        <TooltipContent side="right" sideOffset={8}>
          {item.label}
          {badge > 0 ? ` (${badge})` : ''}
        </TooltipContent>
      </Tooltip>
    )
  }

  return (
    <NavLink
      to={item.path}
      className={cn(
        'group/nav relative flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-all duration-150',
        isActive
          ? 'bg-primary/8 font-medium text-primary'
          : 'text-muted-foreground hover:scale-[1.01] hover:bg-accent/50 hover:text-foreground',
      )}
    >
      <Icon
        className={cn('size-4 shrink-0', isActive ? 'text-primary' : 'text-muted-foreground/70')}
        strokeWidth={isActive ? 1.75 : 1.5}
      />
      <span className="flex-1 truncate">{item.label}</span>
      {badge > 0 && (
        <span className="inline-flex min-w-[18px] items-center justify-center rounded-full bg-primary px-1 py-px text-[10px] font-semibold leading-none text-primary-foreground">
          {badge}
        </span>
      )}
    </NavLink>
  )
}

// ─── Nav Group ───

interface SidebarGroupProps {
  group: NavGroup
  groupCollapsed: boolean
  sidebarCollapsed: boolean
  onToggleGroup: () => void
  getBadge: (badgeKey?: string) => number
}

function SidebarGroup({
  group,
  groupCollapsed,
  sidebarCollapsed,
  onToggleGroup,
  getBadge,
}: SidebarGroupProps) {
  if (sidebarCollapsed) {
    return (
      <div className="space-y-0.5">
        {group.items.map((item) => (
          <SidebarNavItem
            key={item.path}
            item={item}
            sidebarCollapsed
            badge={getBadge(item.badgeKey)}
          />
        ))}
      </div>
    )
  }

  return (
    <div>
      <button
        type="button"
        onClick={onToggleGroup}
        className="group/header mb-1 flex w-full items-center gap-1.5 rounded-md px-3 py-1.5 text-base font-semibold text-muted-foreground/80 transition-colors hover:text-foreground"
        aria-expanded={!groupCollapsed}
      >
        <ChevronDown
          className={cn(
            'size-3.5 shrink-0 text-muted-foreground/50 transition-transform duration-150 group-hover/header:text-muted-foreground',
            groupCollapsed && '-rotate-90',
          )}
        />
        <span className="truncate">{group.group}</span>
      </button>
      {!groupCollapsed && (
        <div className="ml-2 space-y-px">
          {group.items.map((item) => (
            <SidebarNavItem
              key={item.path}
              item={item}
              sidebarCollapsed={false}
              badge={getBadge(item.badgeKey)}
            />
          ))}
        </div>
      )}
    </div>
  )
}

// ─── Header ───

interface SidebarHeaderProps {
  collapsed: boolean
  onToggle: () => void
}

function SidebarHeader({ collapsed, onToggle }: SidebarHeaderProps) {
  const toggleLabel = collapsed ? '展开侧边栏' : '收起侧边栏'

  const toggleButton = (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          className="shrink-0 text-muted-foreground/50 transition-colors hover:bg-accent hover:text-foreground"
          onClick={onToggle}
          aria-expanded={!collapsed}
          aria-label={toggleLabel}
        >
          {collapsed ? <PanelLeftOpen className="size-4" /> : <PanelLeftClose className="size-4" />}
        </Button>
      </TooltipTrigger>
      <TooltipContent side={collapsed ? 'right' : 'bottom'} sideOffset={8}>
        {toggleLabel}
      </TooltipContent>
    </Tooltip>
  )

  if (collapsed) {
    return (
      <div className="flex shrink-0 justify-center border-b border-border/50 px-2 py-3">
        {toggleButton}
      </div>
    )
  }

  return (
    <div className="flex shrink-0 items-center justify-between border-b border-border/50 px-4 py-4">
      <div className="min-w-0 flex-1">
        <img src="/logo.png" alt="Tokenjoy" className="h-6 w-auto" />
      </div>
      {toggleButton}
    </div>
  )
}

// ─── Main ───

export function Sidebar() {
  const { permissions } = usePermissions()
  const navGroups = useMemo(() => getVisibleNavGroups(permissions), [permissions])
  const { data: approvalPendingCount = 0 } = useApprovalPendingCountQuery({ poll: true })
  const { collapsed, toggleCollapsed } = useSidebarLayout()
  const location = useLocation()
  const { isGroupCollapsed, toggle } = useGroupCollapse(navGroups, location.pathname)

  const getBadge = (badgeKey?: string) => {
    if (badgeKey === 'approvalPending' && approvalPendingCount > 0) {
      return approvalPendingCount
    }
    return 0
  }

  // All groups rendered uniformly (including "我的")
  return (
    <TooltipProvider delayDuration={0}>
      <aside
        className={cn(
          'group/sidebar relative flex shrink-0 flex-col overflow-hidden border-r border-border/60 bg-sidebar transition-[width] duration-200 ease-in-out',
          collapsed ? SIDEBAR_COLLAPSED_WIDTH_CLASS : SIDEBAR_EXPANDED_WIDTH_CLASS,
        )}
      >
        <SidebarHeader collapsed={collapsed} onToggle={toggleCollapsed} />

        <nav
          className={cn(
            'flex-1 overflow-y-auto py-3',
            collapsed ? 'space-y-2 px-1.5' : 'space-y-1 px-2',
          )}
        >
          {navGroups.map((group, groupIndex) => (
            <div key={group.group}>
              {collapsed && groupIndex > 0 && (
                <div className="mx-auto mb-2 h-px w-4 rounded-full bg-border/60" />
              )}
              <SidebarGroup
                group={group}
                groupCollapsed={isGroupCollapsed(group)}
                sidebarCollapsed={collapsed}
                onToggleGroup={() => toggle(group.group)}
                getBadge={getBadge}
              />
            </div>
          ))}
        </nav>
      </aside>
    </TooltipProvider>
  )
}
