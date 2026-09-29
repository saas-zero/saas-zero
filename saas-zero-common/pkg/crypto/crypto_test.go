package crypto

import (
	"encoding/hex"
	"strings"
	"testing"
)

func setRawKey(t *testing.T, key string) {
	t.Helper()
	k := []byte(key)
	if len(k) > 32 {
		k = k[:32]
	}
	for len(k) < 32 {
		k = append(k, '0')
	}
	defaultKey = k
}

func setHexKey(t *testing.T, hexKey string) {
	t.Helper()
	k, err := hex.DecodeString(hexKey)
	if err != nil {
		t.Fatalf("invalid hex key: %v", err)
	}
	defaultKey = k
}

func clearKey(t *testing.T) {
	t.Helper()
	defaultKey = nil
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	setRawKey(t, "my-32-byte-encryption-key-1234567")
	defer clearKey(t)

	plaintext := "sensitive-data-123"
	encrypted, err := EncryptString(plaintext)
	if err != nil {
		t.Fatalf("EncryptString failed: %v", err)
	}

	if encrypted == plaintext {
		t.Fatal("encrypted text should not equal plaintext")
	}

	decrypted, err := DecryptString(encrypted)
	if err != nil {
		t.Fatalf("DecryptString failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptDecrypt_Chinese(t *testing.T) {
	setRawKey(t, "my-32-byte-encryption-key-1234567")
	defer clearKey(t)

	plaintext := "手机号码13800138000"
	encrypted, err := EncryptString(plaintext)
	if err != nil {
		t.Fatalf("EncryptString failed: %v", err)
	}

	decrypted, err := DecryptString(encrypted)
	if err != nil {
		t.Fatalf("DecryptString failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptDecrypt_EmptyString(t *testing.T) {
	setRawKey(t, "my-32-byte-encryption-key-1234567")
	defer clearKey(t)

	encrypted, err := EncryptString("")
	if err != nil {
		t.Fatalf("EncryptString failed: %v", err)
	}

	decrypted, err := DecryptString(encrypted)
	if err != nil {
		t.Fatalf("DecryptString failed: %v", err)
	}

	if decrypted != "" {
		t.Fatalf("expected empty string, got %q", decrypted)
	}
}

func TestEncrypt_UniqueNonce(t *testing.T) {
	setRawKey(t, "my-32-byte-encryption-key-1234567")
	defer clearKey(t)

	plaintext := "same-data"
	c1, err := EncryptString(plaintext)
	if err != nil {
		t.Fatalf("first EncryptString failed: %v", err)
	}
	c2, err := EncryptString(plaintext)
	if err != nil {
		t.Fatalf("second EncryptString failed: %v", err)
	}

	if c1 == c2 {
		t.Fatal("expected different ciphertexts due to random nonce")
	}
}

func TestEncrypt_MissingKey(t *testing.T) {
	clearKey(t)
	defer clearKey(t)

	_, err := EncryptString("test")
	if err == nil {
		t.Fatal("expected error when ENCRYPT_KEY is not set")
	}
}

func TestDecrypt_MissingKey(t *testing.T) {
	clearKey(t)
	defer clearKey(t)

	_, err := DecryptString("dGVzdA==")
	if err == nil {
		t.Fatal("expected error when ENCRYPT_KEY is not set")
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	setRawKey(t, "my-32-byte-encryption-key-1234567")
	defer clearKey(t)

	_, err := DecryptString("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64 input")
	}
}

func TestDecrypt_CiphertextTooShort(t *testing.T) {
	setRawKey(t, "my-32-byte-encryption-key-1234567")
	defer clearKey(t)

	_, err := DecryptString("AAECAw==")
	if err == nil {
		t.Fatal("expected error for too-short ciphertext")
	}
}

func TestKey_32ByteHex(t *testing.T) {
	setHexKey(t, strings.Repeat("ab", 32))
	defer clearKey(t)

	if len(defaultKey) != 32 {
		t.Fatalf("expected key length 32, got %d", len(defaultKey))
	}

	enc, err := EncryptString("test")
	if err != nil {
		t.Fatalf("EncryptString with hex key failed: %v", err)
	}
	dec, err := DecryptString(enc)
	if err != nil {
		t.Fatalf("DecryptString with hex key failed: %v", err)
	}
	if dec != "test" {
		t.Fatalf("expected test, got %q", dec)
	}
}

func TestKey_ShortKeyPadded(t *testing.T) {
	setRawKey(t, "short")
	defer clearKey(t)

	if len(defaultKey) != 32 {
		t.Fatalf("expected key length 32 after padding, got %d", len(defaultKey))
	}

	enc, err := EncryptString("test")
	if err != nil {
		t.Fatalf("EncryptString with short key failed: %v", err)
	}
	dec, err := DecryptString(enc)
	if err != nil {
		t.Fatalf("DecryptString with short key failed: %v", err)
	}
	if dec != "test" {
		t.Fatalf("expected test, got %q", dec)
	}
}

func TestKey_LongKeyTruncated(t *testing.T) {
	setRawKey(t, strings.Repeat("a", 64))
	defer clearKey(t)

	if len(defaultKey) != 32 {
		t.Fatalf("expected key length 32 after truncation, got %d", len(defaultKey))
	}
}
