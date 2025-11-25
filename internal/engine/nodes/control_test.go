package nodes

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestWaitNode_Execute(t *testing.T) {
	node := &WaitNode{}
	ctx := context.Background()

	start := time.Now()
	input := map[string]interface{}{
		"duration": 100,
		"unit":     "ms",
	}
	inputBytes, _ := json.Marshal(input)

	result, err := node.Execute(ctx, inputBytes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	elapsed := time.Since(start)

	if result.Status != "completed" {
		t.Errorf("Expected status completed, got %s", result.Status)
	}
	if elapsed < 100*time.Millisecond {
		t.Errorf("Expected to wait at least 100ms, waited %v", elapsed)
	}
}

func TestMergeNode_Execute(t *testing.T) {
	node := &MergeNode{}
	ctx := context.Background()

	result, err := node.Execute(ctx, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Status != "completed" {
		t.Errorf("Expected status completed, got %s", result.Status)
	}
}

func TestSetDataNode_Execute(t *testing.T) {
	node := &SetDataNode{}
	ctx := context.Background()

	input := map[string]interface{}{
		"data": map[string]interface{}{
			"foo": "bar",
			"num": 123,
		},
	}
	inputBytes, _ := json.Marshal(input)

	result, err := node.Execute(ctx, inputBytes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Status != "completed" {
		t.Errorf("Expected status completed, got %s", result.Status)
	}

	if result.OutputData["foo"] != "bar" {
		t.Errorf("Expected foo=bar, got %v", result.OutputData["foo"])
	}
	// JSON numbers are float64
	if result.OutputData["num"] != 123.0 {
		t.Errorf("Expected num=123, got %v", result.OutputData["num"])
	}
}
