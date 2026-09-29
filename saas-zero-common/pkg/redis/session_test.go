package redis

import (
	"errors"
	"testing"
)

type stubStore struct {
	keys       []string
	err        error
	failOnCall int // 第 N 次调用失败（0 表示全部成功）
}

func (s *stubStore) Incr(key string) (int64, error) {
	s.keys = append(s.keys, key)
	if s.err != nil && s.failOnCall == len(s.keys) {
		return 0, s.err
	}
	return 1, nil
}

func TestBumpTokenVersion_Success(t *testing.T) {
	store := &stubStore{}
	if err := BumpTokenVersion(store, 1001); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(store.keys) != 1 || store.keys[0] != "token_version:1001" {
		t.Fatalf("unexpected keys: %v", store.keys)
	}
}

func TestBumpTokenVersion_StoreError(t *testing.T) {
	store := &stubStore{err: errors.New("connection refused"), failOnCall: 1}
	err := BumpTokenVersion(store, 1001)
	if err == nil {
		t.Fatal("redis failure must not be swallowed")
	}
	if !errors.Is(err, store.err) {
		t.Fatalf("expected wrapped store error, got %v", err)
	}
}

func TestBumpTokenVersion_NilStore(t *testing.T) {
	if err := BumpTokenVersion(nil, 1001); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("expected ErrNotInitialized, got %v", err)
	}
}

// 单元测试中 svcCtx.Redis 常为 nil 的 *Client，必须返回错误而不是 panic。
func TestBumpTokenVersion_NilClient(t *testing.T) {
	var client *Client
	if err := BumpTokenVersion(client, 1001); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("expected ErrNotInitialized, got %v", err)
	}
}

func TestBumpTokenVersions_AllSuccess(t *testing.T) {
	store := &stubStore{}
	if err := BumpTokenVersions(store, 1, 2, 3); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(store.keys) != 3 {
		t.Fatalf("expected 3 keys, got %v", store.keys)
	}
}

func TestBumpTokenVersions_StopsAtFirstError(t *testing.T) {
	store := &stubStore{err: errors.New("connection refused"), failOnCall: 2}
	err := BumpTokenVersions(store, 1, 2, 3)
	if err == nil {
		t.Fatal("expected error from second user")
	}
	if len(store.keys) != 2 {
		t.Fatalf("must stop at first failure, got keys %v", store.keys)
	}
}

func TestBumpTokenVersions_Empty(t *testing.T) {
	store := &stubStore{}
	if err := BumpTokenVersions(store); err != nil {
		t.Fatalf("expected success for empty list, got %v", err)
	}
	if len(store.keys) != 0 {
		t.Fatalf("expected no calls, got %v", store.keys)
	}
}
