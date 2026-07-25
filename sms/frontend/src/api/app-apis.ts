import { authApi } from './auth'
import { suppliersApi } from './suppliers'
import { modelsApi } from './models'
import { contractsApi } from './contracts'
import { ordersApi } from './orders'
import { evaluationsApi } from './evaluations'
import { dashboardApi } from './dashboard'
import { usersApi } from './users'
import { newapiApi } from './newapi'

export interface AppApis {
  authApi: typeof authApi
  suppliersApi: typeof suppliersApi
  modelsApi: typeof modelsApi
  contractsApi: typeof contractsApi
  ordersApi: typeof ordersApi
  evaluationsApi: typeof evaluationsApi
  dashboardApi: typeof dashboardApi
  usersApi: typeof usersApi
  newapiApi: typeof newapiApi
}

export const defaultApis: AppApis = {
  authApi,
  suppliersApi,
  modelsApi,
  contractsApi,
  ordersApi,
  evaluationsApi,
  dashboardApi,
  usersApi,
  newapiApi,
}
