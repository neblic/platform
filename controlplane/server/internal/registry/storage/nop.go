package storage

type Nop[K comparable, V any] struct {
}

func NewNop[K comparable, V any]() *Nop[K, V] {
	return &Nop[K, V]{}
}

func (d *Nop[K, V]) Get(key K) (V, error) {
	var value V
	return value, nil
}

func (d *Nop[K, V]) Range(fn func(key K, value V)) error {
	return nil
}

func (d *Nop[K, V]) Set(key K, value V) error {
	return nil
}

func (d *Nop[K, V]) Delete(key K) error {
	return nil
}
