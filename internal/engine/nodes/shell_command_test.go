package nodes

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestShellCommandNode_Execute(t *testing.T) {
	node := &ShellCommandNode{}
	ctx := context.Background()

	t.Run("Echo Command", func(t *testing.T) {
		// On Windows, "echo" is a shell builtin, so we need "cmd /c echo" or similar.
		// However, exec.Command often handles this or we can use a simple command like "whoami" or "hostname" if echo fails.
		// Let's try "cmd /c echo hello world" for Windows compatibility if needed, but "echo" might not work directly with exec.Command on Windows.
		// Better to use "ping" or something standard, or just "cmd /c echo".

		input := map[string]interface{}{
			"command": "cmd",
			"args":    []string{"/c", "echo", "hello world"},
		}
		// Fallback for non-windows (linux/mac) would be just "echo"
		// But since user is on Windows, "cmd /c" is safe.

		inputBytes, _ := json.Marshal(input)

		result, err := node.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("Expected status completed, got %s", result.Status)
		}

		output := result.OutputData["output"].(string)
		if !strings.Contains(output, "hello world") {
			t.Errorf("Expected output containing 'hello world', got '%s'", output)
		}
	})

	t.Run("Invalid Command", func(t *testing.T) {
		input := map[string]interface{}{
			"command": "nonexistentcommand12345",
		}
		inputBytes, _ := json.Marshal(input)

		result, err := node.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Status != "failed" {
			t.Errorf("Expected status failed, got %s", result.Status)
		}
		if result.TriggeredHandle != "output_error" {
			t.Errorf("Expected handle output_error, got %s", result.TriggeredHandle)
		}
	})
}
