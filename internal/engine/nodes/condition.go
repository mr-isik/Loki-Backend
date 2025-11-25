package nodes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type ConditionNode struct{}

type conditionData struct {
	Value1   interface{} `json:"value1"`
	Operator string      `json:"operator"`
	Value2   interface{} `json:"value2"`
}

func (n *ConditionNode) Execute(ctx context.Context, rawData []byte) (*domain.NodeResult, error) {
	var data conditionData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	result := false
	switch data.Operator {
	case "==":
		result = data.Value1 == data.Value2
	case "!=":
		result = data.Value1 != data.Value2
	case ">":
		result = compare(data.Value1, data.Value2) > 0
	case "<":
		result = compare(data.Value1, data.Value2) < 0
	case ">=":
		result = compare(data.Value1, data.Value2) >= 0
	case "<=":
		result = compare(data.Value1, data.Value2) <= 0
	default:
		return &domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Unknown operator: %s", data.Operator),
			OutputData: map[string]interface{}{"error": "Unknown operator"},
		}, fmt.Errorf("unknown operator: %s", data.Operator)
	}

	triggeredHandle := "output_false"
	if result {
		triggeredHandle = "output_true"
	}

	return &domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: triggeredHandle,
		Log:             fmt.Sprintf("Condition evaluated to %v", result),
		OutputData: map[string]interface{}{
			"result": result,
		},
	}, nil
}

// compare compares two values. It returns 1 if v1 > v2, -1 if v1 < v2, 0 if equal.
// It tries to convert to float64 for comparison if possible.
func compare(v1, v2 interface{}) int {
	f1, ok1 := toFloat(v1)
	f2, ok2 := toFloat(v2)

	if ok1 && ok2 {
		if f1 > f2 {
			return 1
		}
		if f1 < f2 {
			return -1
		}
		return 0
	}

	// Fallback to string comparison
	s1 := fmt.Sprintf("%v", v1)
	s2 := fmt.Sprintf("%v", v2)

	if s1 > s2 {
		return 1
	}
	if s1 < s2 {
		return -1
	}
	return 0
}

func toFloat(v interface{}) (float64, bool) {
	switch i := v.(type) {
	case float64:
		return i, true
	case float32:
		return float64(i), true
	case int:
		return float64(i), true
	case int64:
		return float64(i), true
	case int32:
		return float64(i), true
	default:
		return 0, false
	}
}
