package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"gocms/internal/session"
)

type EventConsumer struct {
	users    UserManager
	sessions session.Repository
}

func NewEventConsumer(repo session.Repository, manager UserManager) *EventConsumer {
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
			if !errors.Is(err, domain.ErrExpired) {
				logger.Error("Auth.Service", "Failed to revoke user's sessions", err, "username", payload.Username)
			}
			return err
		}
		return nil
	case domain.UserActionUpdate:
		user, err := s.users.Get(ctx, payload.Username)
		if err != nil {
			return err
		}

		if err := s.sessions.Update(ctx, payload.Username, user); err != nil {
			if !errors.Is(err, domain.ErrExpired) {
				logger.Error("Auth.Service", "Failed to revoke user's sessions", err, "username", payload.Username)
			}
			return err
		}
		return nil
	case domain.UserActionCreate, domain.UserActionLogin:
		return nil
	}
	return fmt.Errorf("%w: unknown action", domain.ErrInval)
}
