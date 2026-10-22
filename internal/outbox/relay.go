package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Relay struct {
	broker  Publisher
	mu      sync.Mutex
	db      *gorm.DB
	timeout time.Duration
}

func NewRelay(db *gorm.DB, broker Publisher) *Relay {
	return &Relay{
		broker: broker,
		db:     db,
	}
}

func (d *Relay) CommitMessages(ctx context.Context, messages []Message) error {
	var failed []uint
	var dead []uint
	var succeced []uint

	for _, message := range messages {
		switch message.Status {
		case StatusFailed:
			if message.Attempt == 5 {
				dead = append(dead, message.ID)
			} else {
				failed = append(failed, message.ID)
			}
		case StatusPending:
			failed = append(failed, message.ID)
		case StatusSucceced:
			succeced = append(succeced, message.ID)
		case StatusDead:
			dead = append(dead, message.ID)
		}
	}

	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := d.Mark(tx, succeced, StatusSucceced); err != nil {
			return err
		}
		if err := d.Mark(tx, failed, StatusFailed); err != nil {
			return err
		}
		if err := d.Mark(tx, dead, StatusDead); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *Relay) GetUnprocessed(ctx context.Context, limit int) ([]Message, error) {
	var messages []Message
	for i := 0; i < 5; i++ {
		err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var ids []uint
			query := tx.Model(&Message{}).Clauses(clause.Locking{
				Strength: "UPDATE",
				Options:  "SKIP LOCKED",
			}).Where("status = ?", StatusIdle).
				Or("status = ?", StatusFailed).
				Order("time ASC").
				Order("id ASC")
			if limit != 0 {
				query = query.Limit(limit)
			}
			result := query.Find(&messages)
			if result.Error != nil {
				return result.Error
			}

			if len(messages) == 0 {
				return nil
			}

			for i := range messages {
				messages[i].Attempt++
				ids = append(ids, messages[i].ID)
			}

			result = tx.Model(&Message{}).
				Where("id IN ? AND status != ?", ids, StatusPending).
				Updates(map[string]interface{}{
					"status":     StatusPending,
					"expires_at": time.Now().Add(d.timeout),
					"attempt":    gorm.Expr("attempt + 1"),
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != int64(len(messages)) {
				return domain.ErrLock
			}
			return nil
		})
		if err == nil {
			return messages, nil
		} else if !errors.Is(err, domain.ErrLock) {
			return nil, err
		}
	}
	return nil, domain.ErrLock
}

func (d *Relay) RecoverStale(ctx context.Context) error {
	result := d.db.WithContext(ctx).Model(&Message{}).
		Where("status = ?", StatusPending).
		Where("expires_at <= ?", time.Now()).
		Updates(map[string]interface{}{
			"expires_at": nil,
			"status":     StatusFailed,
		})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *Relay) ParseStartArgs(params ...any) (time.Duration, error) {
	if len(params) != 1 {
		return 0, fmt.Errorf("%v: invalid argument count", domain.ErrInval)
	}

	interval, ok := params[0].(time.Duration)
	if !ok || interval <= 0 {
		return 0, fmt.Errorf("%v: invalid duration", domain.ErrInval)
	}
	return interval, nil
}

func (d *Relay) Dispatch(ctx context.Context, count int) error {
	messages, err := d.GetUnprocessed(ctx, count)
	if err != nil {
		return err
	}

	processed, err := d.broker.Publish(ctx, messages)
	if err != nil {
		logger.Error("Outbox.Relay", "Publisher failed proccessing messages.", err)
	}

	if err := d.CommitMessages(ctx, processed); err != nil {
		return err
	}
	return nil
}

func (d *Relay) Start(ctx context.Context, group *sync.WaitGroup, params ...any) {
	defer group.Done()

	interval, err := d.ParseStartArgs(params...)
	if err != nil {
		logger.Error("Outbox.Relay", "Error parsing params", err)
		return
	}

	d.timeout = interval * 5
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	recovery := time.NewTicker(interval * 6)
	defer recovery.Stop()
	logger.Info("Outbox.Relay", "Started message dispatcher background worker")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Outbox.Relay", "Message dispatcher received shutdown signal")
			return
		case <-ticker.C:
			if err := d.Dispatch(ctx, 50); err != nil {
				logger.Error("Outbox.Relay", "Failed to dispatch messages.", err)
			}
		case <-recovery.C:
			if err := d.RecoverStale(ctx); err != nil {
				logger.Error("Outbox.Relay", "Failed to recover stale messages.", err)
			}
		}
	}
}

func (d *Relay) Stop(ctx context.Context) error {
	return nil
}

func (d *Relay) Mark(tx *gorm.DB, ids []uint, status Status) error {
	result := tx.Model(&Message{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":     status,
			"expires_at": nil,
		})
	if result.Error != nil {
		return result.Error
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
