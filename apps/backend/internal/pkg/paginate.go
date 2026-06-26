package pkg

func Paginate[T any](items []T, page, pageSize int) (result []T, total, safePage, safeSize int) {
	safePage = page
	if safePage < 1 {
		safePage = 1
	}
	safeSize = pageSize
	if safeSize < 1 {
		safeSize = 1
	}
	total = len(items)
	start := (safePage - 1) * safeSize
	if start >= total {
		return []T{}, total, safePage, safeSize
	}
	end := start + safeSize
	if end > total {
		end = total
	}
	return items[start:end], total, safePage, safeSize
}
