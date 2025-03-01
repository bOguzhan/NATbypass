// internal/utils/sysinfo.go
package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// SystemInfo holds basic information about the system environment
type SystemInfo struct {
	Hostname     string            `json:"hostname"`
	Platform     string            `json:"platform"`
	Architecture string            `json:"architecture"`
	NumCPU       int               `json:"num_cpu"`
	GoVersion    string            `json:"go_version"`
	StartTime    time.Time         `json:"start_time"`
	Environment  map[string]string `json:"environment"`
}

// GetSystemInfo returns information about the current system
func GetSystemInfo() (*SystemInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get relevant environment variables
	envVars := make(map[string]string)
	for _, env := range []string{"PATH", "GOPATH", "GOROOT", "HOME", "USER", "LOGNAME"} {
		if value, exists := os.LookupEnv(env); exists {
			envVars[env] = value
		}
	}

	return &SystemInfo{
		Hostname:     hostname,
		Platform:     runtime.GOOS,
		Architecture: runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		GoVersion:    runtime.Version(),
		StartTime:    time.Now(),
		Environment:  envVars,
	}, nil
}

// String returns a human-readable representation of the system info
func (s *SystemInfo) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Hostname:     %s\n", s.Hostname))
	sb.WriteString(fmt.Sprintf("Platform:     %s\n", s.Platform))
	sb.WriteString(fmt.Sprintf("Architecture: %s\n", s.Architecture))
	sb.WriteString(fmt.Sprintf("CPUs:         %d\n", s.NumCPU))
	sb.WriteString(fmt.Sprintf("Go Version:   %s\n", s.GoVersion))
	sb.WriteString(fmt.Sprintf("Start Time:   %s\n", s.StartTime.Format(time.RFC3339)))

	return sb.String()
}
