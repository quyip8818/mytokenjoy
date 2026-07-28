export const platformCompaniesKeys = {
  all: ['platform', 'companies'] as const,
  overview: () => [...platformCompaniesKeys.all, 'overview'] as const,
}
