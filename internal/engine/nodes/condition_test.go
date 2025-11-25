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
			name: "Equal True",
			data: map[string]interface{}{
				"value1":   10,
				"operator": "==",
				"value2":   10,
			},
			expectedHandle: "output_true",
			expectedResult: true,
		},
		{
			name: "Equal False",
			data: map[string]interface{}{
				"value1":   10,
				"operator": "==",
				"value2":   20,
			},
			expectedHandle: "output_false",
			expectedResult: false,
		},
		{
			name: "Greater Than True",
			data: map[string]interface{}{
				"value1":   20,
				"operator": ">",
				"value2":   10,
			},
			expectedHandle: "output_true",
			expectedResult: true,
		},
		{
			name: "String Comparison",
			data: map[string]interface{}{
				"value1":   "apple",
				"operator": "==",
				"value2":   "apple",
			},
			expectedHandle: "output_true",
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputBytes, _ := json.Marshal(tt.data)
			result, err := node.Execute(ctx, inputBytes)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if result.Status != "completed" {
				t.Errorf("Expected status completed, got %s", result.Status)
			}
			if result.TriggeredHandle != tt.expectedHandle {
				t.Errorf("Expected handle %s, got %s", tt.expectedHandle, result.TriggeredHandle)
			}
			if result.OutputData["result"] != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result.OutputData["result"])
			}
		})
	}
}
