package outbox

import (
	"context"
	"fmt"
	"gocms/internal/logger"
	"sync"
)

type DirectPublisher struct {
	mu          sync.RWMutex
	subscribers map[Event]EventHandler
}

func NewDirectPublisher() *DirectPublisher {
	return &DirectPublisher{
		subscribers: make(map[Event]EventHandler),
	}
}

func (b *DirectPublisher) Publish(ctx context.Context, messages []Message) ([]Message, error) {
	for i := range messages {
		if err := ctx.Err(); err != nil {
			return messages, err
		}

		b.mu.RLock()
		handler := b.subscribers[messages[i].Type]
		b.mu.RUnlock()

		if handler == nil {
			logger.Warn("Outbox.DirectPublisher", fmt.Sprintf("Received message for unhandled event %q", messages[i].Type), nil)
			messages[i].Status = StatusDead
			continue
		}
		if err := ctx.Err(); err != nil {
			return messages, err
		}

		if err := handler(ctx, messages[i].Type, messages[i].Data); err != nil {
			logger.Error(
				"Outbox.DirectPublisher",
				fmt.Sprintf("Handler for event %q failed", messages[i].Type),
				err)
			messages[i].Status = StatusFailed
		} else {
			messages[i].Status = StatusSucceced
		}
	}
	return messages, nil
}

func (b *DirectPublisher) Subscribe(event Event, handler EventHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[event] = handler
}

func (b *DirectPublisher) Unsubscribe(event Event) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.subscribers[event]; !exists {
		return false
	}

	delete(b.subscribers, event)
	return true
}
