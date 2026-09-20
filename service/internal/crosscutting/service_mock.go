package crosscutting

import (
	"context"
	"io"
	"strings"

	"gorm.io/gorm"
)

type TransactionerMock struct {
	RunFunc func(ctx context.Context, fn func(tx *gorm.DB) error) error
}

func (m TransactionerMock) Run(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, fn)
	}
	return fn(nil)
}

type StoreMock struct {
	SaveFunc   func(ctx context.Context, id string, reader io.Reader) error
	OpenFunc   func(ctx context.Context, id string) (io.ReadCloser, error)
	DeleteFunc func(ctx context.Context, id string) error
}

func (m StoreMock) Save(ctx context.Context, id string, reader io.Reader) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, id, reader)
	}
	return nil
}

func (m StoreMock) Open(ctx context.Context, id string) (io.ReadCloser, error) {
	if m.OpenFunc != nil {
		return m.OpenFunc(ctx, id)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m StoreMock) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}
