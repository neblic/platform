package storage

import "fmt"

var (
	ErrUnknownKey = fmt.Errorf("unknown key")
)

type Hasher interface {
	Hash() string
}

type Storage[K Hasher, V any] interface {
	Get(K) (V, error)
	Set(K, V) error
	Delete(K) error
}
