package user

import (
	"context"
	"encoding/json"
	"gocms/internal/search"
	"time"

	"gorm.io/gorm"
)

type Bytes []byte

func (p *Bytes) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p = Bytes(s)
	return nil
}

type Model struct {
	Username    string         `json:"username" gorm:"primaryKey;size:64;"`
	Passwd      Bytes          `json:"passwd,omitempty" gorm:"not null"`
	Email       string         `json:"email"`
	Permissions uint32         `json:"permissions" gorm:"default:0"` // PermCreate
	SuspendedAt gorm.DeletedAt `json:"suspended_at"`
	LastLogin   *time.Time     `json:"last_login"`
	Version     uint           `json:"-" gorm:"default:0"`
}

type Session struct {
	ID   string `json:"session_id"`
	User Model  `json:"user"`
}

type Document struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Permissions uint32 `json:"permissions"`
	SuspendedAt *int64 `json:"suspended_at"`
	LastLogin   *int64 `json:"last_login"`
}

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
	Create(ctx context.Context, user *Model) error
	UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*Model, error)
	UpdatePermissions(ctx context.Context, username string, permissions uint32) (*Model, error)
	UpdateEmail(ctx context.Context, username string, email string) error
	UpdatePasswd(ctx context.Context, username string, pwdhash Bytes) error
	Get(ctx context.Context, username string) (*Model, error)
	GetAll(ctx context.Context) ([]Model, error)
	GetPasswdHash(ctx context.Context, username string) (Bytes, error)
	Delete(ctx context.Context, username string) error
}

type SearchEngine interface {
	Index(ctx context.Context, user *Document) error
	Get(ctx context.Context, username string) (*Document, error)
	Search(ctx context.Context, filters *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, username string, timestamp int64) error
}

type SessionManager interface {
	Create(ctx context.Context, user *Model) (*Session, error)
	Update(ctx context.Context, user *Model) (*Session, error)
	Revoke(ctx context.Context, username string, suid string) error
	RevokeID(ctx context.Context, suid string) error
	RevokeUser(ctx context.Context, username string) error
	Validate(ctx context.Context, suid string) (*Session, error)
}

func (Model) TableName() string {
	return "users"
}

func (um *Model) ToDoc() Document {
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

func (ud *Document) ToModel() Model {
	var um Model
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
