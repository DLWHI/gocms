package outbox

import (
	"context"
	"fmt"
	"gocms/internal/logger"
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

func (b *MessageDispatcher) Publish(ctx context.Context, messages *[]Message) ([]Message, error) {
	if messages == nil {
		return make([]Message, 0), nil
	}
	processed := make([]Message, 0, len(*messages))

	for _, message := range *messages {
		if err := ctx.Err(); err != nil {
			return processed, err
		}

		b.mu.RLock()
		handler := b.subscribers[message.Type]
		b.mu.RUnlock()

		if handler != nil {
			if err := ctx.Err(); err != nil {
				return processed, err
			}
			for i := 0; i < 5; i++ {
				if err := handler(ctx, message.Type, message.Data); err != nil {
					logger.Error(
						"Outbox.MessageDispatcher",
						fmt.Sprintf("Handler for event %q failed on try %d/5", message.Type, i),
						err)
				} else {
					processed = append(processed, message)
					break
				}
			}
		}
	}
	return processed, nil
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
