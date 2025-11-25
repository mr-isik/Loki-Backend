package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type FileReadNode struct{}

type fileReadData struct {
	Path string `json:"path"`
}

func (n *FileReadNode) Execute(ctx context.Context, rawData []byte) (*domain.NodeResult, error) {
	var data fileReadData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	if data.Path == "" {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        "Path is required",
			OutputData: map[string]interface{}{"error": "Path is required"},
		}, fmt.Errorf("path is required")
	}

	content, err := os.ReadFile(data.Path)
	if err != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to read file: %v", err),
			OutputData:      map[string]interface{}{"error": err.Error()},
		}, nil
	}

	return &domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output_success",
		Log:             fmt.Sprintf("Read %d bytes from %s", len(content), data.Path),
		OutputData: map[string]interface{}{
			"content": string(content),
			"size":    len(content),
		},
	}, nil
}
