package queue

import (
	"fmt"
	"sync"
)

var (
	ErrEmptyQueue   = fmt.Errorf("Queue is Empty")
	ErrNoMatchFound = fmt.Errorf("No matching element found")
)

type Node struct {
	value interface{}
	next  *Node
	prev  *Node
}

func NewNode(value interface{}) *Node {
	return &Node{
		value: value,
		next:  nil,
		prev:  nil,
	}
}

func (n *Node) Value() interface{} {
	return n.value
}

type Queue struct {
	mu sync.Mutex
	// items  []*Node
	length int
	head   *Node
	tail   *Node
}

func NewQueue() *Queue {
	return &Queue{
		// items:  make([]*Node, 0),
		length: 0,
		head:   nil,
		tail:   nil,
	}
}

func (q *Queue) Length() int {
	return q.length
}

func (q *Queue) Enqueue(node *Node) {
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

// func (q *Queue) Dequeue() (interface{}, error) {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()
//
// 	var value interface{}
//
// 	if q.length == 0 || q.head == nil {
// 		return value, ErrEmptyQueue
// 	}
//
// 	node := q.head
// 	q.head = q.head.next
// 	q.head.prev = nil
//
// 	if q.head.next == nil {
// 		q.tail = nil
// 	}
//
// 	q.length--
//
// 	return node.value, nil
// }

func (q *Queue) Dequeue() (interface{}, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.length == 0 || q.head == nil {
		return nil, ErrEmptyQueue
	}

	node := q.head
	q.head = q.head.next

	if q.head != nil {
		q.head.prev = nil
	} else {
		q.tail = nil
	}

	q.length--

	return node.value, nil
}

func (q *Queue) Front() (*Node, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.length == 0 || q.head == nil {
		return nil, ErrEmptyQueue
	}

	return q.head, nil
}

func (q *Queue) Back() (*Node, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.length == 0 || q.tail == nil {
		return nil, ErrEmptyQueue
	}

	return q.tail, nil
}

func (q *Queue) Has(value interface{}, comparator func(val1, val2 interface{}) bool) bool {
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

func (q *Queue) Remove(value interface{}, comparator func(val1, val2 interface{}) bool) bool {
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

func (q *Queue) RemoveAll(value interface{}, comparator func(val1, val2 interface{}) bool) bool {
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

// func (q *Queue) Has(value interface{}, comparator func(val1, val2 interface{}) bool) bool {
// 	for _, item := range q.items {
// 		if comparator(item, value) {
// 			return true
// 		}
// 	}
//
// 	return false
// }
//
// func (q *Queue) Remove(value interface{}, comparator func(val1, val2 interface{}) bool) bool {
// 	for i, item := range q.items {
// 		if comparator(item, value) {
// 			q.items = append(q.items[:i], q.items[i+1:]...)
// 			return true
// 		}
// 	}
//
// 	return false
// }
