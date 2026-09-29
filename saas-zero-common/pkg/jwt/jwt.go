package jwt

import (
	"errors"
	"github.com/google/uuid"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserId       int64    `json:"userId"`
	TenantId     int64    `json:"tenantId"`
	UserName     string   `json:"userName"`
	RoleCodes    []string `json:"roleCodes"`
	TokenVersion int64    `json:"tokenVersion"`
	gojwt.RegisteredClaims
}

func Sign(secret string, claims *Claims, expire time.Duration) (string, error) {
	now := time.Now()
	if claims.ID == "" {
		claims.ID = uuid.New().String()
	}
	claims.RegisteredClaims = gojwt.RegisteredClaims{
		ExpiresAt: gojwt.NewNumericDate(now.Add(expire)),
		IssuedAt:  gojwt.NewNumericDate(now),
		ID:        claims.ID,
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Parse(tokenString, secret string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenString, &Claims{}, func(token *gojwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
