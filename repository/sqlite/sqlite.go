package sqlite

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/slaveofcode/hansip/repository"
	"github.com/slaveofcode/hansip/repository/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewRepository(opt *ConnectionOption, timezone *time.Location) repository.Repository {
	return &RepositorySqlite{
		connOption: opt,
		tz:         timezone,
	}
}

type ConnectionOption struct {
	Path string
}

type RepositorySqlite struct {
	connOption *ConnectionOption
	db         *gorm.DB
	tz         *time.Location
}

func (sq *RepositorySqlite) Migrate() error {
	if sq.db == nil {
		return fmt.Errorf("database doesn't connected yet")
	}

	// Activates WAL mode
	sq.db.Exec(`PRAGMA journal_mode=WAL;`)

	err := sq.db.AutoMigrate(
		&models.User{},
		&models.UserCredential{},
		&models.UserKey{},
		&models.AccessToken{},
		&models.FileGroup{},
		&models.FileItem{},
		&models.FileGroupUser{},
		&models.ShortLink{},
	)

	if err != nil {
		return err
	}

	log.Println("database migrated")
	return nil
}

func (sq *RepositorySqlite) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf("file:%s", sq.connOption.Path)

	operationTimezone := time.UTC
	if sq.tz != nil {
		operationTimezone = sq.tz
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		// disable transaction that enable by default on any operation
		// see: https://gorm.io/docs/transactions.html#Disable-Default-Transaction
		SkipDefaultTransaction: true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
		NowFunc: func() time.Time {
			return time.Now().In(operationTimezone)
		},
	})

	if err != nil {
		return fmt.Errorf("error connecting database: [%s] %v", dsn, err)
	}

	sq.db = db
	return nil
}
func (sq *RepositorySqlite) Close() error {
	if sq.db == nil {
		return nil
	}

	sqlDB, err := sq.db.DB()
	if err != nil {
		return fmt.Errorf("error getting db connection instance: %v", err)
	}

	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("error closing db connection: %v", err)
	}

	return nil
}

func (sq *RepositorySqlite) GetDB() *gorm.DB {
	return sq.db
}
