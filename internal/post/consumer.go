package post

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
	var payload domain.PostEventInfo
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	switch payload.Action {
	case domain.PostActionCreate, domain.PostActionUpdate:
		translations, err := s.db.GetTranslations(ctx, payload.PostID)
		if err != nil {
			logger.Error("Post.Service", "Failed to retreive post translations from db", err, "payload", payload)
			return err
		}
		_ = translations
		return nil
	case domain.PostActionDelete:
		if err := s.se.Delete(ctx, payload.PostID); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("%w: unknown action", domain.ErrInval)
}
