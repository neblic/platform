package storage

import "fmt"

var (
	ErrUnknownKey = fmt.Errorf("unknown key")
)

type Storage[T any] interface {
	Set(string, T) error
	Delete(string) error
	Range(func(string, T)) error
}
