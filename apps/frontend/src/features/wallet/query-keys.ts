export const walletKeys = {
  all: ['wallet'] as const,
  wallet: () => [...walletKeys.all, 'wallet'] as const,
  rechargeRecords: () => [...walletKeys.all, 'recharge-records'] as const,
}
