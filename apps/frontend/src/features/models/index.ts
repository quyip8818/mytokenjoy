// === 页面入口（route page 消费）===
export { modelsKeys } from './query-keys'
export { useModelListPage } from './hooks/use-model-list-page'
export { useModelRoutingPage } from './hooks/use-model-routing-page'
export { ModelListPageShell } from './components/model-list-page-shell'
export { ModelRoutingPageShell } from './components/model-routing-page-shell'

// === 跨 feature 共享 ===
// consumed by: workflow (approval-review, key-form, approval-submit), keys (platform-keys-page-shell)
export { useModelLabels } from './hooks/use-model-labels'
// consumed by: workflow/whitelist-config, self components
export { modelRefLabel } from './lib/model-catalog'
// consumed by: workflow/model-picker, workflow/whitelist-config
export { isBuiltinModel } from './lib/model-kind'
// consumed by: workflow/model-edit
export { isCustomModel } from './lib/model-kind'
// consumed by: workflow/key-form, workflow/approval-submit — shared inline model multi-select
export { InlineModelPicker } from './components/inline-model-picker'
// consumed by: platform/models (shared model table component)
export { ModelTable } from './components/model-table'
