package test

import (
	"context"
	"fmt"
	"gocms/internal/advertisement"
	"gocms/internal/category"
	"gocms/internal/domain"
	"gocms/internal/editor"
	"gocms/internal/lang"
	"gocms/internal/post"
	"gocms/internal/user"
)

type Fixtures struct {
	DefaultPassword  Password               `json:"default_password"`
	Actors           map[string]*Actor      `json:"actors"`
	Users            []domain.User          `json:"users"`
	Languages        []lang.Language        `json:"languages"`
	FallbackLanguage string                 `json:"fallback_language"`
	Categories       []domain.Category      `json:"categories"`
	Editors          []domain.Editor        `json:"editors"`
	Posts            []domain.Post          `json:"posts"`
	Ads              []domain.Advertisement `json:"ads"`
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

	for _, usr := range f.Users {
		record := user.ToRecord(&usr)
		record.Password = hash
		if err := db.WithContext(ctx).Save(record).Error; err != nil {
			return err
		}
		task, err := ms.UsersIndex.AddDocumentsWithContext(ctx, &usr, nil)
		if err != nil {
			return err
		}
		if err := ms.WaitForTask(ctx, task); err != nil {
			return err
		}
	}

	for _, language := range f.Languages {
		if err := e.DB.WithContext(ctx).Save(&language).Error; err != nil {
			return err
		}
	}

	if err := e.Langs.Reload(ctx, nil); err != nil {
		return err
	}

	for _, ctg := range f.Categories {
		record := category.ToRecord(&ctg)
		if err := db.WithContext(ctx).Save(record).Error; err != nil {
			return err
		}

		ctg.ID = fmt.Sprintf("%s-%s", ctg.Slug, ctg.Lang)
		task, err := ms.CategoriesIndex.AddDocumentsWithContext(ctx, &ctg, nil)
		if err != nil {
			return err
		}
		if err := ms.WaitForTask(ctx, task); err != nil {
			return err
		}
	}

	for _, ed := range f.Editors {
		record := editor.ToRecord(&ed)
		if err := db.WithContext(ctx).Save(record).Error; err != nil {
			return err
		}

		ed.ID = fmt.Sprintf("%s-%s", ed.Slug, ed.Lang)
		task, err := ms.EditorsIndex.AddDocumentsWithContext(ctx, &ed, nil)
		if err != nil {
			return err
		}
		if err := ms.WaitForTask(ctx, task); err != nil {
			return err
		}
	}

	for _, pt := range f.Posts {
		record := post.ToRecord(&pt)
		if err := db.WithContext(ctx).Save(record).Error; err != nil {
			return err
		}

		pt.ID = fmt.Sprintf("%d-%s", pt.PostID, pt.Lang)
		task, err := ms.PostsIndex.AddDocumentsWithContext(ctx, &pt, nil)
		if err != nil {
			return err
		}
		if err := ms.WaitForTask(ctx, task); err != nil {
			return err
		}
	}

	for _, ad := range f.Ads {
		record := advertisement.ToRecord(&ad)
		if err := db.WithContext(ctx).Save(record).Error; err != nil {
			return err
		}

		task, err := ms.AdsIndex.AddDocumentsWithContext(ctx, &ad, nil)
		if err != nil {
			return err
		}
		if err := ms.WaitForTask(ctx, task); err != nil {
			return err
		}
	}

	return nil
}
