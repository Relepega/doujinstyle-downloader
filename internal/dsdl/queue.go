// The package implements a Queue, a Task Tracker and a Wrapper to keep both in sync
//
// Queue: A basic queue implementation based on a doubly linked-list.
//
// Tracker: A map that keeps track of the progress of every task added in it.
//
// TQWrapper: The recommended way of interacting with the package functionality if you need both queuing and tracking functionality.
//
// This wrapper ensures that everything is synchronized correctly.
package dsdl

import (
	"fmt"
	"sync"
)

var (
	ErrEmptyQueue   = fmt.Errorf("Queue is Empty")
	ErrNoMatchFound = fmt.Errorf("No matching element found")
)

// Data type representing a linked-list node
type Node struct {
	// Value of the node. Is set on the constructor and is not editable
	value      interface{}
	next, prev *Node
}

// Constructor for the Node data type
//
// The value cannot be modified after being attached to a Node
func NewNode(value interface{}) *Node {
	return &Node{
		value: value,
		next:  nil,
		prev:  nil,
	}
}

// Returns the value hold within the node
func (n *Node) Value() interface{} {
	return n.value
}

// Data type representing a Queue, based on a doubly linked-list
type Queue struct {
	mu     sync.Mutex
	length int
	head   *Node
	tail   *Node
}

// Constructor for the Queue data type
func NewQueue() *Queue {
	return &Queue{
		length: 0,
		head:   nil,
		tail:   nil,
	}
}

// Returns the length of the queue
func (q *Queue) Length() int {
	return q.length
}

// Inserts a Node in the queue at its tail
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

// Removes the first Node of the queue
func (q *Queue) Dequeue() (any, error) {
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

// Returns either the first Node of the queue without dequeueing it
//
// or an error if the queue is empty
func (q *Queue) Front() (*Node, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.length == 0 || q.head == nil {
		return nil, ErrEmptyQueue
	}

	return q.head, nil
}

// Returns either the last Node of the queue without dequeueing it
//
// or an error if the queue is empty
func (q *Queue) Back() (*Node, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.length == 0 || q.tail == nil {
		return nil, ErrEmptyQueue
	}

	return q.tail, nil
}

// Returns wether or not a node with the same value exists in the queue
//
// The comparation between values is done in a comparator function
func (q *Queue) Has(
	value interface{},
	comparator func(val1, val2 interface{}) bool,
) (bool, interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	node := q.head

	for {
		if comparator(node.value, value) {
			return true, node.value
		}

		if node.next == nil {
			break
		}

		node = node.next
	}

	return false, nil
}

// Removes A SINGLE NODE with the same value if it exists in the queue. The comparation between values is done in a comparator function
//
// Returns wether or not the node has been found and removed
func (q *Queue) Remove(
	value interface{},
	comparator func(val1, val2 interface{}) bool,
) (bool, interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	node := q.head

	for {
		if node == nil {
			break
		}

		if comparator(node.value, value) {
			if node.next != nil {
				node.next.prev = node.prev
			}

			if node.prev != nil {
				node.prev.next = node.next
			}

			return true, node.value
		}

		node = node.next
	}

	return false, nil
}

// Removes ALL THE NODE(S) with the same value if it exists in the queue. The comparation between values is done in a comparator function
//
// Returns wether or not the node(s) has been found and removed
func (q *Queue) RemoveAll(value interface{}, comparator func(val1, val2 interface{}) bool) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	matched := false

	node := q.head

	for {
		if node == nil {
			break
		}

		if comparator(node.value, value) {
			if node.next != nil {
				node.next.prev = node.prev
			}

			if node.prev != nil {
				node.prev.next = node.next
			}

			matched = true
		}

		node = node.next
	}

	return matched
}

// Clears the queue
func (q *Queue) Reset(value interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.head = nil
	q.tail = nil
	q.length = 0
}
