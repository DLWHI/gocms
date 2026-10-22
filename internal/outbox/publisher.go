package outbox

import (
	"context"
	"fmt"
	"gocms/internal/logger"
	"sync"
)

type DirectPublisher struct {
	mu          sync.RWMutex
	subscribers map[Event][]*Subscription
}

func NewDirectPublisher() *DirectPublisher {
	return &DirectPublisher{
		subscribers: make(map[Event][]*Subscription),
	}
}

func (b *DirectPublisher) Publish(ctx context.Context, messages []Message) ([]Message, error) {
	for i := range messages {
		if err := ctx.Err(); err != nil {
			return messages, err
		}

		b.mu.RLock()
		subscriptions := b.subscribers[messages[i].Type]
		b.mu.RUnlock()

		if subscriptions == nil {
			logger.Warn("Outbox.DirectPublisher", fmt.Sprintf("Received message for unhandled event %q", messages[i].Type), nil)
			messages[i].Status = StatusDead
			continue
		}
		if err := ctx.Err(); err != nil {
			return messages, err
		}

		for _, sub := range subscriptions {
			if err := sub.handler(ctx, messages[i].Data); err != nil {
				logger.Error(
					"Outbox.DirectPublisher",
					fmt.Sprintf("Handler for event %q failed", messages[i].Type),
					err)
				messages[i].Status = StatusFailed
			} else {
				messages[i].Status = StatusSucceced
			}
		}
	}
	return messages, nil
}

func (b *DirectPublisher) Subscribe(event Event, handler EventHandler) *Subscription {
	if handler == nil {
		return nil
	}

	sub := Subscription{
		Event:   event,
		handler: handler,
	}

	sub.Unsubscribe = func() {
		b.Unsubscribe(&sub)
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[event] = append(b.subscribers[event], &sub)
	return &sub
}

func (b *DirectPublisher) Unsubscribe(sub *Subscription) {
	event := sub.Event
	handlers := b.subscribers[event]
	cp := make([]*Subscription, 0, len(handlers)-1)
	for _, handler := range handlers {
		if handler != sub {
			cp = append(cp, handler)
		}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(cp) > 0 {
		b.subscribers[event] = cp
	} else {
		delete(b.subscribers, event)
	}
}
