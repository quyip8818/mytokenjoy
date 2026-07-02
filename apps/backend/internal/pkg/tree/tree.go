package tree

func Flatten[T any, C ~[]T](nodes []T, children func(T) C, clearChildren func(*T)) []T {
	result := make([]T, 0)
	for _, node := range nodes {
		cloned := node
		if clearChildren != nil {
			clearChildren(&cloned)
		}
		result = append(result, cloned)
		childNodes := children(node)
		if len(childNodes) > 0 {
			result = append(result, Flatten(childNodes, children, clearChildren)...)
		}
	}
	return result
}
