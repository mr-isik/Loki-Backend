package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type FileWriteNode struct{}

type fileWriteData struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Append  bool   `json:"append"`
}

func (n *FileWriteNode) Execute(ctx context.Context, rawData []byte) (domain.NodeResult, error) {
	var data fileWriteData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	if data.Path == "" {
		return domain.NodeResult{
			Status:     "failed",
			Log:        "Path is required",
			OutputData: map[string]interface{}{"error": "Path is required"},
		}, fmt.Errorf("path is required")
	}

	// Ensure directory exists
	dir := filepath.Dir(data.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to create directory: %v", err),
			OutputData:      map[string]interface{}{"error": err.Error()},
		}, nil
	}

	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if data.Append {
		flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}

	file, err := os.OpenFile(data.Path, flags, 0644)
	if err != nil {
		return domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to open file: %v", err),
			OutputData:      map[string]interface{}{"error": err.Error()},
		}, nil
	}
	defer file.Close()

	nBytes, err := file.WriteString(data.Content)
	if err != nil {
		return domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to write to file: %v", err),
			OutputData:      map[string]interface{}{"error": err.Error()},
		}, nil
	}

	return domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output_success",
		Log:             fmt.Sprintf("Wrote %d bytes to %s", nBytes, data.Path),
		OutputData:      map[string]interface{}{"bytes_written": nBytes},
	}, nil
}
