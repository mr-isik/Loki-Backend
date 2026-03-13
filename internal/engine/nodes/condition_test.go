package nodes

import (
	"context"
	"encoding/json"
	"testing"
)

func TestConditionNode_Execute(t *testing.T) {
	node := &ConditionNode{}
	ctx := context.Background()

	tests := []struct {
		name           string
		data           map[string]interface{}
		expectedHandle string
		expectedResult bool
	}{
		{
			name: "Simple Boolean True",
			data: map[string]interface{}{
				"expression": "10 == 10",
			},
			expectedHandle: "output_true",
			expectedResult: true,
		},
		{
			name: "Simple Boolean False",
			data: map[string]interface{}{
				"expression": "10 == 20",
			},
			expectedHandle: "output_false",
			expectedResult: false,
		},
		{
			name: "With Input Variables True",
			data: map[string]interface{}{
				"expression": "input.status === 'success' && input.retryCount < 3",
				"input": map[string]interface{}{
					"status":     "success",
					"retryCount": 1,
				},
			},
			expectedHandle: "output_true",
			expectedResult: true,
		},
		{
			name: "With Input Variables False",
			data: map[string]interface{}{
				"expression": "input.status === 'success' && input.retryCount < 3",
				"input": map[string]interface{}{
					"status":     "success",
					"retryCount": 5,
				},
			},
			expectedHandle: "output_false",
			expectedResult: false,
		},
		{
			name: "Invalid expression",
			data: map[string]interface{}{
				"expression": "+++",
			},
			expectedHandle: "output_error",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputBytes, _ := json.Marshal(tt.data)
			result, err := node.Execute(ctx, inputBytes)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if result.Status != "completed" && result.Status != "failed" {
				t.Errorf("Expected status completed or failed, got %s", result.Status)
			}
			if result.TriggeredHandle != tt.expectedHandle {
				t.Errorf("Expected handle %s, got %s", tt.expectedHandle, result.TriggeredHandle)
			}
			if result.OutputData["result"] != nil && result.OutputData["result"] != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result.OutputData["result"])
			}
		})
	}
}
