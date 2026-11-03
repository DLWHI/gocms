package test

import (
	"context"
	"gocms/internal/category"
	"gocms/internal/domain"
	"gocms/internal/editor"
	"gocms/internal/lang"
	"gocms/internal/user"
	"time"

	"gorm.io/gorm"
)

type Fixtures struct {
	DefaultPassword Password          `json:"default_password"`
	Actors          map[string]*Actor `json:"actors"`
	Users           []domain.User     `json:"users"`
	Languages       []lang.Language   `json:"languages"`
	Categories      []domain.Category `json:"categories"`
	Editors         []domain.Editor   `json:"editors"`
}

func (f *Fixtures) ApplyFixtures(ctx context.Context, e *Environment) error {
	db := e.DB
	ms := e.Search
	hash, err := e.Password.Hash([]byte(f.DefaultPassword))
	if err != nil {
		return err
	}

	for _, actor := range f.Actors {
		if err := db.WithContext(ctx).Save(actor.ToRecord(hash)).Error; err != nil {
			return err
		}
		task, err := ms.UsersIndex.AddDocumentsWithContext(ctx, actor.ToDomain(), nil)
		if err != nil {
			return err
		}
		if err := ms.WaitForTask(ctx, task); err != nil {
			return err
		}

		if err := actor.PerformLogin(ctx, e, f.DefaultPassword); err != nil {
			return err
		}
	}

	if err := e.Outbox.Dispatch(ctx, len(f.Actors)); err != nil {
		return err
	}

	for _, u := range f.Users {
		record := user.ToRecord(&u)
		record.Password = hash
		if u.SuspendedAt != nil {
			record.SuspendedAt = gorm.DeletedAt{
				Valid: true,
				Time:  time.Unix(*u.SuspendedAt, 0),
			}
		}
		if u.LastLogin != nil {
			t := time.Unix(*u.LastLogin, 0)
			record.LastLogin = &t
		}
		if err := db.WithContext(ctx).Save(record).Error; err != nil {
			return err
		}
		task, err := ms.UsersIndex.AddDocumentsWithContext(ctx, user.ToSearchDocument(&u), nil)
		if err != nil {
			return err
		}
		if err := ms.WaitForTask(ctx, task); err != nil {
			return err
		}
	}

	for _, l := range f.Languages {
		if err := e.Lang.Add(ctx, &l); err != nil {
			return err
		}
	}

	for _, c := range f.Categories {
		record := category.ToRecord(&c)
		if err := db.WithContext(ctx).Save(record).Error; err != nil {
			return err
		}

		task, err := ms.CategoriesIndex.AddDocumentsWithContext(ctx, category.ToSearchDocument(&c), nil)
		if err != nil {
			return err
		}
		if err := ms.WaitForTask(ctx, task); err != nil {
			return err
		}
	}

	for _, e := range f.Editors {
		record := editor.ToRecord(&e)
		if err := db.WithContext(ctx).Save(record).Error; err != nil {
			return err
		}

		task, err := ms.EditorsIndex.AddDocumentsWithContext(ctx, editor.ToSearchDocument(&e), nil)
		if err != nil {
			return err
		}
		if err := ms.WaitForTask(ctx, task); err != nil {
			return err
		}
	}

	return nil
}
