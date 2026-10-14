package outbox

import (
	"context"
	"fmt"
	"sync"
)

type MessageDispatcher struct {
	mu          sync.RWMutex
	subscribers map[Event]EventHandler
}

func NewMessageDispatcher() *MessageDispatcher {
	return &MessageDispatcher{
		subscribers: make(map[Event]EventHandler),
	}
}

func (b *MessageDispatcher) Publish(ctx context.Context, messages *[]Message) error {
	if messages == nil {
		return fmt.Errorf("messages must not be nil")
	}

	for _, message := range *messages {
		if err := ctx.Err(); err != nil {
			return err
		}

		b.mu.RLock()
		handler := b.subscribers[message.Type]
		b.mu.RUnlock()

		if handler != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := handler(ctx, message.Type, message.Data); err != nil {
				return fmt.Errorf("Handler for event %q failed: %w", message.Type, err)
			}
		}
	}
	return nil
}

func (b *MessageDispatcher) Subscribe(event Event, handler EventHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[event] = handler
}

func (b *MessageDispatcher) Unsubscribe(event Event) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.subscribers[event]; !exists {
		return false
	}

	delete(b.subscribers, event)
	return true
}
