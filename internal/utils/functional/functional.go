package functional

func Map[T any, R any](items []T, f func(T) R) []R {
	result := make([]R, len(items))
	for i, v := range items {
		result[i] = f(v)
	}
	return result
}
