package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type WaitNode struct{}

type waitData struct {
	Duration int    `json:"duration"` // Duration in milliseconds
	Unit     string `json:"unit"`     // "ms", "s", "m", "h"
}

func (n *WaitNode) Execute(ctx context.Context, rawData []byte) (domain.NodeResult, error) {
	var data waitData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	duration := time.Duration(data.Duration) * time.Millisecond
	switch data.Unit {
	case "s":
		duration = time.Duration(data.Duration) * time.Second
	case "m":
		duration = time.Duration(data.Duration) * time.Minute
	case "h":
		duration = time.Duration(data.Duration) * time.Hour
	}

	select {
	case <-time.After(duration):
		return domain.NodeResult{
			Status:          "completed",
			TriggeredHandle: "output",
			Log:             fmt.Sprintf("Waited for %v", duration),
			OutputData:      map[string]interface{}{"waited": true},
		}, nil
	case <-ctx.Done():
		return domain.NodeResult{
			Status:     "cancelled",
			Log:        "Wait cancelled",
			OutputData: map[string]interface{}{"error": "cancelled"},
		}, ctx.Err()
	}
}
