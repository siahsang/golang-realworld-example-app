package collectionutils

import "sync"

type SafeMap[K comparable, V any] struct {
	data   map[K]V
	mutext sync.RWMutex
}

func (safeMap *SafeMap[K, V]) Store(newKey K, newValue V) {
	safeMap.mutext.Lock()
	defer safeMap.mutext.Unlock()
	safeMap.data[newKey] = newValue

}

func (safeMap *SafeMap[K, V]) Get(key K) (V, bool) {
	safeMap.mutext.RLock()
	defer safeMap.mutext.RUnlock()
	value, exists := safeMap.data[key]

	return value, exists
}

func (safeMap *SafeMap[K, V]) Delete(key K) {
	safeMap.mutext.Lock()
	defer safeMap.mutext.Unlock()
	delete(safeMap.data, key)
}

func New[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		data: make(map[K]V),
	}
}
