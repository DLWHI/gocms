package outbox

import (
	"context"
	"time"
)

const (
	StatusIdle = 1 << iota
	StatusPending
	StatusSucceced
	StatusFailed
	StatusDead
)

type Event string
type EventHandler func(context.Context, []byte) error

type Status int

type Subscription struct {
	Event       Event
	Unsubscribe func()
	handler     EventHandler
}

type Message struct {
	SpecVersion string     `json:"specversion" gorm:"not null"`
	ID          uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	Source      string     `json:"source" gorm:"not null"`
	Time        time.Time  `json:"time" gorm:"not null;default:current_timestamp"`
	Type        Event      `json:"type" gorm:"not null"`
	Data        []byte     `json:"data"`
	Status      Status     `json:"-" gorm:"default:1"`
	ExpiresAt   *time.Time `json:"-"`
	Attempt     uint       `json:"-" gorm:"default:0"`
}

type MessageRepository interface {
	Dispatch(ctx context.Context, count int) error
	GetUnprocessed(ctx context.Context, limit int) ([]Message, error)
}

type Publisher interface {
	Publish(ctx context.Context, messages []Message) ([]Message, error)
	Subscribe(event Event, handler EventHandler) *Subscription
	Unsubscribe(sub *Subscription)
}

func (Message) TableName() string {
	return "outbox"
}
