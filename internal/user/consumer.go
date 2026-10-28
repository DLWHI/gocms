package user

import (
	"context"
	"encoding/json"
	"errors"
	"gocms/internal/domain"
	"gocms/internal/logger"
)

type EventConsumer struct {
	db Repository
	se SearchEngine
}

func NewEventConsumer(db Repository, se SearchEngine) *EventConsumer {
	return &EventConsumer{
		db: db,
		se: se,
	}
}

func (s *EventConsumer) SyncSearch(ctx context.Context, data []byte) error {
	var payload domain.UserEvent
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if payload.Action == domain.UserActionPassword {
		return nil
	}

	user, err := s.db.Get(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		logger.Error("User.Service", "Failed to retreive user from db", err, "payload", payload)
		return err
	}

	if err := s.se.Index(ctx, user); err != nil {
		logger.Error("User.Service", "Failed to index user", err, "payload", payload, "user", user)
		return err
	}
	return nil
}
