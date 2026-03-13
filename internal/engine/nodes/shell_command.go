package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mr-isik/loki-backend/internal/domain"
	"github.com/mr-isik/loki-backend/internal/engine/docker"
)

type ShellCommandNode struct{}

type shellData struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Dir     string   `json:"dir"`
}

func (n *ShellCommandNode) Execute(ctx context.Context, rawData []byte) (*domain.NodeResult, error) {
	var data shellData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	if data.Command == "" {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        "Command is required",
			OutputData: map[string]interface{}{"error": "Command is required"},
		}, fmt.Errorf("command is required")
	}

	runner, err := docker.NewRunner()
	if err != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to initialize Docker runner: %v", err),
			OutputData: map[string]interface{}{
				"error": err.Error(),
			},
		}, nil
	}

	commandArgs := []string{data.Command}
	commandArgs = append(commandArgs, data.Args...)

	outputStr, err := runner.RunContainer(ctx, docker.RunRequest{
		Image:   "alpine:latest",
		Command: commandArgs,
	})

	if err != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Command failed: %v\nOutput: %s", err, outputStr),
			OutputData: map[string]interface{}{
				"error":  err.Error(),
				"output": outputStr,
			},
		}, nil
	}

	return &domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output_success",
		Log:             fmt.Sprintf("Command executed successfully. Output length: %d", len(outputStr)),
		OutputData: map[string]interface{}{
			"output": strings.TrimSpace(outputStr),
		},
	}, nil
}
