package utils

import (
	"errors"
	"net"
	"strings"
)

// IsClosedNetworkError checks if the error is due to a closed network connection
func IsClosedNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific closed connection error
	if errors.Is(err, net.ErrClosed) {
		return true
	}

	// Check for "use of closed" message which is common in Go
	if strings.Contains(err.Error(), "use of closed") {
		return true
	}

	// Check for "connection reset" which can happen when the connection is closed
	if strings.Contains(err.Error(), "connection reset") {
		return true
	}

	return false
}
