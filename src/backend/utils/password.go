package utils

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLength = 6
	MaxPasswordLength = 128
)

// HashPassword now exclusively uses AES encryption.
// The name is kept for backward compatibility in some parts of the code,
// but its function is now to encrypt, not to hash.
func HashPassword(password string) (string, error) {
	if len(password) < MinPasswordLength {
		return "", errors.New("密码长度不能少于6位")
	}
	return EncryptPassword(password)
}

// CheckPassword intelligently checks a password against either a bcrypt hash or an AES encrypted string.
func CheckPassword(password, storedPassword string) bool {
	if strings.HasPrefix(storedPassword, "$2a$") || strings.HasPrefix(storedPassword, "$2b$") || strings.HasPrefix(storedPassword, "$2y$") {
		// It's a bcrypt hash
		err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
		return err == nil
	}
	
	// Assume it's an AES encrypted string
	decryptedPassword, err := DecryptPassword(storedPassword)
	if err != nil {
		return false // Decryption failed, so it's not a valid password
	}
	return password == decryptedPassword
}

// ValidatePassword validates the password strength.
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return errors.New("密码长度不能少于6位")
	}

	if len(password) > MaxPasswordLength {
		return errors.New("密码长度不能超过128位")
	}

	// More password strength rules can be added here.
	return nil
}
