package pipelines

import (
	"sync"
)

// Queue represents a concurrent Queue.
type Queue[T any] struct {
	items []T
	mutex sync.Mutex
	cond  *sync.Cond
}

// Enqueue adds an item to the end of the queue.
func (q *Queue[T]) Enqueue(item T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, item)
	q.cond.Signal() // Notify one waiting goroutine
}

// Dequeue removes and returns the item at the front of the queue.
// If the queue is empty, it blocks until an item is available.
func (q *Queue[T]) Dequeue() T {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for len(q.items) == 0 {
		q.cond.Wait() // Wait for an item to be available
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

// IsEmpty checks if the queue is empty.
func (q *Queue[T]) IsEmpty() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.items) == 0
}

// NewQueue creates a new Queue.
func NewQueue[T any]() *Queue[T] {
	q := &Queue[T]{}
	q.cond = sync.NewCond(&q.mutex)
	return q
}
