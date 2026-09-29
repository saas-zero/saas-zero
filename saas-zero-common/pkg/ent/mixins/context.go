package mixins

import "context"

type ctxKey string

const (
	KeyUserId   ctxKey = "current_user_id"
	KeyUserName ctxKey = "current_user_name"
	KeyTenantId ctxKey = "current_tenant_id"
)

func SetCurrentUserId(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, KeyUserId, id)
}

func GetCurrentUserId(ctx context.Context) int64 {
	if v, ok := ctx.Value(KeyUserId).(int64); ok {
		return v
	}
	return 0
}

func SetCurrentUserName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, KeyUserName, name)
}

func GetCurrentUserName(ctx context.Context) string {
	if v, ok := ctx.Value(KeyUserName).(string); ok {
		return v
	}
	return ""
}

func SetCurrentTenantId(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, KeyTenantId, id)
}

func GetCurrentTenantId(ctx context.Context) int64 {
	if v, ok := ctx.Value(KeyTenantId).(int64); ok {
		return v
	}
	return 0
}
