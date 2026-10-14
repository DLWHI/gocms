package outbox

import (
	"context"
	"fmt"
	"gocms/internal/tokens"
	"sync"
)

type SubscriptionID int64

type subscription struct {
	id      SubscriptionID
	handler EventHandler
}

type MessageDispatcher struct {
	mu          sync.RWMutex
	ids         tokens.IDGenerator
	subscribers map[Event][]subscription
}

func NewMessageBroker() *MessageDispatcher {
	return &MessageDispatcher{
		subscribers: make(map[Event][]subscription),
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
		handlers := b.subscribers[message.Type]
		b.mu.RUnlock()

		for _, subscriber := range handlers {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := subscriber.handler(message.Type, message.Data); err != nil {
				return fmt.Errorf("handler for event %q failed: %w", message.Type, err)
			}
		}
	}
	return nil
}

func (b *MessageDispatcher) Subscribe(event Event, handler EventHandler) SubscriptionID {
	if handler == nil {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	id := SubscriptionID(b.ids.Next())
	b.subscribers[event] = append(b.subscribers[event], subscription{
		id:      id,
		handler: handler,
	})
	return id
}

// Unsubscribe removes a handler from an event and reports whether it existed.
func (b *MessageDispatcher) Unsubscribe(event Event, id SubscriptionID) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	current, exists := b.subscribers[event]
	if !exists {
		return false
	}

	for i, subscriber := range current {
		if subscriber.id == id {
			updated := make([]subscription, len(current)-1)
			copy(updated, current[:i])
			copy(updated[i:], current[i+1:])
			if len(updated) == 0 {
				delete(b.subscribers, event)
			} else {
				b.subscribers[event] = updated
			}
			break
		}
	}

	return true
}
