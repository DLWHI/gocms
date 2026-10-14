package outbox

import (
	"context"
	"time"
)

type Event string
type EventHandler func(context.Context, Event, []byte) error

type Message struct {
	SpecVersion string     `json:"specversion" gorm:"not null"`
	ID          uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	Source      string     `json:"source" gorm:"not null"`
	Time        time.Time  `json:"time" gorm:"not null;default:current_timestamp"`
	Type        Event      `json:"type" gorm:"not null"`
	Data        []byte     `json:"data"`
	Proccessed  *time.Time `json:"-" gorm:"default:false"`
}

type MessageRepository interface {
	Update(ctx context.Context, messages *[]Message) error
	GetUnprocessed(ctx context.Context, limit int) (*[]Message, error)
}

type PublishDriver interface {
	Publish(ctx context.Context, messages *[]Message) error
	Subscribe(event Event, handler EventHandler)
	Unsubscribe(event Event) bool
}

func (Message) TableName() string {
	return "outbox"
}
