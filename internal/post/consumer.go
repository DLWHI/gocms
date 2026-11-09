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
	case domain.ActionCreate, domain.ActionEdit:
		translations, err := s.db.GetTranslations(ctx, payload.PostID)
		if err != nil {
			logger.Error("Post.Consumer", "Failed to retreive post translations from db", err, "payload", payload)
			return err
		}

		if err := s.se.Index(ctx, translations...); err != nil {
			logger.Error("Post.Consumer", "Failed to insert post translations to search index", err, "payload", payload)
			return err
		}
		return nil
	case domain.ActionDelete:
		if err := s.se.Delete(ctx, payload.PostID); err != nil {
			logger.Error("Post.Consumer", "Failed to remove post from index", err, "payload", payload)
			return err
		}
		return nil
	}
	return fmt.Errorf("%w: unknown action", domain.ErrInval)
}

func (s *EventConsumer) ReIndexCategory(ctx context.Context, data []byte) error {
	var payload domain.CategoryEventInfo
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	switch payload.Action {
	case domain.ActionDelete, domain.ActionEdit:
		// count := 20
		// filters := &SearchParams{
		// 	Categories: []string{payload.Slug},
		// 	Params: search.Params{
		// 		Page:  1,
		// 		Count: int64(count),
		// 	},
		// }
		// for count == 20 {
		// 	batch, err := s.se.Search(ctx, filters)
		// 	if err != nil {
		// 		logger.Error("Post.Consumer", "Failed to retreive post translations from db on category reindex", err, "payload", payload)
		// 	}

		// 	if err := s.se.Index(ctx, batch.Results...); err != nil {
		// 		logger.Error("Post.Consumer", "Failed to insert post translations to search index", err, "payload", payload)
		// 		return err
		// 	}
		// 	count = len(batch.Results)
		// 	filters.Page++
		// }
		return nil
	}
	return fmt.Errorf("%w: unknown action", domain.ErrInval)
}
