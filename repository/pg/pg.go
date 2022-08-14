package pg

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/slaveofcode/securi/repository"
	"github.com/slaveofcode/securi/repository/pg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewRepository(opt *ConnectionOption, timezone *time.Location) repository.Repository {
	return &RepositoryPostgres{
		connOption: opt,
		tz:         timezone,
	}
}

type ConnectionOption struct {
	DBName string
	Host   string
	Port   string
	User   string
	Pass   string
}

type RepositoryPostgres struct {
	connOption *ConnectionOption
	db         *gorm.DB
	tz         *time.Location
}

func (pg *RepositoryPostgres) Migrate() error {
	if pg.db == nil {
		return fmt.Errorf("database doesn't connected yet")
	}

	err := pg.db.AutoMigrate(
		&models.User{},
		&models.UserCredential{},
		&models.UserKey{},
		&models.AccessToken{},
		&models.FileGroup{},
		&models.FileItem{},
		&models.FileGroupSignature{},
		&models.ShortLink{},
	)

	if err != nil {
		return err
	}

	log.Println("database migrated")
	return nil
}

func (pg *RepositoryPostgres) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		pg.connOption.Host, pg.connOption.User, pg.connOption.Pass, pg.connOption.DBName, pg.connOption.Port)

	operationTimezone := time.UTC
	if pg.tz != nil {
		operationTimezone = pg.tz
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
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

	pg.db = db

	return nil
}

func (pg *RepositoryPostgres) Close() error {
	if pg.db == nil {
		return nil
	}

	sqlDB, err := pg.db.DB()
	if err != nil {
		return fmt.Errorf("error getting db connection instance: %v", err)
	}

	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("error closing db connection: %v", err)
	}

	return nil
}

func (pg *RepositoryPostgres) GetDB() *gorm.DB {
	return pg.db
}
