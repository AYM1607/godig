package auth

import (
	"encoding/base32"
	"fmt"
	"strings"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	// Verify it's valid base32
	_, err = base32.StdEncoding.DecodeString(token)
	if err != nil {
		t.Errorf("generated token is not valid base32: %v", err)
	}

	// Verify the decoded length is 64 bytes (default)
	decoded, err := base32.StdEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("failed to decode token: %v", err)
	}

	if len(decoded) != 32 {
		t.Errorf("expected decoded length 64, got %d", len(decoded))
	}
}

func TestGenerateString(t *testing.T) {
	tests := []struct {
		name         string
		length       int
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid length 16",
			length:      16,
			expectError: false,
		},
		{
			name:        "valid length 32",
			length:      32,
			expectError: false,
		},
		{
			name:        "valid length 1",
			length:      1,
			expectError: false,
		},
		{
			name:         "zero length",
			length:       0,
			expectError:  true,
			errorMessage: "token length must be greater than 0",
		},
		{
			name:         "negative length",
			length:       -1,
			expectError:  true,
			errorMessage: "token length must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateString(tt.length)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("expected error message to contain %q, got %q", tt.errorMessage, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if token == "" {
				t.Error("expected non-empty token")
			}

			// Verify it's valid base32
			_, err = base32.StdEncoding.DecodeString(token)
			if err != nil {
				t.Errorf("generated token is not valid base32: %v", err)
			}
		})
	}
}

func TestGenerateString_ProperLength(t *testing.T) {
	tests := []int{1, 8, 16, 32, 64}

	for _, length := range tests {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			token, err := GenerateString(length)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Decode to verify the original byte length
			decoded, err := base32.StdEncoding.DecodeString(token)
			if err != nil {
				t.Fatalf("failed to decode token: %v", err)
			}

			if len(decoded) != length {
				t.Errorf("expected decoded length %d, got %d", length, len(decoded))
			}
		})
	}
}
