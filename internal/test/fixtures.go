package test

import (
	"context"
	"gocms/internal/category"
	"gocms/internal/domain"
	"gocms/internal/editor"
	"gocms/internal/lang"
	"gocms/internal/user"
)

type Fixtures struct {
	DefaultPassword  Password          `json:"default_password"`
	Actors           map[string]*Actor `json:"actors"`
	Users            []domain.User     `json:"users"`
	Languages        []lang.Language   `json:"languages"`
	FallbackLanguage string            `json:"fallback_language"`
	Categories       []domain.Category `json:"categories"`
	Editors          []domain.Editor   `json:"editors"`
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
		if err := db.WithContext(ctx).Save(record).Error; err != nil {
			return err
		}
		task, err := ms.UsersIndex.AddDocumentsWithContext(ctx, &u, nil)
		if err != nil {
			return err
		}
		if err := ms.WaitForTask(ctx, task); err != nil {
			return err
		}
	}

	for _, l := range f.Languages {
		if err := e.DB.WithContext(ctx).Save(&l).Error; err != nil {
			return err
		}
	}

	if err := e.Lang.Reload(ctx, nil); err != nil {
		return err
	}

	for _, c := range f.Categories {
		record := category.ToRecord(&c)
		if err := db.WithContext(ctx).Save(record).Error; err != nil {
			return err
		}

		task, err := ms.CategoriesIndex.AddDocumentsWithContext(ctx, &c, nil)
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

		task, err := ms.EditorsIndex.AddDocumentsWithContext(ctx, &e, nil)
		if err != nil {
			return err
		}
		if err := ms.WaitForTask(ctx, task); err != nil {
			return err
		}
	}

	return nil
}
