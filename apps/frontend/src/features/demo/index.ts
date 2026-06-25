export { DemoProvider } from './demo-provider'

export {
  DEFAULT_DEMO_MEMBER_ID,
  DEMO_SWITCHABLE_MEMBERS,
  getSwitchableMember,
  type DemoSwitchableMember,
} from './roles/constants'
export { DemoRoleProvider } from './roles/provider'
export { useDemoRole } from './roles/use-demo-role'
export { DemoRoleNavigationBridge } from './roles/navigation-bridge'
export { createDemoRoleStore, type DemoRoleStoreState } from './roles/store'

export {
  DEMO_CTA_IDS,
  DEMO_GUIDE_STEPS,
  DEMO_GUIDE_STORAGE_KEY,
  type DemoCtaKey,
  type DemoCtaId,
  type DemoGuideStep,
} from './guide/constants'
export { DemoGuideProvider } from './guide/provider'
export { useDemoGuide, useDemoGuideHighlight, useDemoCta } from './guide/use-demo-guide'
export { DemoGuidePanel } from './guide/demo-guide-panel'
export { createDemoGuideStore, type DemoGuideStoreState } from './guide/store'
export { useDemoApprovalPendingCount } from './nav/use-approval-pending-count'

export { DemoBadge } from './chrome/demo-banner'
export { DemoToolbar } from './chrome/demo-toolbar'
export { DesktopOnlyHint } from './chrome/desktop-only-hint'
