package database

import (
	"fmt"
	"gocms/internal/domain"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds the parameters required to connect to the database
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func NewMySQLConnection(cfg *DBConfig, level logger.LogLevel) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(level),
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func MapSQLError(err *driver.MySQLError) error {
	switch err.Number {
	case 1761, 1062:
		// 1761: Foreign key constraint update leading to primary key duplicate
		// 1062: Direct duplicate entry for primary/unique key
		return domain.ErrDuplicate
	default:
		return err
	}
}
