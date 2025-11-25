package nodes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type SetDataNode struct{}

type setDataData struct {
	Data map[string]interface{} `json:"data"`
}

func (n *SetDataNode) Execute(ctx context.Context, rawData []byte) (domain.NodeResult, error) {
	var data setDataData
	if err := json.Unmarshal(rawData, &data); err != nil {
		// If it's not the expected structure, maybe the rawData IS the data?
		// But let's stick to the structure: {"data": {...}}
		return domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	return domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output",
		Log:             "Data set",
		OutputData:      data.Data,
	}, nil
}
