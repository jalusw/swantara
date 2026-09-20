package reference

import "context"

type variantExistsStub struct {
	exists bool
}

func (s variantExistsStub) VariantExists(ctx context.Context, id uint64) (bool, error) {
	return s.exists, nil
}
