import { approvalApi } from './approval'
import { auditApi } from './audit'
import { authApi } from './auth'
import { billingApi } from './billing'
import { budgetApi } from './budget'
import { dashboardApi } from './dashboard'
import { keysApi } from './keys'
import { meApi } from './me'
import { modelsApi } from './models'
import { notificationApi } from './notifications'
import { orgApi } from './org'
import { platformApi } from './platform'
import { sessionApi } from './session'

export interface AppApis {
  authApi: typeof authApi
  billingApi: typeof billingApi
  budgetApi: typeof budgetApi
  auditApi: typeof auditApi
  dashboardApi: typeof dashboardApi
  modelsApi: typeof modelsApi
  orgApi: typeof orgApi
  keysApi: typeof keysApi
  approvalApi: typeof approvalApi
  meApi: typeof meApi
  notificationApi: typeof notificationApi
  sessionApi: typeof sessionApi
  platformApi: typeof platformApi
}

export const defaultApis: AppApis = {
  authApi,
  billingApi,
  budgetApi,
  auditApi,
  dashboardApi,
  modelsApi,
  orgApi,
  keysApi,
  approvalApi,
  meApi,
  notificationApi,
  sessionApi,
  platformApi,
}
