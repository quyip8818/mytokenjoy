import { orgHandlers } from './org'
import { budgetHandlers } from './budget'
import { keysHandlers } from './keys'
import { modelsHandlers } from './models'
import { dashboardHandlers } from './dashboard'
import { auditHandlers } from './audit'
import { sessionHandlers } from './session'
import { fallbackHandlers } from './fallback'

export const domainHandlers = [
  ...sessionHandlers,
  ...orgHandlers,
  ...budgetHandlers,
  ...keysHandlers,
  ...modelsHandlers,
  ...dashboardHandlers,
  ...auditHandlers,
]

export const browserHandlers = [...domainHandlers, ...fallbackHandlers]

export const serverHandlers = domainHandlers
