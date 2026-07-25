import { request, buildQuery } from './client'
import type { Contract } from './contracts'
import type { PurchaseOrder } from './orders'
import type { Evaluation } from './evaluations'

export interface Supplier {
  id: string
  name: string
  code: string
  category?: string
  website?: string
  status: string
  description?: string
  createdBy?: string
  createdAt: string
  updatedAt: string
}

export interface SupplierContact {
  id: string
  supplierId: string
  name: string
  position?: string
  phone?: string
  email?: string
  isPrimary: boolean
  createdAt: string
}

export interface SupplierModel {
  id: string
  supplierId: string
  modelName: string
  modelType?: string
  contextLength?: number
  inputPrice?: number
  outputPrice?: number
  discount?: number
  status: string
  description?: string
  createdAt: string
  updatedAt: string
}

export interface SupplierDetail extends Supplier {
  contacts: SupplierContact[]
  models: SupplierModel[]
  contracts: Contract[]
  orders: PurchaseOrder[]
  evaluations: Evaluation[]
}

export interface SupplierCreateInput {
  name: string
  code: string
  category?: string
  website?: string
  status?: string
  description?: string
}

export interface SupplierUpdateInput {
  name?: string
  category?: string
  website?: string
  status?: string
  description?: string
}

export interface ContactCreateInput {
  name: string
  position?: string
  phone?: string
  email?: string
  isPrimary?: boolean
}

export interface ContactUpdateInput {
  name?: string
  position?: string
  phone?: string
  email?: string
  isPrimary?: boolean
}

export const suppliersApi = {
  list: (params: Record<string, unknown>) =>
    request<{ items: Supplier[]; total: number; page: number; pageSize: number }>(
      `/suppliers${buildQuery(params)}`,
    ),
  detail: (id: string) => request<SupplierDetail>(`/suppliers/${id}`),
  options: () => request<{ id: string; name: string }[]>('/suppliers/options'),
  create: (data: SupplierCreateInput) =>
    request<Supplier>('/suppliers', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: SupplierUpdateInput) =>
    request<Supplier>(`/suppliers/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => request<void>(`/suppliers/${id}`, { method: 'DELETE' }),
  createContact: (supplierId: string, data: ContactCreateInput) =>
    request<SupplierContact>(`/suppliers/${supplierId}/contacts`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateContact: (supplierId: string, contactId: string, data: ContactUpdateInput) =>
    request<void>(`/suppliers/${supplierId}/contacts/${contactId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  deleteContact: (supplierId: string, contactId: string) =>
    request<void>(`/suppliers/${supplierId}/contacts/${contactId}`, { method: 'DELETE' }),
}
