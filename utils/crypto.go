package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
)

var secretKey []byte

func InitCrypto() {
	key := os.Getenv("ENCRYPTION_KEY")
	if len(key) != 32 {
		log.Panic("[ERROR] ENCRYPTION_KEY must be 32 characters long for AES-256 encryption")
	}
	secretKey = []byte(key)
	log.Println("[INFO] Encryption key initialized successfully")
}

func Encrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		log.Printf("[ERROR] Encrypt: failed to create cipher block: %v\n", err)
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("[ERROR] Encrypt: failed to create GCM block: %v\n", err)
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("[ERROR] Encrypt: failed to generate nonce: %v\n", err)
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func Decrypt(encrypted string) (string, error) {
	cipherData, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		log.Printf("[ERROR] Decrypt: failed to base64 decode: %v\n", err)
		return "", err
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		log.Printf("[ERROR] Decrypt: failed to create cipher block: %v\n", err)
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("[ERROR] Decrypt: failed to create GCM block: %v\n", err)
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(cipherData) < nonceSize {
		log.Println("[ERROR] Decrypt: cipher text too short")
		return "", errors.New("cipher text too short")
	}

	nonce := cipherData[:nonceSize]
	cipherText := cipherData[nonceSize:]

	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		log.Printf("[ERROR] Decrypt: failed to decrypt message: %v\n", err)
		return "", err
	}

	return string(plainText), nil
}

func Hash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}
