import { request, buildQuery } from './client'

export interface PurchaseOrder {
  id: string
  orderNo: string
  supplierId: string
  contractId?: string
  totalAmount?: number
  orderDate?: string
  status: string
  description?: string
  createdBy?: string
  createdAt: string
  updatedAt: string
  supplierName?: string
  contractNo?: string
  creatorName?: string
}

export interface OrderCreateInput {
  orderNo: string
  supplierId: string
  contractId?: string
  totalAmount?: number
  orderDate?: string
  status: string
  description?: string
}

export interface OrderUpdateInput {
  orderNo?: string
  supplierId?: string
  contractId?: string
  totalAmount?: number
  orderDate?: string
  status?: string
  description?: string
}

export const ordersApi = {
  list: (params: Record<string, unknown>) =>
    request<{ items: PurchaseOrder[]; total: number; page: number; pageSize: number }>(
      `/purchase-orders${buildQuery(params)}`,
    ),
  detail: (id: string) => request<PurchaseOrder>(`/purchase-orders/${id}`),
  create: (data: OrderCreateInput) =>
    request<PurchaseOrder>('/purchase-orders', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: OrderUpdateInput) =>
    request<PurchaseOrder>(`/purchase-orders/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => request<void>(`/purchase-orders/${id}`, { method: 'DELETE' }),
}
