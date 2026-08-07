export const platformCurrenciesKeys = {
  all: ['platform', 'currencies'] as const,
  list: () => [...platformCurrenciesKeys.all, 'list'] as const,
  history: (code: string) => [...platformCurrenciesKeys.all, 'history', code] as const,
}
