package bcrypt

import (
	"testing"
)

func TestHash_ReturnsValidHash(t *testing.T) {
	hash, err := Hash("password123")
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	if len(hash) == 0 {
		t.Fatal("expected non-empty hash")
	}
}

func TestVerify_CorrectPassword(t *testing.T) {
	password := "test-password-xyz"
	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	if !Verify(password, hash) {
		t.Fatal("Verify returned false for correct password")
	}
}

func TestVerify_WrongPassword(t *testing.T) {
	hash, err := Hash("correct-password")
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	if Verify("wrong-password", hash) {
		t.Fatal("Verify returned true for wrong password")
	}
}

func TestVerify_EmptyPassword(t *testing.T) {
	hash, err := Hash("somepassword")
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	if Verify("", hash) {
		t.Fatal("Verify returned true for empty password")
	}
}

func TestHash_DifferentSalts(t *testing.T) {
	h1, err := Hash("samepassword")
	if err != nil {
		t.Fatalf("first Hash failed: %v", err)
	}
	h2, err := Hash("samepassword")
	if err != nil {
		t.Fatalf("second Hash failed: %v", err)
	}
	if h1 == h2 {
		t.Fatal("expected different hashes due to random salt")
	}
}

func TestVerify_InvalidHash(t *testing.T) {
	if Verify("password", "not-a-valid-hash") {
		t.Fatal("Verify returned true for invalid hash")
	}
}

func TestHash_EmptyString(t *testing.T) {
	hash, err := Hash("")
	if err != nil {
		t.Fatalf("Hash failed for empty string: %v", err)
	}
	if !Verify("", hash) {
		t.Fatal("Verify returned false for empty string")
	}
}

func TestHash_SpecialCharacters(t *testing.T) {
	password := "密码!@#$%^&*()_+-=[]{}|;:',.<>?/~`"
	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash failed for special chars: %v", err)
	}
	if !Verify(password, hash) {
		t.Fatal("Verify failed for special chars password")
	}
}
