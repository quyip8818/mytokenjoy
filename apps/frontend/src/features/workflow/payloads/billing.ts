export interface BillingWorkflowPayloads {
  'lot-audit': {
    companyId: string
    companyName: string
    readonly: boolean
    onSuccess?: () => void
  }
}
