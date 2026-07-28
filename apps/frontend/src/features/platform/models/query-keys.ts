export const platformKeys = {
  all: ['platform'] as const,
  models: () => [...platformKeys.all, 'models'] as const,
}
