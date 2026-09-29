package logic

import (
	"context"
	"fmt"
	"github.com/saas-zero/saas-zero-common/pkg/id"
	"strings"

	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/jwt"
	"github.com/saas-zero/saas-zero-common/pkg/redis"
	"google.golang.org/grpc/metadata"
)

type ctxKey string

const tokenKey ctxKey = "auth_token"

func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

func GetToken(ctx context.Context) string {
	if v, ok := ctx.Value(tokenKey).(string); ok {
		return v
	}
	return ""
}

func ExtractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1]
	}
	return ""
}

func tokenExistsInRedis(rds *redis.Client, jti string) bool {
	if jti == "" {
		return false
	}
	key := fmt.Sprintf("token:%s", jti)
	exists, err := rds.Exists(key)
	return err == nil && exists
}

// tokenVersionMatches verifies the JWT tokenVersion equals the current
// per-user version in Redis. Passwords, roles and permissions changes bump
// this value so that stale sessions cannot be refreshed into new privileges.
func tokenVersionMatches(rds *redis.Client, claims *jwt.Claims) bool {
	if claims == nil {
		return false
	}
	key := fmt.Sprintf("token_version:%d", claims.UserId)
	cur, err := rds.Get(key)
	if err != nil || cur == "" || cur != fmt.Sprintf("%d", claims.TokenVersion) {
		return false
	}
	return true
}

// validateSession is the unified Auth validation chain: JWT signature +
// expiry, Redis JTI existence and per-user tokenVersion. It is shared by
// /oauth/userinfo, /oauth/menus, /oauth/permissions, /oauth/refresh,
// password change and password reset so a revoked token is rejected the same
// way everywhere.
func validateSession(rds *redis.Client, secret, token string) (*jwt.Claims, *errno.Errno) {
	if token == "" {
		return nil, errno.TokenExpired
	}
	claims, err := jwt.Parse(token, secret)
	if err != nil {
		return nil, errno.TokenExpired
	}
	if !tokenExistsInRedis(rds, claims.ID) {
		return nil, errno.TokenInvalidated
	}
	if !tokenVersionMatches(rds, claims) {
		return nil, errno.TokenVersionMismatch
	}
	return claims, nil
}

func withAuthContext(ctx context.Context, secret string) context.Context {
	token := GetToken(ctx)
	if token == "" {
		return ctx
	}
	claims, err := jwt.Parse(token, secret)
	if err != nil {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"x-user-id", id.ToString(claims.UserId),
		"x-user-name", claims.UserName,
		"x-tenant-id", id.ToString(claims.TenantId),
	))
}
