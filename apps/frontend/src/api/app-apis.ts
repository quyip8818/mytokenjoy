import { auditApi } from './audit'
import { authApi } from './auth'
import { billingApi } from './billing'
import { budgetApi } from './budget'
import { dashboardApi } from './dashboard'
import { approvalApi, platformKeyApi, providerKeyApi } from './keys'
import { modelApi, routingApi } from './models'
import { dataSourceApi, departmentApi, memberApi, roleApi, syncApi } from './org'
import { sessionApi } from './session'

export interface AppApis {
  authApi: typeof authApi
  billingApi: typeof billingApi
  budgetApi: typeof budgetApi
  auditApi: typeof auditApi
  dashboardApi: typeof dashboardApi
  modelApi: typeof modelApi
  routingApi: typeof routingApi
  dataSourceApi: typeof dataSourceApi
  syncApi: typeof syncApi
  departmentApi: typeof departmentApi
  memberApi: typeof memberApi
  roleApi: typeof roleApi
  providerKeyApi: typeof providerKeyApi
  platformKeyApi: typeof platformKeyApi
  approvalApi: typeof approvalApi
  sessionApi: typeof sessionApi
}

export const defaultApis: AppApis = {
  authApi,
  billingApi,
  budgetApi,
  auditApi,
  dashboardApi,
  modelApi,
  routingApi,
  dataSourceApi,
  syncApi,
  departmentApi,
  memberApi,
  roleApi,
  providerKeyApi,
  platformKeyApi,
  approvalApi,
  sessionApi,
}
