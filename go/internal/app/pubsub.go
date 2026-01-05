package app

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrPubSubClosed = errors.New("pubsub is closed")
)

type PubSub struct {
	mu     sync.RWMutex
	closed bool
	nextID int
	subs   map[int]chan Event
}

func NewPubSub() *PubSub {
	return &PubSub{
		subs: make(map[int]chan Event),
	}
}

func (p *PubSub) Publish(ctx context.Context, event Event) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return ErrPubSubClosed
	}

	for _, ch := range p.subs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- event:
		}
	}
	return nil
}

func (p *PubSub) Subscribe(buffer int) (<-chan Event, func()) {
	if buffer <= 0 {
		buffer = 1
	}
	ch := make(chan Event, buffer)

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		close(ch)
		return ch, func() {}
	}
	if len(p.subs) > 0 {
		p.mu.Unlock()
		close(ch)
		return ch, func() {}
	}
	id := p.nextID
	p.nextID++
	p.subs[id] = ch
	p.mu.Unlock()

	cancel := func() {
		p.mu.Lock()
		if existing, ok := p.subs[id]; ok {
			delete(p.subs, id)
			close(existing)
		}
		p.mu.Unlock()
	}

	return ch, cancel
}

func (p *PubSub) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	for id, ch := range p.subs {
		delete(p.subs, id)
		close(ch)
	}
	p.mu.Unlock()
}
