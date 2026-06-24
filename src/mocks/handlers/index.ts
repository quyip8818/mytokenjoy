import { orgHandlers } from './org'
import { budgetHandlers } from './budget'
import { keysHandlers } from './keys'
import { modelsHandlers } from './models'
import { dashboardHandlers } from './dashboard'
import { auditHandlers } from './audit'

export const handlers = [
  ...orgHandlers,
  ...budgetHandlers,
  ...keysHandlers,
  ...modelsHandlers,
  ...dashboardHandlers,
  ...auditHandlers,
]
