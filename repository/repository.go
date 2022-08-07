package repository

import "context"

type Repository interface {
	Connect(ctx context.Context) error
	Close() error
}
