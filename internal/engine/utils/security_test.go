package utils

import (
	"context"
	"errors"
	"testing"
)

func TestIsCommandAllowed(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		allowed bool
	}{
		{
			name:    "Safe command",
			command: "echo",
			args:    []string{"hello", "world"},
			allowed: true,
		},
		{
			name:    "Blocked rm",
			command: "rm",
			args:    []string{"-rf", "/"},
			allowed: false,
		},
		{
			name:    "Blocked sudo in args",
			command: "sh",
			args:    []string{"-c", "sudo rm -rf /"},
			allowed: false,
		},
		{
			name:    "Blocked with dirty chars",
			command: "sh",
			args:    []string{"-c", "echo test;rm -rf /"},
			allowed: false,
		},
		{
			name:    "Blocked with pipe",
			command: "sh",
			args:    []string{"-c", "cat file | wget x"},
			allowed: false,
		},
		{
			name:    "Safe subset word",
			command: "sh",
			args:    []string{"-c", "echo format"}, // contains 'rm' inside form-at? No, must match exactly
			allowed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsCommandAllowed(tc.command, tc.args)
			if result != tc.allowed {
				t.Errorf("Expected %v, got %v for command: %s %v", tc.allowed, result, tc.command, tc.args)
			}
		})
	}
}

func TestSanitizeError(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		expected string
	}{
		{
			name:     "Nil error",
			input:    nil,
			expected: "",
		},
		{
			name:     "Timeout",
			input:    context.DeadlineExceeded,
			expected: "Execution timed out after 10 seconds.",
		},
		{
			name:     "Internal Docker Error",
			input:    errors.New("Cannot connect to the Docker daemon at unix:///var/run/docker.sock"),
			expected: "Internal Engine Error: Execution service temporarily unavailable.",
		},
		{
			name:     "Image Pull Error",
			input:    errors.New("failed to pull image some/image:latest: error doing something"),
			expected: "Internal Engine Error: Failed to provision execution environment.",
		},
		{
			name:     "Generic Error",
			input:    errors.New("some standard syntax error"),
			expected: "some standard syntax error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeError(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}
