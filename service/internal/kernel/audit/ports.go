package audit

import (
	"context"
)

type Recorder interface {
	Append(ctx context.Context, entry *Log) error
}
