package editor

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
	var payload domain.EditorEvent
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	switch payload.Action {
	case domain.EditorActionCreate, domain.EditorActionUpdate, domain.EditorActionDelete:
		translations, err := s.db.GetTranslations(ctx, payload.Slug)
		if err != nil {
			logger.Error("Editor.Service", "Failed to retreive editor translations from db", err, "payload", payload)
			return err
		}

		if err := s.se.Index(ctx, payload.Slug, translations...); err != nil {
			logger.Error("Editor.Service", "Failed to editors", err, "payload", payload)
			return err
		}
		return nil
	}
	return fmt.Errorf("%w: unknown action", domain.ErrInval)
}
