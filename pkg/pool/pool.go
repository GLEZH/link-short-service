// Package pool provides a typed pool for resettable objects
package pool

import "sync"

// Resetter resets an object to its initial state
type Resetter interface {
	Reset()
}

// Pool stores reusable objects of one type
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New creates a pool that uses create when it is empty
func New[T Resetter](create func() T) *Pool[T] {
	if create == nil {
		panic("pool: nil constructor")
	}

	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return create()
			},
		},
	}
}

// Get returns an object from the pool
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put resets an object and returns it to the pool
func (p *Pool[T]) Put(value T) {
	value.Reset()
	p.pool.Put(value)
}
