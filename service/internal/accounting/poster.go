package accounting

import (
	"context"

	"gorm.io/gorm"
)

type Poster interface {
	Post(ctx context.Context, request PostRequest) (*JournalEntry, error)
	PostTx(ctx context.Context, tx *gorm.DB, request PostRequest) (*JournalEntry, error)
	Reverse(ctx context.Context, request ReverseRequest) (*JournalEntry, error)
	ReverseTx(ctx context.Context, tx *gorm.DB, request ReverseRequest) (*JournalEntry, error)
}
