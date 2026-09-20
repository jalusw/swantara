package db

import (
	"context"

	"gorm.io/gorm"
)

type Transactioner interface {
	Run(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type DBTransactioner struct {
	database *gorm.DB
}

func NewDBTransactioner(database *gorm.DB) DBTransactioner {
	return DBTransactioner{database: database}
}

func (t DBTransactioner) Run(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return t.database.WithContext(ctx).Transaction(fn)
}
