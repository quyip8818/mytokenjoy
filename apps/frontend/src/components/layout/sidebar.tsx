import { NavLink, useLocation } from 'react-router'
import type { LucideIcon } from 'lucide-react'
import { PanelLeftClose, PanelLeftOpen } from 'lucide-react'
import { cn } from '@/lib/utils'
import { getVisibleNavGroups, type NavItem } from '@/config/nav'
import { useApprovalPendingCount } from '@/hooks/use-approval-pending-count'
import { usePermissions } from '@/hooks/use-permissions'
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
  const isActive = location.pathname.startsWith(item.path)
  const Icon = item.icon

  const className = cn(
    'group/nav relative flex items-center rounded-lg transition-all duration-150',
    collapsed ? 'justify-center p-1' : 'gap-2.5 px-2 py-1.5',
    !collapsed &&
      isActive &&
      'bg-sidebar-accent font-medium text-sidebar-accent-foreground ring-1 ring-primary/5',
    !collapsed &&
      !isActive &&
      'text-sidebar-foreground hover:bg-sidebar-accent/70 hover:text-sidebar-accent-foreground',
    collapsed && !isActive && 'hover:bg-sidebar-accent/70',
  )

  const content = (
    <>
      <SidebarNavIcon icon={Icon} active={isActive} collapsed={collapsed} badge={badge} />
      {!collapsed && (
        <>
          <span className="flex-1 truncate">{item.label}</span>
          {badge > 0 && (
            <span className="inline-flex min-w-5 items-center justify-center rounded-full bg-primary px-1.5 py-0.5 text-xs font-semibold text-primary-foreground">
              {badge}
            </span>
          )}
        </>
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
      <TooltipTrigger
        render={<NavLink to={item.path} aria-label={item.label} className={className} />}
      >
        {content}
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
      <TooltipTrigger
        render={
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
          />
        }
      >
        {collapsed ? <PanelLeftOpen className="h-4 w-4" /> : <PanelLeftClose className="h-4 w-4" />}
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
        <h1 className="truncate text-xl font-extrabold tracking-tight text-gradient">TokenJoy</h1>
        <p className="mt-0.5 truncate text-[11px] text-muted-foreground">LLM API 管理平台</p>
      </div>
      <div className="pt-0.5">{toggleButton}</div>
    </div>
  )
}

export function Sidebar() {
  const location = useLocation()
  const { permissions } = usePermissions()
  const navGroups = getVisibleNavGroups(permissions)
  const approvalPendingCount = useApprovalPendingCount()
  const { collapsed, toggleCollapsed } = useSidebarLayout()

  const getBadge = (badgeKey?: string) => {
    if (badgeKey === 'approvalPending' && approvalPendingCount > 0) {
      return approvalPendingCount
    }
    return 0
  }

  return (
    <TooltipProvider delay={0}>
      <aside
        className={cn(
          'group/sidebar relative flex shrink-0 flex-col overflow-hidden border-r border-sidebar-border bg-sidebar transition-[width] duration-200 ease-in-out',
          collapsed ? SIDEBAR_COLLAPSED_WIDTH_CLASS : SIDEBAR_EXPANDED_WIDTH_CLASS,
        )}
        style={{ boxShadow: 'var(--shadow-sidebar)' }}
      >
        <div className="pointer-events-none absolute inset-0 bg-gradient-to-b from-primary/[0.02] via-transparent to-sky-500/[0.015]" />
        <div className="pointer-events-none absolute -bottom-24 -left-24 h-64 w-64 rounded-full bg-primary/4 blur-3xl" />
        <div className="pointer-events-none absolute -top-12 -right-12 h-40 w-40 rounded-full bg-sky-400/4 blur-3xl" />

        <SidebarHeader collapsed={collapsed} onToggle={toggleCollapsed} />

        <nav
          className={cn(
            'relative z-10 flex-1 space-y-5 overflow-y-auto py-3',
            collapsed ? 'px-1.5' : 'px-2.5',
          )}
        >
          {navGroups.map((group, groupIndex) => {
            const isGroupActive = group.items.some((item) =>
              location.pathname.startsWith(item.path),
            )
            return (
              <div key={group.group}>
                {!collapsed && (
                  <div
                    className={cn(
                      'mb-1.5 px-2.5 text-xs font-semibold uppercase tracking-wider',
                      isGroupActive ? 'text-primary' : 'text-muted-foreground',
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
