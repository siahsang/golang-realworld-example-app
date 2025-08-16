package utils

func Map[T any, R any](items []T, f func(T) R) []R {
	result := make([]R, len(items))
	for i, v := range items {
		result[i] = f(v)
	}
	return result
}

func GetOrDefault[K comparable, T any](m map[K]T, key K, defaultValue T) T {
	v, ok := m[key]
	if !ok {
		return defaultValue
	} else {
		return v
	}
}
