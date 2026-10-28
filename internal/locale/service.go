package locale

import (
	"context"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/logger"
	"strings"
)

const MaxCodeLength = 2

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func IsInvalidCodeRune(r rune) bool {
	return !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z')
}

func NormalizeCode(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", fmt.Errorf("%w: locale code cannot be empty", domain.ErrInval)
	}
	if len(normalized) > MaxCodeLength {
		return "", fmt.Errorf("%w: locale code is longer than %d chars", domain.ErrInval, MaxCodeLength)
	}
	if strings.IndexFunc(normalized, IsInvalidCodeRune) >= 0 {
		return "", fmt.Errorf("%w: locale code must contain only letters", domain.ErrInval)
	}

	return strings.ToLower(normalized), nil
}

func (s *Service) Create(ctx context.Context, name string, code string) error {
	if !auth.HasPermission(ctx, "locales", "create") {
		return domain.ErrCred
	}

	if err := s.repo.Create(ctx, name, code); err != nil {
		logger.Error("Locale.Service", "Failed to create locale", err)
		return err
	}

	return nil
}

func (s *Service) GetAll(ctx context.Context) ([]Locale, error) {
	locales, err := s.repo.GetAll(ctx)
	if err != nil {
		logger.Error("Locale.Service", "Failed to list locales", err)
		return nil, err
	}

	return locales, nil
}

func (s *Service) Delete(ctx context.Context, code string) error {
	if !auth.HasPermission(ctx, "locales", "delete") {
		return domain.ErrCred
	}

	if err := s.repo.Delete(ctx, code); err != nil {
		logger.Error("Locale.Service", "Failed to delete locale", err)
		return err
	}

	return nil
}
