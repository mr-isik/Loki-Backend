package nodes

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestLogNode_Execute(t *testing.T) {
	node := &LogNode{}
	ctx := context.Background()

	input := map[string]interface{}{
		"message": "test log",
		"level":   "info",
	}
	inputBytes, _ := json.Marshal(input)

	result, err := node.Execute(ctx, inputBytes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Status != "completed" {
		t.Errorf("Expected status completed, got %s", result.Status)
	}
}

func TestFileReadWriteNode_Execute(t *testing.T) {
	writeNode := &FileWriteNode{}
	readNode := &FileReadNode{}
	ctx := context.Background()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := "hello world"

	t.Run("Write File", func(t *testing.T) {
		input := map[string]interface{}{
			"path":    filePath,
			"content": content,
		}
		inputBytes, _ := json.Marshal(input)

		result, err := writeNode.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("Expected status completed, got %s", result.Status)
		}
	})

	t.Run("Read File", func(t *testing.T) {
		input := map[string]interface{}{
			"path": filePath,
		}
		inputBytes, _ := json.Marshal(input)

		result, err := readNode.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("Expected status completed, got %s", result.Status)
		}

		if result.OutputData["content"] != content {
			t.Errorf("Expected content %s, got %v", content, result.OutputData["content"])
		}
	})

	t.Run("Read Non-existent File", func(t *testing.T) {
		input := map[string]interface{}{
			"path": filepath.Join(tmpDir, "nonexistent.txt"),
		}
		inputBytes, _ := json.Marshal(input)

		result, err := readNode.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Status != "failed" {
			t.Errorf("Expected status failed, got %s", result.Status)
		}
	})
}
