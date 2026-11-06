package ads

import (
	"context"
	"encoding/json"
	"errors"
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
	var payload domain.AdvertisementEventInfo
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	switch payload.Action {
	case domain.AdvertisemenActionCreate, domain.AdvertisementActionUpdate:
		ad, err := s.db.Get(ctx, payload.ID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil
			}
			logger.Error("Advertisement.Consumer", "Failed to remove advertisement from index", err, "payload", payload)
			return err
		}

		if err := s.se.Index(ctx, ad); err != nil {
			return err
		}
		return nil
	case domain.AdvertisementActionDelete:
		if err := s.se.Delete(ctx, payload.ID); err != nil {
			logger.Error("Advertisement.Consumer", "Failed to remove advertisement from index", err, "payload", payload)
			return err
		}
		return nil
	}
	return fmt.Errorf("%w: unknown action", domain.ErrInval)
}
