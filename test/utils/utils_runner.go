// test/utils/utils_runner.go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bOguzhan/NATbypass/internal/utils"
	"github.com/bOguzhan/NATbypass/pkg/networking"
)

func main() {
	fmt.Println("===== Testing Core Utilities =====")

	// Test ID Generation
	testIDGeneration()

	// Test Logger
	testLogger()

	// Test Retry Mechanism
	testRetry()

	// Test System Info
	testSystemInfo()

	// Test Network Utils
	testNetworkUtils()
}

func testIDGeneration() {
	fmt.Println("\n----- Testing ID Generation -----")

	// Generate Peer ID
	peerID, err := utils.GeneratePeerID()
	if err != nil {
		fmt.Printf("❌ Failed to generate peer ID: %v\n", err)
	} else {
		fmt.Printf("✓ Generated Peer ID: %s (length: %d)\n", peerID, len(peerID))
	}

	// Generate Session ID
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		fmt.Printf("❌ Failed to generate session ID: %v\n", err)
	} else {
		fmt.Printf("✓ Generated Session ID: %s (length: %d)\n", sessionID, len(sessionID))
	}

	// Test Validation
	if utils.ValidateID(peerID, 32) {
		fmt.Println("✓ Peer ID passed validation")
	} else {
		fmt.Println("❌ Peer ID failed validation")
	}

	if utils.ValidateID(sessionID, 16) {
		fmt.Println("✓ Session ID passed validation")
	} else {
		fmt.Println("❌ Session ID failed validation")
	}

	// Test invalid ID
	invalidID := "not-hex-1234"
	if !utils.ValidateID(invalidID, len(invalidID)) {
		fmt.Println("✓ Invalid ID correctly rejected")
	} else {
		fmt.Println("❌ Invalid ID wrongly accepted")
	}
}

func testLogger() {
	fmt.Println("\n----- Testing Logger -----")

	// Create logger
	logger := utils.NewLogger("test-component", "debug")

	// Log at different levels
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.WithFields(map[string]interface{}{
		"user":   "test-user",
		"action": "login",
	}).Info("This is a message with fields")

	fmt.Println("✓ Logger test complete")
}

func testRetry() {
	fmt.Println("\n----- Testing Retry Mechanism -----")

	// Create retry config
	config := utils.RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        500 * time.Millisecond,
		BackoffFactor:     2.0,
		TimeoutPerAttempt: 200 * time.Millisecond,
	}

	// Test with a function that always succeeds
	ctx := context.Background()
	err := utils.RetryWithBackoff(ctx, config, func(ctx context.Context) error {
		fmt.Println("✓ Operation succeeded on first attempt")
		return nil
	})

	if err == nil {
		fmt.Println("✓ Retry mechanism handled success case correctly")
	} else {
		fmt.Printf("❌ Retry mechanism failed on success case: %v\n", err)
	}

	// Test with a function that succeeds after retries
	attempts := 0
	err = utils.RetryWithBackoff(ctx, config, func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			fmt.Printf("Attempt %d: Simulating failure\n", attempts)
			return errors.New("simulated failure")
		}
		fmt.Printf("Attempt %d: Success after retries\n", attempts)
		return nil
	})

	if err == nil {
		fmt.Println("✓ Retry mechanism handled retry case correctly")
	} else {
		fmt.Printf("❌ Retry mechanism failed on retry case: %v\n", err)
	}
}

func testSystemInfo() {
	fmt.Println("\n----- Testing System Info -----")

	// Get system info
	info, err := utils.GetSystemInfo()
	if err != nil {
		fmt.Printf("❌ Failed to get system info: %v\n", err)
		return
	}

	fmt.Println("System Information:")
	fmt.Println(info.String())
	fmt.Println("✓ System info retrieved successfully")
}

func testNetworkUtils() {
	fmt.Println("\n----- Testing Network Utils -----")

	// Get local IP
	ip, err := networking.GetLocalIP()
	if err != nil {
		fmt.Printf("❌ Failed to get local IP: %v\n", err)
	} else {
		fmt.Printf("✓ Local IP: %s\n", ip)
	}

	// Check UDP port availability
	port := 12345
	available := networking.CheckUDPPort(port)
	fmt.Printf("✓ UDP port %d availability: %v\n", port, available)

	// Create UDP listener
	fmt.Println("Creating UDP listener on port 12346...")
	conn, err := networking.CreateUDPListener(":12346", 1*time.Second)
	if err != nil {
		fmt.Printf("❌ Failed to create UDP listener: %v\n", err)
	} else {
		fmt.Println("✓ UDP listener created successfully")
		conn.Close()
	}
}
