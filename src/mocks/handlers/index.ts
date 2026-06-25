import { orgHandlers } from './org'
import { budgetHandlers } from './budget'
import { keysHandlers } from './keys'
import { modelsHandlers } from './models'
import { dashboardHandlers } from './dashboard'
import { auditHandlers } from './audit'
import { sessionHandlers } from './session'

export const handlers = [
  ...sessionHandlers,
  ...orgHandlers,
  ...budgetHandlers,
  ...keysHandlers,
  ...modelsHandlers,
  ...dashboardHandlers,
  ...auditHandlers,
]
