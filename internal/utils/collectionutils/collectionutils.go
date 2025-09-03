package collectionutils

// Associate transforms a slice of items into a map by applying the transform function to each item.
// The transform function returns a key-value pair for each item, which is then added to the resulting map.
func Associate[T any, K comparable, V any](items []T, transform func(T) (K, V)) map[K]V {
	m := make(map[K]V, len(items))
	for _, item := range items {
		k, v := transform(item)
		m[k] = v
	}

	return m
}

// GroupBy groups a slice of items into a map based on a key selector function.
// The key selector function extracts a key from each item, and the resulting map contains slices of items for each unique key.
func GroupBy[T any, K comparable](items []T, keySelector func(T) K) map[K][]T {
	m := make(map[K][]T)
	for _, item := range items {
		k := keySelector(item)
		m[k] = append(m[k], item)
	}

	return m
}

// GetOrDefault returns the value associated with the given key from the map `m`.
// If the key does not exist in the map, it returns the provided `defaultValue`.
// This is useful for safely retrieving values from a map without having to check for key existence.
func GetOrDefault[K comparable, T any](m map[K]T, key K, defaultValue T) T {
	v, ok := m[key]
	if !ok {
		return defaultValue
	} else {
		return v
	}
}
