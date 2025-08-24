package auth

import (
	"fmt"
	"os"
)

const apiKeyEnvKey = "GODIG_API_KEY"

func GetServerKey() (string, error) {
	key := os.Getenv(apiKeyEnvKey)
	if key == "" {
		return "", fmt.Errorf("api key must be provided through the %s environment variable", key)
	}
	return key, nil
}
