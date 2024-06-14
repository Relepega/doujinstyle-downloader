package queue

import (
	"fmt"
	"sync"
)

var (
	ErrEmptyQueue   = fmt.Errorf("Queue is Empty")
	ErrNoMatchFound = fmt.Errorf("No matching element found")
)

type Node[T any] struct {
	value T
	next  *Node[T]
	prev  *Node[T]
}

func NewNode[T any](value T) *Node[T] {
	return &Node[T]{
		value: value,
		next:  nil,
		prev:  nil,
	}
}

func (n *Node[T]) Value() T {
	return n.value
}

type Queue[T any] struct {
	mu sync.Mutex
	// items  []*Node[T]
	length int
	head   *Node[T]
	tail   *Node[T]
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		// items:  make([]*Node[T], 0),
		length: 0,
		head:   nil,
		tail:   nil,
	}
}

func (q *Queue[T]) Length() int {
	return q.length
}

func (q *Queue[T]) Enqueue(node *Node[T]) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.length++

	if q.head == nil {
		q.head = node
		q.tail = node

		return
	}

	node.prev = q.tail
	q.tail.next = node
	q.tail = node
}

func (q *Queue[T]) Dequeue() (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var value T

	if q.length == 0 || q.head == nil {
		return value, ErrEmptyQueue
	}

	node := q.head
	q.head = q.head.next
	q.head.prev = nil

	if q.head.next == nil {
		q.tail = nil
	}

	q.length--

	return node.value, nil
}

func (q *Queue[T]) Front() (*Node[T], error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.length == 0 || q.head == nil {
		return nil, ErrEmptyQueue
	}

	return q.head, nil
}

func (q *Queue[T]) Back() (*Node[T], error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.length == 0 || q.tail == nil {
		return nil, ErrEmptyQueue
	}

	return q.tail, nil
}

func (q *Queue[T]) Has(value T, comparator func(val1, val2 T) bool) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	node := q.head

	for {
		if comparator(node.value, value) {
			return true
		}

		if node.next == nil {
			break
		}

		node = node.next
	}

	return false
}

func (q *Queue[T]) Remove(value T, comparator func(val1, val2 T) bool) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	node := q.head

	for {
		if comparator(node.value, value) {
			node.next.prev = node.prev
			node.prev.next = node.next

			return true
		}

		if node.next == nil {
			break
		}

		node = node.next
	}

	return false
}

func (q *Queue[T]) RemoveAll(value T, comparator func(val1, val2 T) bool) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	matched := false

	node := q.head

	for {
		if comparator(node.value, value) {
			node.next.prev = node.prev
			node.prev.next = node.next

			matched = true
		}

		if node.next == nil {
			break
		}

		node = node.next
	}

	return matched
}

// func (q *Queue[T]) Has(value T, comparator func(val1, val2 T) bool) bool {
// 	for _, item := range q.items {
// 		if comparator(item, value) {
// 			return true
// 		}
// 	}
//
// 	return false
// }
//
// func (q *Queue[T]) Remove(value T, comparator func(val1, val2 T) bool) bool {
// 	for i, item := range q.items {
// 		if comparator(item, value) {
// 			q.items = append(q.items[:i], q.items[i+1:]...)
// 			return true
// 		}
// 	}
//
// 	return false
// }
