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
	var payload domain.UserEvent
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	switch payload.Action {
	case domain.UserActionPassword, domain.UserActionSuspend:
		if err := s.sessions.RevokeUser(ctx, payload.Username); err != nil {
			logger.Error("Auth.Service", "Failed to revoke user's sessions", err, "username", payload.Username)
			return err
		}
		return nil
	case domain.UserActionCreate, domain.UserActionLogin, domain.UserActionUpdate:
		return nil
	}
	return fmt.Errorf("%w: unknown action", domain.ErrInval)
}
