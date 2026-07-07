export const walletKeys = {
  all: ['billing'] as const,
  wallet: () => [...walletKeys.all, 'wallet'] as const,
  rechargeRecords: () => [...walletKeys.all, 'recharge-records'] as const,
}
