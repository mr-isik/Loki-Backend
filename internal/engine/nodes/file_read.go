package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mr-isik/loki-backend/internal/domain"
	"github.com/mr-isik/loki-backend/internal/engine/utils"
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

	if err := utils.ValidateFilePath(data.Path); err != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Security violation: %v", err),
			OutputData:      map[string]interface{}{"error": err.Error()},
		}, nil
	}

	fileInfo, err := os.Stat(data.Path)
	if err != nil {
		sanitizedErr := utils.SanitizeError(err)
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to stat file: %s", sanitizedErr),
			OutputData:      map[string]interface{}{"error": "File not found or inaccessible"},
		}, nil
	}

	if fileInfo.Size() > utils.MaxFileSize {
		errMsg := "File size exceeds the 5MB limit"
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             errMsg,
			OutputData:      map[string]interface{}{"error": errMsg},
		}, nil
	}

	file, err := os.Open(data.Path)
	if err != nil {
		sanitizedErr := utils.SanitizeError(err)
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to open file: %s", sanitizedErr),
			OutputData:      map[string]interface{}{"error": "Failed to open file"},
		}, nil
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, utils.MaxFileSize))
	if err != nil {
		sanitizedErr := utils.SanitizeError(err)
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to read file: %s", sanitizedErr),
			OutputData:      map[string]interface{}{"error": "Failed to read file"},
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
