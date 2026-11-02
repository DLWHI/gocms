package database

import (
	"errors"
	"fmt"
	"gocms/internal/domain"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func MapError(err error) error {
	var mysqlErr *mysql.MySQLError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &mysqlErr):
		return MapSQLError(mysqlErr)
	default:
		return MapGORMError(mysqlErr)
	}
}

func MapGORMError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return domain.ErrDuplicate
	case errors.Is(err, gorm.ErrForeignKeyViolated), errors.Is(err, gorm.ErrCheckConstraintViolated), errors.Is(err, gorm.ErrInvalidValue):
		return fmt.Errorf("%w: %v", domain.ErrInval, err)
	case errors.Is(err, gorm.ErrRecordNotFound):
		return domain.ErrNotFound
	default:
		return err
	}
}
