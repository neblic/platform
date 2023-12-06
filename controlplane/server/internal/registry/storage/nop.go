package storage

type Nop[K comparable, V any] struct {
}

func NewNop[K comparable, V any]() *Nop[K, V] {
	return &Nop[K, V]{}
}

func (d *Nop[K, V]) Get(_ K) (V, error) {
	var value V
	return value, nil
}

func (d *Nop[K, V]) Range(_ func(key K, value V)) error {
	return nil
}

func (d *Nop[K, V]) Set(_ K, _ V) error {
	return nil
}

func (d *Nop[K, V]) Delete(_ K) error {
	return nil
}
