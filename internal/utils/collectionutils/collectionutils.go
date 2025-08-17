package collectionutils

func Associate[T any, K comparable, V any](items []T, transform func(T) (K, V)) map[K]V {
	m := make(map[K]V, len(items))
	for _, item := range items {
		k, v := transform(item)
		m[k] = v
	}

	return m
}

func GetOrDefault[K comparable, T any](m map[K]T, key K, defaultValue T) T {
	v, ok := m[key]
	if !ok {
		return defaultValue
	} else {
		return v
	}
}
