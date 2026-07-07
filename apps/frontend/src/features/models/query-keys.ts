export const modelsKeys = {
  all: ['models'] as const,
  list: () => [...modelsKeys.all, 'list'] as const,
  routing: () => [...modelsKeys.all, 'routing'] as const,
}
