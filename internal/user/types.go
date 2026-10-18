package user

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/search"
	"time"

	"gorm.io/gorm"
)

type Password = domain.Password

type User struct {
	Username    string         `json:"username" gorm:"primaryKey;size:64;"`
	Passwd      Password       `json:"passwd,omitempty" gorm:"not null"`
	Email       string         `json:"email"`
	Permissions uint32         `json:"permissions" gorm:"default:0"` // PermCreate
	SuspendedAt gorm.DeletedAt `json:"suspended_at"`
	LastLogin   *time.Time     `json:"last_login"`
	Version     uint           `json:"-" gorm:"default:0"`
}

type Document = domain.User

type MessagePayload struct {
	Username string `json:"username"`
	Action   string `json:"action"`
}

type SearchParams struct {
	search.Params
	Permissions     uint32     `json:"permissions"`
	Suspended       *bool      `json:"suspended"`
	SuspendedAfter  *time.Time `json:"suspended_after"`
	SuspendedBefore *time.Time `json:"suspended_before"`
	LastLoginAfter  *time.Time `json:"last_login_after"`
	LastLoginBefore *time.Time `json:"last_login_before"`
}

type SearchResult = search.Result[Document]

type Repository interface {
	Create(ctx context.Context, user *User) error
	UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*User, error)
	Update(ctx context.Context, username string, diff map[string]interface{}) (*User, error)
	UpdatePasswd(ctx context.Context, username string, pwdhash Password) error
	Get(ctx context.Context, username string) (*User, error)
	GetAll(ctx context.Context) ([]User, error)
	GetPasswdHash(ctx context.Context, username string) (Password, error)
	Delete(ctx context.Context, username string) error
}

type SearchEngine interface {
	Index(ctx context.Context, user *Document) error
	Get(ctx context.Context, username string) (*Document, error)
	Search(ctx context.Context, filters *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, username string, timestamp int64) error
}

func (um *User) ToDocument() Document {
	var ud Document

	if um.SuspendedAt.Valid {
		ts := um.SuspendedAt.Time.Unix()
		ud.SuspendedAt = &ts
	}

	if um.LastLogin != nil {
		ts := um.LastLogin.Unix()
		ud.LastLogin = &ts
	}

	ud.Username = um.Username
	ud.Email = um.Email
	ud.Permissions = um.Permissions
	return ud
}

func ToDomain(ud *Document) User {
	var um User
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
	um.Email = ud.Email
	um.Permissions = ud.Permissions
	return um
}

func (User) TableName() string {
	return "users"
}
