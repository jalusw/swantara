package helper

func Ptr[T any](value T) *T {
	return &value
}

func StringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func Deref[T any](value *T, fallback T) T {
	if value == nil {
		return fallback
	}
	return *value
}
