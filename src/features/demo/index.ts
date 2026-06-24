export { DemoProvider } from './demo-provider'

export type { DemoRole } from './roles/constants'
export {
  DEMO_ROLE_PROFILES,
  DEMO_ROLES,
  DEMO_ROLE_HOME,
  getDefaultHomePath,
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

export { useApprovalPendingCount } from './nav/use-approval-pending-count'

export { DemoBanner } from './chrome/demo-banner'
export { DemoToolbar } from './chrome/demo-toolbar'
export { DesktopOnlyHint } from './chrome/desktop-only-hint'
