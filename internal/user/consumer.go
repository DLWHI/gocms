package user

import (
	"context"
	"encoding/json"
	"errors"
	"gocms/internal/errs"
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
	var username MessagePayload
	if err := json.Unmarshal(data, &username); err != nil {
		return err
	}

	user, err := s.db.Get(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil
		}
		return err
	}

	doc := user.ToDomain()
	if err := s.se.Index(ctx, &doc); err != nil {
		return err
	}
	return nil
}
