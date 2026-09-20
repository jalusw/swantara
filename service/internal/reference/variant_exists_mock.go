package reference

import "context"

type VariantExistenceMock struct {
	VariantExistsFunc func(ctx context.Context, id uint64) (bool, error)
}

func (m VariantExistenceMock) VariantExists(ctx context.Context, id uint64) (bool, error) {
	if m.VariantExistsFunc != nil {
		return m.VariantExistsFunc(ctx, id)
	}
	return true, nil
}
