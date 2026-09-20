package helper

import "context"

type requestIDKey struct{}

var RequestIDKey requestIDKey

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(RequestIDKey).(string)
	return id
}
