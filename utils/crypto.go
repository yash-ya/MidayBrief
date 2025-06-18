package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

var secretKey []byte

func InitCrypto() {
	key := os.Getenv("ENCRYPTION_KEY")
	if len(key) != 32 {
		panic("ENCRYPTION_SECRET must be 32 bytes long (AES-256)")
	}
	secretKey = []byte(key)
}

func Encrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(secretKey)
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

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func Decrypt(encrypted string) (string, error) {
	cipherData, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(cipherData) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce := cipherData[:nonceSize]
	cipherText := cipherData[nonceSize:]

	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}