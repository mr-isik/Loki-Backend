package nodes

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCodeJsNode_Execute(t *testing.T) {
	node := &CodeJsNode{}
	ctx := context.Background()

	t.Run("Simple Math", func(t *testing.T) {
		input := map[string]interface{}{
			"code": `
				var result = input.a + input.b;
				result;
			`,
			"input": map[string]interface{}{
				"a": 10,
				"b": 20,
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

		if result.OutputData["result"] != int64(30) { // goja exports numbers as int64 or float64
			// Check for float64 as well just in case
			if f, ok := result.OutputData["result"].(float64); !ok || f != 30 {
				t.Errorf("Expected result 30, got %v (%T)", result.OutputData["result"], result.OutputData["result"])
			}
		}
	})

	t.Run("Return Object", func(t *testing.T) {
		input := map[string]interface{}{
			"code": `
				({ message: "hello " + input.name });
			`,
			"input": map[string]interface{}{
				"name": "world",
			},
		}
		inputBytes, _ := json.Marshal(input)

		result, err := node.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.OutputData["message"] != "hello world" {
			t.Errorf("Expected message 'hello world', got %v", result.OutputData["message"])
		}
	})

	t.Run("Syntax Error", func(t *testing.T) {
		input := map[string]interface{}{
			"code": `var x = ;`,
		}
		inputBytes, _ := json.Marshal(input)

		result, err := node.Execute(ctx, inputBytes)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Status != "failed" {
			t.Errorf("Expected status failed, got %s", result.Status)
		}
	})
}
