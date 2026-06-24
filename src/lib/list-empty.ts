import type { DataSectionEmptyProps } from '@/components/layout/data-section'

export function listEmpty<T>(
  loading: boolean,
  items: readonly T[],
  config: DataSectionEmptyProps,
): DataSectionEmptyProps | null {
  return !loading && items.length === 0 ? config : null
}
