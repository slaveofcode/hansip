package repository

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Connect(ctx context.Context) error
	Migrate() error
	Close() error

	GetDB() *gorm.DB
}
