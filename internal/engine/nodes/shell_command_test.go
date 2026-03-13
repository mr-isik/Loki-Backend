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
		input := map[string]interface{}{
			"command": "sh",
			"args":    []string{"-c", "echo hello world"},
		}

		inputBytes, _ := json.Marshal(input)

		result, err := node.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Status != "completed" {
			t.Fatalf("Expected status completed, got %s. Log: %s", result.Status, result.Log)
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

	t.Run("Blacklisted Command", func(t *testing.T) {
		input := map[string]interface{}{
			"command": "rm",
			"args":    []string{"-rf", "/"},
		}

		inputBytes, _ := json.Marshal(input)

		result, err := node.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Status != "failed" {
			t.Fatalf("Expected status failed, got %s. Log: %s", result.Status, result.Log)
		}

		errStr := result.OutputData["error"].(string)
		if errStr != "Command rejected due to security policy constraints." {
			t.Errorf("Expected security policy error, got '%s'", errStr)
		}
	})
}
