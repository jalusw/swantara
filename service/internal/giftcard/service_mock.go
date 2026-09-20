package giftcard

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"gorm.io/gorm"
)

type PosterMock struct {
	PostFunc      func(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error)
	PostTxFunc    func(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error)
	ReverseFunc   func(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
	ReverseTxFunc func(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
}

func (m PosterMock) Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostFunc != nil {
		return m.PostFunc(ctx, request)
	}
	return &accounting.JournalEntry{}, nil
}

func (m PosterMock) PostTx(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostTxFunc != nil {
		return m.PostTxFunc(ctx, tx, request)
	}
	return m.Post(ctx, request)
}

func (m PosterMock) Reverse(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	if m.ReverseFunc != nil {
		return m.ReverseFunc(ctx, request)
	}
	return &accounting.JournalEntry{}, nil
}

func (m PosterMock) ReverseTx(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	if m.ReverseTxFunc != nil {
		return m.ReverseTxFunc(ctx, tx, request)
	}
	return m.Reverse(ctx, request)
}

type TransactionerMock struct {
	RunFunc func(ctx context.Context, fn func(tx *gorm.DB) error) error
}

func (m TransactionerMock) Run(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, fn)
	}
	return fn(nil)
}
