package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"sync"
	"time"

	"gorm.io/gorm"
)

type Relay struct {
	broker PublishDriver
	db     *gorm.DB
}

func NewRelay(db *gorm.DB, broker PublishDriver) *Relay {
	return &Relay{
		broker: broker,
		db:     db,
	}
}

func (d *Relay) Update(ctx context.Context, messages *[]Message) error {
	var ids []uint

	for _, message := range *messages {
		ids = append(ids, message.ID)
	}

	result := d.db.Model(&Message{}).
		Where("id IN ?", ids).
		Update("proccessed", time.Now())
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *Relay) GetUnprocessed(ctx context.Context, limit int) (*[]Message, error) {
	var messages []Message
	result := d.db.WithContext(ctx).Model(&Message{}).
		Where("processed IS NULL").
		Find(&messages)
	if result.Error != nil {
		return nil, result.Error
	}
	return &messages, nil
}

func (d *Relay) ParseStartArgs(params ...any) (time.Duration, error) {
	if len(params) != 1 {
		return 0, fmt.Errorf("%v: invalid argument count", errs.ErrInval)
	}

	interval, ok := params[0].(time.Duration)
	if !ok {
		return 0, fmt.Errorf("%v: invalid duration", errs.ErrInval)
	}
	return interval, nil
}

func (d *Relay) Dispatch(ctx context.Context) error {
	messages, err := d.GetUnprocessed(ctx, 50)
	if err != nil {
		return err
	}

	if err := d.broker.Publish(ctx, messages); err != nil {
		return err
	}

	if err := d.Update(ctx, messages); err != nil {
		return err
	}

	return nil
}

func (d *Relay) Start(ctx context.Context, group *sync.WaitGroup, params ...any) {
	defer group.Done()

	interval, err := d.ParseStartArgs(params...)
	if err != nil {
		logger.Error("Outbox.MessageDispatcher", "Error parsing params", err)
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	logger.Info("Outbox.MessageDispatcher", "Started message dispatcher background worker")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Outbox.MessageDispatcher", "Message dispatcher received shutdown signal")
			return
		case <-ticker.C:
			if err := d.Dispatch(ctx); err != nil {
				logger.Error("Outbox.MessageDispatcher", "Failed to dispatch messages", err)
			}
		}
	}
}

func (d *Relay) Stop(ctx context.Context) error {
	logger.Info("Outbox.MessageDispatcher", "Executing final message dispatch")
	if err := d.Dispatch(ctx); err != nil {
		logger.Error("Outbox.MessageDispatcher", "Failed to dispatch messages", err)
		return err
	}
	return nil
}

func WriteMessage(tx *gorm.DB, action Event, source string, target any) error {
	payload, err := json.Marshal(target)
	if err != nil {
		return err
	}
	message := Message{
		SpecVersion: "1.0",
		Source:      source,
		Time:        time.Now(),
		Type:        action,
		Data:        payload,
	}
	if err := tx.Create(message).Error; err != nil {
		return err
	}
	return nil
}
