package post

import (
	"context"
	"encoding/json"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"gocms/internal/search"
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

func (s *EventConsumer) SyncCategories(ctx context.Context, data []byte) error {
	var payload domain.CategoryEventInfo
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	switch payload.Action {
	case domain.ActionDelete, domain.ActionEdit:
		processed := make(map[uint]interface{}, 0)
		count := 20
		filters := &SearchParams{
			Categories: []string{payload.Slug},
			Params: search.Params{
				Page:  1,
				Count: int64(count),
				Select: []string{
					"post_id",
				},
			},
		}
		for count == 20 {
			targets, err := s.se.Search(ctx, filters)
			if err != nil {
				logger.Error("Post.Consumer", "Failed to retreive posts on category reindex", err, "payload", payload)
				return err
			}

			ids := make([]uint, 0, len(targets.Results))
			for _, id := range targets.Results {
				if _, ok := processed[id.PostID]; !ok {
					ids = append(ids, id.PostID)
					processed[id.PostID] = nil
				}
			}

			batch, err := s.db.GetBatch(ctx, ids)
			if err != nil {
				logger.Error("Post.Consumer", "Failed to retreive post translations on category reindex", err, "payload", payload)
				return err
			}
			if err := s.se.Index(ctx, batch...); err != nil {
				logger.Error("Post.Consumer", "Failed to insert post translations to search index", err, "payload", payload)
				return err
			}
			count = len(targets.Results)
			filters.Page++
		}
		logger.Info("Post.Consumer", "Finished category update reindex", "slug", payload.Slug, "processed", len(processed))
		return nil
	case domain.ActionCreate:
		return nil
	}
	return fmt.Errorf("%w: unknown action", domain.ErrInval)
}
