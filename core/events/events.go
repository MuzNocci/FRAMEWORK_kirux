package events

import "sync"

type Handler func(payload any)

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{handlers: make(map[string][]Handler)}
}

func (b *Bus) Subscribe(event string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[event] = append(b.handlers[event], h)
}

func (b *Bus) Publish(event string, payload any) {
	b.mu.RLock()
	handlers := b.handlers[event]
	b.mu.RUnlock()

	for _, h := range handlers {
		go h(payload)
	}
}

func (b *Bus) Unsubscribe(event string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.handlers, event)
}
