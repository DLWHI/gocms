package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/logger"
)

type EventConsumer struct {
	users    UserRepository
	sessions SessionRepository
}

func NewEventConsumer(repo SessionRepository, manager UserRepository) *EventConsumer {
	return &EventConsumer{
		sessions: repo,
		users:    manager,
	}
}

func (s *EventConsumer) RevokeSessions(ctx context.Context, data []byte) error {
	var payload domain.UserEventInfo
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	switch payload.Action {
	case domain.UserActionPassword, domain.ActionDelete:
		if err := s.sessions.RevokeUser(ctx, payload.Username); err != nil {
			logger.Error("Auth.Service", "Failed to revoke user's sessions", err, "username", payload.Username)
			return err
		}
		return nil
	case domain.ActionCreate, domain.ActionEdit, domain.UserActionLogin:
		return nil
	}
	return fmt.Errorf("%w: unknown action", domain.ErrInval)
}
