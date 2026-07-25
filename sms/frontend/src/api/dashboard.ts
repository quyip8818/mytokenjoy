import { request } from './client'

export interface DashboardCards {
  supplierTotal: number
  activeSuppliers: number
  modelTotal: number
  activeContracts: number
}

export interface LabelCount {
  label: string
  count: number
}

export interface DashboardCharts {
  gradeDistribution: LabelCount[]
  modelCountBySupplier: LabelCount[]
}

export interface ExpiringContract {
  id: string
  title: string
  contractNo: string
  endDate: string
  supplierName: string
}

export const dashboardApi = {
  cards: () => request<DashboardCards>('/dashboard/cards'),
  charts: () => request<DashboardCharts>('/dashboard/charts'),
  expiring: () => request<ExpiringContract[]>('/dashboard/expiring'),
  recentOrders: () => request<import('./orders').PurchaseOrder[]>('/dashboard/recent-orders'),
}
