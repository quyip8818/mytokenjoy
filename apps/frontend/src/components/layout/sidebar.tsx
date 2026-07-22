import { NavLink, useLocation } from 'react-router'
import type { LucideIcon } from 'lucide-react'
import { PanelLeftClose, PanelLeftOpen } from 'lucide-react'
import { cn } from '@/lib/utils'
import { getVisibleNavGroups, type NavItem } from '@/config/nav'
import { useApprovalPendingCountQuery } from '@/features/org'
import { usePermissions } from '@/features/session'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import {
  SIDEBAR_COLLAPSED_WIDTH_CLASS,
  SIDEBAR_EXPANDED_WIDTH_CLASS,
} from './sidebar-layout-constants'
import { useSidebarLayout } from './use-sidebar-layout'

interface SidebarNavIconProps {
  icon: LucideIcon
  active: boolean
  collapsed: boolean
  badge: number
}

function SidebarNavIcon({ icon: Icon, active, collapsed, badge }: SidebarNavIconProps) {
  return (
    <span
      className={cn(
        'relative flex shrink-0 items-center justify-center rounded-md transition-colors',
        collapsed ? 'size-9' : 'size-8',
        active
          ? 'bg-primary/10 text-primary'
          : 'text-muted-foreground group-hover/nav:text-sidebar-accent-foreground',
      )}
    >
      <Icon className="size-[18px]" strokeWidth={1.75} />
      {badge > 0 && collapsed && (
        <span className="absolute top-1 right-1 size-1.5 rounded-full bg-primary ring-2 ring-sidebar" />
      )}
    </span>
  )
}

interface SidebarNavItemProps {
  item: NavItem
  collapsed: boolean
  badge: number
}

function SidebarNavItem({ item, collapsed, badge }: SidebarNavItemProps) {
  const location = useLocation()
  const isActive = location.pathname === item.path
  const Icon = item.icon

  const className = cn(
    'group/nav relative flex items-center transition-colors duration-100',
    collapsed ? 'justify-center rounded-lg p-1' : 'gap-2.5 rounded-md px-3 py-2 text-sm',
    !collapsed && isActive && 'bg-muted font-medium text-foreground',
    !collapsed && !isActive && 'text-muted-foreground hover:bg-muted hover:text-foreground',
    collapsed && !isActive && 'hover:bg-sidebar-accent/70',
  )

  const content = collapsed ? (
    <SidebarNavIcon icon={Icon} active={isActive} collapsed={collapsed} badge={badge} />
  ) : (
    <>
      <Icon className="h-4 w-4 shrink-0" strokeWidth={1.5} />
      <span className="flex-1 truncate">{item.label}</span>
      {badge > 0 && (
        <span className="inline-flex min-w-5 items-center justify-center rounded-full bg-primary px-1.5 py-0.5 text-xs font-semibold text-primary-foreground">
          {badge}
        </span>
      )}
    </>
  )

  if (!collapsed) {
    return (
      <NavLink to={item.path} className={className}>
        {content}
      </NavLink>
    )
  }

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <NavLink to={item.path} aria-label={item.label} className={className}>
          {content}
        </NavLink>
      </TooltipTrigger>
      <TooltipContent side="right" sideOffset={8}>
        {item.label}
        {badge > 0 ? ` (${badge})` : ''}
      </TooltipContent>
    </Tooltip>
  )
}

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
          className={cn(
            'shrink-0 text-muted-foreground/70 transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
            'group-hover/sidebar:text-muted-foreground',
          )}
          onClick={onToggle}
          aria-expanded={!collapsed}
          aria-label={toggleLabel}
        >
          {collapsed ? (
            <PanelLeftOpen className="h-4 w-4" />
          ) : (
            <PanelLeftClose className="h-4 w-4" />
          )}
        </Button>
      </TooltipTrigger>
      <TooltipContent side={collapsed ? 'right' : 'bottom'} sideOffset={8}>
        {toggleLabel}
      </TooltipContent>
    </Tooltip>
  )

  if (collapsed) {
    return (
      <div className="relative z-10 flex shrink-0 justify-center border-b border-sidebar-border/70 px-2 py-3">
        {toggleButton}
      </div>
    )
  }

  return (
    <div className="relative z-10 flex shrink-0 items-start gap-1 border-b border-sidebar-border/70 px-4 py-5">
      <div className="min-w-0 flex-1">
        <img src="/logo.png" alt="Tokenjoy" className="h-7 w-auto" />
        <p className="mt-0.5 truncate text-xs text-muted-foreground">LLM API 管理平台</p>
      </div>
      <div className="pt-0.5">{toggleButton}</div>
    </div>
  )
}

export function Sidebar() {
  const { permissions } = usePermissions()
  const navGroups = getVisibleNavGroups(permissions)
  const { data: approvalPendingCount = 0 } = useApprovalPendingCountQuery({ poll: true })
  const { collapsed, toggleCollapsed } = useSidebarLayout()

  const getBadge = (badgeKey?: string) => {
    if (badgeKey === 'approvalPending' && approvalPendingCount > 0) {
      return approvalPendingCount
    }
    return 0
  }

  return (
    <TooltipProvider delayDuration={0}>
      <aside
        className={cn(
          'group/sidebar relative flex shrink-0 flex-col overflow-hidden border-r border-sidebar-border bg-sidebar transition-[width] duration-200 ease-in-out',
          collapsed ? SIDEBAR_COLLAPSED_WIDTH_CLASS : SIDEBAR_EXPANDED_WIDTH_CLASS,
        )}
        style={{ boxShadow: 'var(--shadow-sidebar)' }}
      >
        <SidebarHeader collapsed={collapsed} onToggle={toggleCollapsed} />

        <nav
          className={cn(
            'relative z-10 flex-1 space-y-5 overflow-y-auto py-3',
            collapsed ? 'px-1.5' : 'px-2.5',
          )}
        >
          {navGroups.map((group, groupIndex) => {
            return (
              <div key={group.group}>
                {!collapsed && (
                  <div
                    className={cn(
                      'mb-1.5 px-2 text-xs font-medium uppercase tracking-wide text-muted-foreground',
                      group.collapsed && 'text-muted-foreground/60',
                    )}
                  >
                    {group.group}
                  </div>
                )}
                {collapsed && groupIndex > 0 && (
                  <div className="mx-auto mb-2 h-px w-5 bg-sidebar-border" />
                )}
                <div className="space-y-0.5">
                  {group.items.map((item) => (
                    <SidebarNavItem
                      key={item.path}
                      item={item}
                      collapsed={collapsed}
                      badge={getBadge(item.badgeKey)}
                    />
                  ))}
                </div>
              </div>
            )
          })}
        </nav>
      </aside>
    </TooltipProvider>
  )
}
