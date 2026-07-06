import type { Platform } from '@/api/types'

export interface CredentialFormPayload {
  connected?: boolean
  currentPlatform?: Platform | null
  onSuccess?: () => void
}

export interface SyncConfigPayload {
  onTriggerSync?: () => void
  triggeringSync?: boolean
  onSuccess?: () => void
}
