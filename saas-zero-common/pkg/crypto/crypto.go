package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

var defaultKey []byte

func init() {
	key := os.Getenv("ENCRYPT_KEY")
	if key == "" {
		return
	}
	if len(key) == 64 {
		if k, err := hex.DecodeString(key); err == nil {
			defaultKey = k
		}
	} else {
		k := []byte(key)
		if len(k) > 32 {
			k = k[:32]
		}
		for len(k) < 32 {
			k = append(k, '0')
		}
		defaultKey = k
	}
}

func Encrypt(plaintext []byte) (string, error) {
	if len(defaultKey) == 0 {
		return "", errors.New("missing ENCRYPT_KEY")
	}
	block, err := aes.NewCipher(defaultKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(cipherBase64 string) ([]byte, error) {
	if len(defaultKey) == 0 {
		return nil, errors.New("missing ENCRYPT_KEY")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(cipherBase64)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(defaultKey)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

func EncryptString(plaintext string) (string, error) {
	return Encrypt([]byte(plaintext))
}

func DecryptString(cipherBase64 string) (string, error) {
	bytes, err := Decrypt(cipherBase64)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
