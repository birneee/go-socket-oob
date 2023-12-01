package go_socket_oob

type Iterator[T any] interface {
	HasNext() bool
	Next() T
}
