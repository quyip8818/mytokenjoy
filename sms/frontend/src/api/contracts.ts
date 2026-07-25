import { request, buildQuery } from './client'

export interface Contract {
  id: string
  supplierId: string
  contractNo: string
  title: string
  amount?: number
  signDate?: string
  startDate?: string
  endDate?: string
  status: string
  remarks?: string
  createdBy?: string
  createdAt: string
  updatedAt: string
  supplierName?: string
}

export interface ContractAttachment {
  id: string
  contractId: string
  fileName: string
  fileSize: number
  uploadedBy?: string
  createdAt: string
}

export interface ContractDetail extends Contract {
  attachments: ContractAttachment[]
}

export interface ContractCreateInput {
  contractNo: string
  supplierId: string
  title: string
  amount?: number
  signDate?: string
  startDate?: string
  endDate?: string
  status: string
  remarks?: string
}

export interface ContractUpdateInput {
  contractNo?: string
  supplierId?: string
  title?: string
  amount?: number
  signDate?: string
  startDate?: string
  endDate?: string
  status?: string
  remarks?: string
}

export const contractsApi = {
  list: (params: Record<string, unknown>) =>
    request<{ items: Contract[]; total: number; page: number; pageSize: number }>(
      `/contracts${buildQuery(params)}`,
    ),
  detail: (id: string) => request<ContractDetail>(`/contracts/${id}`),
  create: (data: ContractCreateInput) =>
    request<Contract>('/contracts', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: ContractUpdateInput) =>
    request<Contract>(`/contracts/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => request<void>(`/contracts/${id}`, { method: 'DELETE' }),
  uploadAttachment: (contractId: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    return request<ContractAttachment>(`/contracts/${contractId}/attachments`, {
      method: 'POST',
      body: form,
    })
  },
  deleteAttachment: (contractId: string, attachmentId: string) =>
    request<void>(`/contracts/${contractId}/attachments/${attachmentId}`, { method: 'DELETE' }),
}
