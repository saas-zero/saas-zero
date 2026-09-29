package jwt

import (
	"testing"
	"time"
)

const testSecret = "test-secret-key-for-unit-testing"

func TestSignAndParse_Success(t *testing.T) {
	claims := &Claims{UserId: 1001, TenantId: 2001, UserName: "testuser"}
	token, err := Sign(testSecret, claims, time.Hour)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	parsed, err := Parse(token, testSecret)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if parsed.UserId != 1001 {
		t.Fatalf("expected UserId=1001, got %d", parsed.UserId)
	}
	if parsed.TenantId != 2001 {
		t.Fatalf("expected TenantId=2001, got %d", parsed.TenantId)
	}
	if parsed.UserName != "testuser" {
		t.Fatalf("expected UserName=testuser, got %s", parsed.UserName)
	}
	if parsed.ID == "" {
		t.Fatal("expected non-empty JTI")
	}
}

func TestParse_WrongSecret(t *testing.T) {
	claims := &Claims{UserId: 1, TenantId: 1, UserName: "test"}
	token, err := Sign(testSecret, claims, time.Hour)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	_, err = Parse(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error when parsing with wrong secret")
	}
}

func TestParse_ExpiredToken(t *testing.T) {
	claims := &Claims{UserId: 1, TenantId: 1, UserName: "test"}
	token, err := Sign(testSecret, claims, -time.Hour)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	_, err = Parse(token, testSecret)
	if err == nil {
		t.Fatal("expected error when parsing expired token")
	}
}

func TestParse_MalformedToken(t *testing.T) {
	_, err := Parse("not-a-valid-token", testSecret)
	if err == nil {
		t.Fatal("expected error when parsing malformed token")
	}
}

func TestParse_EmptyToken(t *testing.T) {
	_, err := Parse("", testSecret)
	if err == nil {
		t.Fatal("expected error when parsing empty token")
	}
}

func TestParse_TokenNotExpired(t *testing.T) {
	claims := &Claims{UserId: 1, TenantId: 1, UserName: "fresh"}
	token, err := Sign(testSecret, claims, time.Hour)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	parsed, err := Parse(token, testSecret)
	if err != nil {
		t.Fatalf("Parse failed for valid token: %v", err)
	}
	_ = parsed
}

// TestSignAndParse_RoleCodesAndTokenVersion 验证 roleCodes 与 tokenVersion
// 往返无损（/oauth/refresh 重签与权限踢出依赖这两个字段）。
func TestSignAndParse_RoleCodesAndTokenVersion(t *testing.T) {
	claims := &Claims{
		UserId:       1001,
		TenantId:     2001,
		UserName:     "testuser",
		RoleCodes:    []string{"admin", "user"},
		TokenVersion: 7,
	}
	token, err := Sign(testSecret, claims, time.Hour)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	parsed, err := Parse(token, testSecret)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if parsed.TokenVersion != 7 {
		t.Fatalf("expected TokenVersion=7, got %d", parsed.TokenVersion)
	}
	if len(parsed.RoleCodes) != 2 || parsed.RoleCodes[0] != "admin" || parsed.RoleCodes[1] != "user" {
		t.Fatalf("expected roleCodes [admin user], got %v", parsed.RoleCodes)
	}
}

// TestSignAndParse_RoleCodesEmpty 验证无角色时解析结果 len==0（nil 也安全，
// 下游 casbinauth 通过 len==0 判为无权限，fail-closed）。
func TestSignAndParse_RoleCodesEmpty(t *testing.T) {
	claims := &Claims{UserId: 1, TenantId: 1, UserName: "noroles"}
	token, err := Sign(testSecret, claims, time.Hour)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	parsed, err := Parse(token, testSecret)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(parsed.RoleCodes) != 0 {
		t.Fatalf("expected 0 role codes, got %d", len(parsed.RoleCodes))
	}
}
