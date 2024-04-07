package pubsub

import "sync"

type PublishEvent struct {
	EvtType string
	Data    interface{}
}

type Publisher struct {
	subscribers []chan *PublishEvent
	mu          sync.RWMutex
	closed      bool
}

// This allows us to reuse an existing Publisher
// and not instantiating new ones or passing the
// existing one through a hundred functions
var pub *Publisher = nil

func NewPublisher() *Publisher {
	return &Publisher{
		subscribers: make([]chan *PublishEvent, 0),
		closed:      false,
	}
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

func GetExistingPublisher() *Publisher {
	if pub == nil {
		pub = NewPublisher()
	}

	return pub
}
