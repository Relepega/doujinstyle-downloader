package pubsub

import (
	"fmt"
	"sync"
)

type PublishEvent struct {
	EvtType string
	Data    any
}

type Publisher struct {
	subscribers []chan *PublishEvent
	mu          sync.RWMutex
	closed      bool
}

type globalPublishers struct {
	publishers map[string]*Publisher
	mu         sync.RWMutex
}

// This allows us to reuse existing Publisher(s)
// and not instantiating new ones or passing the
// existing one through a hundred functions
var global_pubs = globalPublishers{
	publishers: make(map[string]*Publisher),
}

func NewPublisher() *Publisher {
	return &Publisher{
		subscribers: make([]chan *PublishEvent, 0),
		closed:      false,
	}
}

func NewGlobalPublisher(id string) *Publisher {
	global_pubs.mu.Lock()
	defer global_pubs.mu.Unlock()

	pub := NewPublisher()
	global_pubs.publishers[id] = pub

	return pub
}

func GetGlobalPublisher(id string) (*Publisher, error) {
	val, ok := global_pubs.publishers[id]
	if !ok {
		return nil, fmt.Errorf("Publisher not found: %s", id)
	}

	return val, nil
}

func (p *Publisher) Subscribe() <-chan *PublishEvent {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	newSubscriber := make(chan *PublishEvent)
	p.subscribers = append(p.subscribers, newSubscriber)

	return newSubscriber
}

func (p *Publisher) Publish(val *PublishEvent) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return
	}

	for _, subscriber := range p.subscribers {
		subscriber <- val
	}
}

func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	for _, subscriber := range p.subscribers {
		close(subscriber)
	}

	p.closed = true
}

func (p *Publisher) IsClosed() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.closed
}
