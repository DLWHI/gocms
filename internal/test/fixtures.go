package test

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/user"
)

type Fixtures struct {
	DefaultPassword Password          `json:"default_password"`
	Actors          map[string]*Actor `json:"actors"`
	Users           []domain.User     `json:"users"`
}

func (f *Fixtures) ApplyFixtures(ctx context.Context, e *Environment) error {
	db := e.DB
	ms := e.Search
	hash, err := e.Password.Hash([]byte(f.DefaultPassword))
	if err != nil {
		return err
	}

	for _, actor := range f.Actors {
		if actor.Permissions == 0 {
			actor.Permissions = GetActorPermissions(domain.UserEntityName, actor)
		}
		record := actor.ToRecord(hash)
		if err := db.WithContext(ctx).Save(&record).Error; err != nil {
			return err
		}
		udomain := actor.ToDomain()
		task, err := ms.UsersIndex.AddDocumentsWithContext(ctx, &udomain, nil)
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

	for _, u := range f.Users {
		record := user.ToRecord(&u)
		record.Password = hash
		if err := db.WithContext(ctx).Save(&record).Error; err != nil {
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

	return nil
}

func (f *Fixtures) ResetFixtures(ctx context.Context, e *Environment) error {
	db := e.DB
	ms := e.Search

	for _, actor := range f.Actors {
		actor.Session = ""
		err := db.WithContext(ctx).
			Select("permissions", "email").
			Where("username = ?", actor.Username).
			Save(actor.ToRecord(nil)).Error
		if err != nil {
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

	for _, u := range f.Users {
		err := db.WithContext(ctx).
			Select("permissions", "email").
			Where("username = ?", u.Username).
			Updates(user.ToRecord(&u)).Error
		if err != nil {
			return err
		}
		task, err := ms.UsersIndex.AddDocumentsWithContext(ctx, u, nil)
		if err != nil {
			return err
		}
		if err := ms.WaitForTask(ctx, task); err != nil {
			return err
		}
	}

	return nil
}
