import { authApi } from './auth'
import { suppliersApi } from './suppliers'
import { contractsApi } from './contracts'
import { ordersApi } from './orders'
import { evaluationsApi } from './evaluations'
import { dashboardApi } from './dashboard'
import { usersApi } from './users'

export interface AppApis {
  authApi: typeof authApi
  suppliersApi: typeof suppliersApi
  contractsApi: typeof contractsApi
  ordersApi: typeof ordersApi
  evaluationsApi: typeof evaluationsApi
  dashboardApi: typeof dashboardApi
  usersApi: typeof usersApi
}

export const defaultApis: AppApis = {
  authApi,
  suppliersApi,
  contractsApi,
  ordersApi,
  evaluationsApi,
  dashboardApi,
  usersApi,
}
