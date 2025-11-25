package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type LogNode struct{}

type logData struct {
	Message string `json:"message"`
	Level   string `json:"level"` // info, warn, error
}

func (n *LogNode) Execute(ctx context.Context, rawData []byte) (domain.NodeResult, error) {
	var data logData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	// In a real system, this might log to a structured logger or database.
	// For now, we use standard log.
	logMsg := fmt.Sprintf("[%s] %s", data.Level, data.Message)
	log.Println(logMsg)

	return domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output",
		Log:             logMsg,
		OutputData:      map[string]interface{}{"logged": true},
	}, nil
}
