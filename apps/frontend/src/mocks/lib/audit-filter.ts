export function filterByDateRange<T extends { createdAt: string }>(
  items: T[],
  from?: string,
  to?: string,
): T[] {
  if (!from && !to) return items
  return items.filter((item) => {
    const day = item.createdAt.slice(0, 10)
    if (from && day < from) return false
    if (to && day > to) return false
    return true
  })
}

export function filterByKeyword<T>(
  items: T[],
  keyword: string | undefined,
  fields: (keyof T)[],
): T[] {
  if (!keyword?.trim()) return items
  const q = keyword.trim().toLowerCase()
  return items.filter((item) =>
    fields.some((field) => String(item[field] ?? '').toLowerCase().includes(q)),
  )
}
