// internal/utils/idgen.go
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateRandomID creates a random ID with the specified byte length
// Returns a hex-encoded string representation
func GenerateRandomID(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GeneratePeerID creates a random peer ID with standard length (16 bytes -> 32 hex chars)
func GeneratePeerID() (string, error) {
	return GenerateRandomID(16)
}

// GenerateSessionID creates a random session ID with standard length (8 bytes -> 16 hex chars)
func GenerateSessionID() (string, error) {
	return GenerateRandomID(8)
}

// ValidateID checks if the provided ID is a valid hex-encoded ID of the expected length
func ValidateID(id string, expectedHexLength int) bool {
	// Check if length matches
	if len(id) != expectedHexLength {
		return false
	}

	// Check if it's valid hex
	_, err := hex.DecodeString(id)
	return err == nil
}
