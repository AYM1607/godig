package auth

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
)

func GenerateToken() (string, error) {
	return GenerateString(32)
}

func GenerateString(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("token length must be greater than 0")
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	token := base32.StdEncoding.EncodeToString(bytes)
	return token, nil
}
