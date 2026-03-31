package appctx

import "context"

type contextKey string

const (
	RequestIDKey   contextKey = "request_id"
	AdminUserIDKey contextKey = "admin_user_id"
)

func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(RequestIDKey).(string)
	return value
}

func AdminUserID(ctx context.Context) int64 {
	value, _ := ctx.Value(AdminUserIDKey).(int64)
	return value
}
