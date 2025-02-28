// internal/utils/idgen_test.go
package utils

import (
	"encoding/hex"
	"testing"
)

func TestGenerateRandomID(t *testing.T) {
	// Test with different lengths
	lengths := []int{4, 8, 16, 32}

	for _, length := range lengths {
		// Generate ID with specific byte length
		id, err := GenerateRandomID(length)

		// Check for errors
		if err != nil {
			t.Fatalf("Failed to generate random ID: %v", err)
		}

		// Check length (hex encoding means 2 chars per byte)
		expectedHexLength := length * 2
		if len(id) != expectedHexLength {
			t.Errorf("Expected ID length to be %d, got %d", expectedHexLength, len(id))
		}

		// Check that it's valid hex
		_, err = hex.DecodeString(id)
		if err != nil {
			t.Errorf("Generated ID is not valid hex: %v", err)
		}
	}

	// Generate two IDs and ensure they are different
	id1, _ := GenerateRandomID(16)
	id2, _ := GenerateRandomID(16)

	if id1 == id2 {
		t.Errorf("Generated IDs should be random, but got the same value twice: %s", id1)
	}
}

func TestGeneratePeerID(t *testing.T) {
	// Generate peer ID
	peerID, err := GeneratePeerID()

	// Check for errors
	if err != nil {
		t.Fatalf("Failed to generate peer ID: %v", err)
	}

	// Check length (should be 16 bytes = 32 hex chars)
	if len(peerID) != 32 {
		t.Errorf("Expected peer ID length to be 32, got %d", len(peerID))
	}
}

func TestGenerateSessionID(t *testing.T) {
	// Generate session ID
	sessionID, err := GenerateSessionID()

	// Check for errors
	if err != nil {
		t.Fatalf("Failed to generate session ID: %v", err)
	}

	// Check length (should be 8 bytes = 16 hex chars)
	if len(sessionID) != 16 {
		t.Errorf("Expected session ID length to be 16, got %d", len(sessionID))
	}
}

func TestValidateID(t *testing.T) {
	// Test valid IDs
	validID := "0123456789abcdef"
	if !ValidateID(validID, 16) {
		t.Errorf("Expected '%s' to be a valid ID of length 16, but it was rejected", validID)
	}

	// Test invalid length
	if ValidateID(validID, 14) {
		t.Errorf("Expected '%s' to be rejected as invalid length, but it was accepted", validID)
	}

	// Test invalid characters
	invalidID := "0123456789abcdefg" // contains 'g' which is not hex
	if ValidateID(invalidID, 17) {
		t.Errorf("Expected '%s' to be rejected as invalid hex, but it was accepted", invalidID)
	}
}
