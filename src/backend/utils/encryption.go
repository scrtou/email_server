package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"

	"email_server/config"
)

// getEncryptionKey retrieves the encryption key from the global AppConfig.
// It panics if the key is not 32 bytes long, as this is a critical startup error.
func getEncryptionKey() []byte {
	keyString := config.AppConfig.Security.EncryptionKey
	log.Printf("[DEBUG] Using ENCRYPTION_KEY: '%s' (Length: %d)", keyString, len(keyString))
	key := []byte(keyString)
	if len(key) != 32 {
		panic("ENCRYPTION_KEY must be 32 bytes long for AES-256")
	}
	return key
}

// Encrypt 使用AES-GCM加密数据
func Encrypt(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	block, err := aes.NewCipher(getEncryptionKey())
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 使用AES-GCM解密数据
func Decrypt(encryptedData string) ([]byte, error) {
	if encryptedData == "" {
		return nil, nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(getEncryptionKey())
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("密文太短")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open GCM: %w", err)
	}

	return plaintext, nil
}

// EncryptPassword 使用AES加密密码
func EncryptPassword(password string) (string, error) {
	return Encrypt([]byte(password))
}

// DecryptPassword 解密密码
var DecryptPassword = func(encryptedPassword string) (plaintext string, err error) {
	// Add a panic recovery mechanism to prevent crashes from invalid encrypted data
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("decryption failed due to a panic: %v", r)
		}
	}()

	decryptedBytes, err := Decrypt(encryptedPassword)
	if err != nil {
		return "", err
	}
	return string(decryptedBytes), nil
}

// IsEncryptedPassword 检查密码是否是新的加密格式
// 通过尝试解密来判断，如果解密成功则是新格式，否则是bcrypt格式
func IsEncryptedPassword(password string) bool {
	if password == "" {
		return false
	}

	// 尝试解密，如果成功则是新格式
	_, err := DecryptPassword(password)
	return err == nil
}
