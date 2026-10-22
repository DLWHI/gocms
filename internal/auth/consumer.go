package auth

import (
	"context"
	"encoding/json"
	"gocms/internal/domain"
	"gocms/internal/session"
)

type EventConsumer struct {
	sessions session.Repository
}

func NewEventConsumer(repo session.Repository) *EventConsumer {
	return &EventConsumer{
		sessions: repo,
	}
}

func (s *EventConsumer) RevokeSessions(ctx context.Context, data []byte) error {
	var payload domain.UserEvent
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if payload.Action != domain.UserActionPasswd && payload.Action != domain.UserActionSuspend {
		return nil
	}

	if err := s.sessions.RevokeUser(ctx, payload.Username); err != nil {
		return err
	}
	return nil
}
