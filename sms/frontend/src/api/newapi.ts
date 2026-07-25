import { request } from './client'

export interface SyncStatus {
  modelId: string
  modelName: string
  localInput: number | null
  localOutput: number | null
  remoteInput?: number | null
  remoteOutput?: number | null
  status: 'synced' | 'diverged' | 'missing' | 'skipped'
}

export interface NewAPIModel {
  modelId: string
  inputPrice: number
  outputPrice: number
  modelRatio: number
  completionRatio: number
}

export interface SyncResult {
  synced: number
}

export const newapiApi = {
  status: () => request<SyncStatus[]>('/newapi/status'),
  models: () => request<NewAPIModel[]>('/newapi/models'),
  sync: () => request<SyncResult>('/newapi/sync', { method: 'POST' }),
}
