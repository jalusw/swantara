package inventory

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"gorm.io/gorm"
)

type Poster interface {
	Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error)
	PostTx(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error)
}
