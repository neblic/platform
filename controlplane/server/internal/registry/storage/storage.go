package storage

import "fmt"

var (
	ErrUnknownKey = fmt.Errorf("unknown key")
)

type Storage[K comparable, V any] interface {
	Get(K) (V, error)
	Range(func(key K, value V)) error
	Set(K, V) error
	Delete(K) error
}
