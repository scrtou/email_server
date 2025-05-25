package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLength = 6
	MaxPasswordLength = 128
)

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	if len(password) < MinPasswordLength {
		return "", errors.New("密码长度不能少于6位")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return errors.New("密码长度不能少于6位")
	}

	if len(password) > MaxPasswordLength {
		return errors.New("密码长度不能超过128位")
	}

	// 可以添加更多密码强度验证规则
	return nil
}
