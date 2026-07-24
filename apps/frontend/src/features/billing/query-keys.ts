export const billingKeys = {
  all: ['billing'] as const,
  wallet: () => [...billingKeys.all, 'wallet'] as const,
  rechargeRecords: () => [...billingKeys.all, 'recharge-records'] as const,
}
