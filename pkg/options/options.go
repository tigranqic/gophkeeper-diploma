// Package options provides a generic functional-options pattern.
package options

// Option is a generic functional option that mutates a value of type T.
// Callers construct named options with constructor functions:
//
//	func WithFoo(v V) Option[T] { return func(t *T) { t.foo = v } }
type Option[T any] func(*T)

// Apply calls every option in opts on v in the order they were supplied.
func Apply[T any](v *T, opts []Option[T]) {
	for _, o := range opts {
		o(v)
	}
}
