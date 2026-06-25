import { approvalKeysHandlers } from './approval'
import { platformKeysHandlers } from './platform'
import { providerKeysHandlers } from './provider'

export const keysHandlers = [
  ...providerKeysHandlers,
  ...platformKeysHandlers,
  ...approvalKeysHandlers,
]
