package logic

import (
	"context"
	"testing"
	"time"

	"github.com/saas-zero/saas-zero-common/pkg/id"
	"github.com/saas-zero/saas-zero-common/pkg/jwt"
	"google.golang.org/grpc/metadata"
)

const testAuthSecret = "auth-unit-test-secret-2024"

func TestExtractBearerToken(t *testing.T) {
	cases := []struct {
		header string
		want   string
	}{
		{"Bearer abc123.def", "abc123.def"},
		{"bearer abc", "abc"},
		{"", ""},
		{"Token abc", ""},
		{"Bearer", ""},
	}
	for _, c := range cases {
		if got := ExtractBearerToken(c.header); got != c.want {
			t.Fatalf("ExtractBearerToken(%q) = %q, want %q", c.header, got, c.want)
		}
	}
}

func TestWithToken_GetToken(t *testing.T) {
	ctx := WithToken(context.Background(), "my-token")
	if got := GetToken(ctx); got != "my-token" {
		t.Fatalf("GetToken = %q, want my-token", got)
	}
	if got := GetToken(context.Background()); got != "" {
		t.Fatalf("GetToken on empty ctx should be empty, got %q", got)
	}
}

func TestWithAuthContext_InjectMetadata(t *testing.T) {
	claims := &jwt.Claims{UserId: 1001, TenantId: 2001, UserName: "admin"}
	token, err := jwt.Sign(testAuthSecret, claims, time.Hour)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	ctx := WithToken(context.Background(), token)
	out := withAuthContext(ctx, testAuthSecret)

	md, ok := metadata.FromOutgoingContext(out)
	if !ok {
		t.Fatal("expected outgoing gRPC metadata")
	}
	if got := md.Get("x-user-id"); len(got) != 1 || got[0] != id.ToString(1001) {
		t.Fatalf("unexpected x-user-id: %v", got)
	}
	if got := md.Get("x-tenant-id"); len(got) != 1 || got[0] != id.ToString(2001) {
		t.Fatalf("unexpected x-tenant-id: %v", got)
	}
	if got := md.Get("x-user-name"); len(got) != 1 || got[0] != "admin" {
		t.Fatalf("unexpected x-user-name: %v", got)
	}
}

func TestWithAuthContext_NoToken(t *testing.T) {
	out := withAuthContext(context.Background(), testAuthSecret)
	if _, ok := metadata.FromOutgoingContext(out); ok {
		t.Fatal("expected no metadata when token absent")
	}
}

func TestWithAuthContext_InvalidToken(t *testing.T) {
	ctx := WithToken(context.Background(), "garbage-token")
	out := withAuthContext(ctx, testAuthSecret)
	if _, ok := metadata.FromOutgoingContext(out); ok {
		t.Fatal("expected no metadata when token invalid")
	}
}
