package utils

import (
	"context"
	"errors"
	"strings"
)

var blacklistedCommands = []string{
	"rm", "sudo", "chmod", "shutdown", "chown", "mkfs", "reboot", "halt", "poweroff",
	"mv", "cp", "ln", "dd", "mount", "umount", "wget", "curl", "nc", "ncat", "netcat", "telnet",
}

// IsCommandAllowed checks if the shell command or its arguments contain a blacklisted command.
func IsCommandAllowed(command string, args []string) bool {
	check := func(s string) bool {
		s = strings.ReplaceAll(s, ";", " ")
		s = strings.ReplaceAll(s, "&", " ")
		s = strings.ReplaceAll(s, "|", " ")
		s = strings.ReplaceAll(s, ">", " ")
		s = strings.ReplaceAll(s, "<", " ")
		s = strings.ReplaceAll(s, "(", " ")
		s = strings.ReplaceAll(s, ")", " ")
		s = strings.ReplaceAll(s, "`", " ")
		s = strings.ReplaceAll(s, "\"", " ")
		s = strings.ReplaceAll(s, "'", " ")

		tokens := strings.Fields(s)
		for _, token := range tokens {
			for _, bc := range blacklistedCommands {
				if token == bc {
					return false
				}
			}
		}
		return true
	}

	if !check(command) {
		return false
	}
	for _, arg := range args {
		if !check(arg) {
			return false
		}
	}
	return true
}

// SanitizeError formats timeout Context errors and hides sensitive Docker/internal errors.
// It returns a safe error string for user consumption.
func SanitizeError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "Execution timed out after 10 seconds."
	}

	errMsg := err.Error()

	// Hide Docker daemon related errors
	if strings.Contains(errMsg, "Cannot connect to the Docker daemon") ||
		strings.Contains(errMsg, "failed to initialize docker client") {
		return "Internal Engine Error: Execution service temporarily unavailable."
	}
	if strings.Contains(errMsg, "failed to pull image") ||
		strings.Contains(errMsg, "failed to create container") ||
		strings.Contains(errMsg, "failed to start container") ||
		strings.Contains(errMsg, "failed to attach to container") {
		return "Internal Engine Error: Failed to provision execution environment."
	}

	return errMsg
}
