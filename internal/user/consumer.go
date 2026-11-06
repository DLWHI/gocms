package user

import (
	"context"
	"encoding/json"
	"fmt"
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
	var payload domain.UserEventInfo
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	switch payload.Action {
	case domain.UserActionPassword:
		return nil
	case domain.ActionCreate, domain.ActionEdit, domain.ActionDelete, domain.UserActionLogin:
		user, err := s.db.Get(ctx, payload.Username)
		if err != nil {
			logger.Error("User.Consumer", "Failed to retreive user from db", err, "payload", payload)
			return err
		}

		if err := s.se.Index(ctx, user); err != nil {
			logger.Error("User.Consumer", "Failed to index user", err, "payload", payload, "user", user)
			return err
		}
		return nil
	}
	return fmt.Errorf("%w: unknown action", domain.ErrInval)
}
