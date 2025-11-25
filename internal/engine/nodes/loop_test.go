package nodes

import (
	"context"
	"encoding/json"
	"testing"
)

func TestLoopNode_Execute(t *testing.T) {
	node := &LoopNode{}
	ctx := context.Background()

	t.Run("Loop Array", func(t *testing.T) {
		input := map[string]interface{}{
			"items": []interface{}{"a", "b", "c"},
		}
		inputBytes, _ := json.Marshal(input)

		result, err := node.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Status != "completed" {
			t.Errorf("Expected status completed, got %s", result.Status)
		}

		items := result.OutputData["items"].([]interface{})
		if len(items) != 3 {
			t.Errorf("Expected 3 items, got %d", len(items))
		}
	})

	t.Run("Loop JSON String", func(t *testing.T) {
		input := map[string]interface{}{
			"items": `["x", "y"]`,
		}
		inputBytes, _ := json.Marshal(input)

		result, err := node.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		items := result.OutputData["items"].([]interface{})
		if len(items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(items))
		}
	})
}
