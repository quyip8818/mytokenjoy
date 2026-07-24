// 11 exports: 9 external + 2 self-barrel (components must import via barrel)
// === 页面入口（route page 消费）===
export { orgKeys } from './query-keys'
export { useStructurePage } from './hooks/use-structure-page'
export { useRolesPage } from './hooks/use-roles-page'
export { useDataSourcePage } from './hooks/use-data-source-page'
export { StructurePageShell } from './components/structure/structure-page-shell'
export { RolesPageShell } from './components/roles-page-shell'
export { DataSourcePageShell } from './components/data-source/data-source-page-shell'

// === 跨 feature 共享 ===
// consumed by: workflow/whitelist-config
export { findParentDeptId } from './lib/departments'
// consumed by: keys/use-platform-keys-page
export { filterDepartmentTree } from './lib/departments'

// === 自身 components 通过 self-barrel 消费 ===
export { flattenDepartments } from './lib/departments'
export { PLATFORM_LABELS } from './lib/labels'
