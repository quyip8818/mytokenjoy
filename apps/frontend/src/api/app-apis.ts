import { approvalApi } from './approval'
import { auditApi } from './audit'
import { authApi } from './auth'
import { billingApi } from './billing'
import { budgetApi } from './budget'
import { dashboardApi } from './dashboard'
import { devApi } from './dev'
import { platformKeyApi, providerKeyApi } from './keys'
import { meApi } from './me'
import { modelApi, routingApi } from './models'
import { notificationApi } from './notifications'
import { dataSourceApi, departmentApi, memberApi, roleApi, syncApi } from './org'
import { sessionApi } from './session'

export interface AppApis {
  authApi: typeof authApi
  billingApi: typeof billingApi
  budgetApi: typeof budgetApi
  auditApi: typeof auditApi
  dashboardApi: typeof dashboardApi
  devApi: typeof devApi
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
  meApi: typeof meApi
  notificationApi: typeof notificationApi
  sessionApi: typeof sessionApi
}

export const defaultApis: AppApis = {
  authApi,
  billingApi,
  budgetApi,
  auditApi,
  dashboardApi,
  devApi,
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
  meApi,
  notificationApi,
  sessionApi,
}
