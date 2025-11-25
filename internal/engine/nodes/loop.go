package nodes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type LoopNode struct{}

type loopData struct {
	Items interface{} `json:"items"`
}

func (n *LoopNode) Execute(ctx context.Context, rawData []byte) (*domain.NodeResult, error) {
	var data loopData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	// The LoopNode in this architecture is a bit tricky.
	// Usually, a loop node in a workflow engine either:
	// 1. Emits multiple events (one for each item) - this requires the engine to handle multiple triggers.
	// 2. Returns a list, and the next node handles the list.
	// 3. Is a "start" of a loop block, and the engine iterates.

	// Based on the db.go definition:
	// Inputs: [{"id": "input", "label": "Start"}]
	// Outputs: [{"id": "output_item", "label": "For Each Item"}, {"id": "output_done", "label": "Done"}]

	// This suggests the engine handles the iteration. The node itself just prepares the items.
	// However, `Execute` returns a SINGLE `NodeResult`.
	// If the engine expects the node to manage state, we might need to return the list.
	// Let's assume the engine handles the "output_item" handle repeatedly if we return a list,
	// OR the engine handles the iteration logic itself based on the node type.

	// BUT, since we are implementing `Execute` which returns `NodeResult`, we can't easily "loop" here without engine support.
	// A common pattern in simple engines is:
	// The LoopNode returns the list of items in `OutputData`.
	// The ENGINE sees "output_item" handle and the list, and spawns execution for each item.

	// Let's normalize the input to a slice.
	items, err := toSlice(data.Items)
	if err != nil {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to convert items to slice: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	return &domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output_item", // The engine should probably handle this special case
		Log:             fmt.Sprintf("Looping over %d items", len(items)),
		OutputData: map[string]interface{}{
			"items": items,
		},
	}, nil
}

func toSlice(v interface{}) ([]interface{}, error) {
	if v == nil {
		return []interface{}{}, nil
	}

	switch val := v.(type) {
	case []interface{}:
		return val, nil
	case []string:
		res := make([]interface{}, len(val))
		for i, v := range val {
			res[i] = v
		}
		return res, nil
	case []int:
		res := make([]interface{}, len(val))
		for i, v := range val {
			res[i] = v
		}
		return res, nil
	// Add more slice types as needed
	default:
		// Try to parse if it's a string JSON
		if s, ok := v.(string); ok {
			var res []interface{}
			if err := json.Unmarshal([]byte(s), &res); err == nil {
				return res, nil
			}
		}
		return nil, fmt.Errorf("input is not a slice or valid JSON array")
	}
}
