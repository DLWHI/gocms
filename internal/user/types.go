package user

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/search"
	"time"

	"gorm.io/gorm"
)

const (
	EntityName = "users"
)

type Password = domain.Password

type Record struct {
	Username    string         `json:"username" gorm:"primaryKey;size:64;"`
	Passwd      Password       `json:"passwd,omitempty" gorm:"not null"`
	Email       string         `json:"email"`
	Permissions uint32         `json:"permissions" gorm:"default:0"` // PermCreate
	SuspendedAt gorm.DeletedAt `json:"suspended_at"`
	LastLogin   *time.Time     `json:"last_login"`
	Version     uint           `json:"-" gorm:"default:0"`
}

type DTO struct {
	domain.User
	Passwd Password `json:"passwd,omitempty"`
}

type SearchParams struct {
	search.Params
	Permissions     *uint32    `json:"permissions"`
	Suspended       *bool      `json:"suspended"`
	SuspendedAfter  *time.Time `json:"suspended_after"`
	SuspendedBefore *time.Time `json:"suspended_before"`
	LastLoginAfter  *time.Time `json:"last_login_after"`
	LastLoginBefore *time.Time `json:"last_login_before"`
}

type SearchResult = search.Result[domain.User]

type Repository interface {
	Create(ctx context.Context, user *Record) error
	UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*Record, error)
	Update(ctx context.Context, username string, diff map[string]interface{}) error
	UpdatePasswd(ctx context.Context, username string, pwdhash Password) error
	Get(ctx context.Context, username string) (*Record, error)
	GetActive(ctx context.Context) ([]Record, error)
	GetPasswdHash(ctx context.Context, username string) (Password, error)
	Suspend(ctx context.Context, username string) error
}

type SearchEngine interface {
	Index(ctx context.Context, user *domain.User) error
	Get(ctx context.Context, username string) (*domain.User, error)
	Search(ctx context.Context, filters *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, username string) error
}

func (ur *Record) ToDomain() domain.User {
	var ud domain.User

	if ur.SuspendedAt.Valid {
		ts := ur.SuspendedAt.Time.Unix()
		ud.SuspendedAt = &ts
	}

	if ur.LastLogin != nil {
		ts := ur.LastLogin.Unix()
		ud.LastLogin = &ts
	}

	email := ur.Email
	perms := ur.Permissions

	ud.Username = ur.Username
	ud.Email = &email
	ud.Permissions = &perms
	return ud
}

func (ur *DTO) ToDiff(ud *domain.User) map[string]interface{} {
	diff := make(map[string]interface{})

	if ur.Permissions != nil {
		diff["permissions"] = *ur.Permissions
	}
	if ur.Email != nil {
		diff["email"] = *ur.Email
	}
	return diff
}

func ToRecord(ud *domain.User) Record {
	var um Record
	if ud.SuspendedAt != nil {
		um.SuspendedAt = gorm.DeletedAt{
			Time:  time.Unix(*ud.SuspendedAt, 0).UTC(),
			Valid: true,
		}
	}

	if ud.LastLogin != nil {
		ts := time.Unix(*ud.LastLogin, 0).UTC()
		um.LastLogin = &ts
	}

	um.Username = ud.Username
	if ud.Email != nil {
		um.Email = *ud.Email
	}
	if ud.Permissions != nil {
		um.Permissions = *ud.Permissions
	}
	return um
}

func (Record) TableName() string {
	return "users"
}
