package storage

type Noop[T any] struct {
}

func NewNoop[T any]() *Noop[T] {
	return &Noop[T]{}
}

func (d *Noop[T]) Set(key string, value T) error {
	return nil
}

func (d *Noop[T]) Delete(key string) error {
	return nil
}

func (d *Noop[T]) Range(callback func(key string, value T)) error {
	return nil
}
