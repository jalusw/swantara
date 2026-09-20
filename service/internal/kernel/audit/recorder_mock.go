package audit

import (
	"context"
)

type recorderMock struct {
	entries []*Log
}

func (r *recorderMock) Append(_ context.Context, entry *Log) error {
	r.entries = append(r.entries, entry)
	return nil
}
