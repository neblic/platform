package storage

type Memory[K comparable, V any] struct {
	data map[K]V
}

func NewMemory[K comparable, V any]() *Memory[K, V] {
	return &Memory[K, V]{
		data: map[K]V{},
	}
}

func (s *Memory[K, V]) Get(key K) (V, error) {
	value, ok := s.data[key]
	if !ok {
		return value, ErrUnknownKey
	}
	return value, nil
}

func (s *Memory[K, V]) Set(key K, value V) error {
	s.data[key] = value

	return nil
}

func (s *Memory[K, V]) Delete(key K) error {
	if _, ok := s.data[key]; !ok {
		return ErrUnknownKey
	}

	delete(s.data, key)

	return nil
}
