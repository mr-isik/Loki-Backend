package nodes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mr-isik/loki-backend/internal/domain"
	"github.com/mr-isik/loki-backend/internal/engine/docker"
	"github.com/mr-isik/loki-backend/internal/engine/utils"
)

type ConditionNode struct{}

type conditionData struct {
	Expression string                 `json:"expression"`
	Input      map[string]interface{} `json:"input"`
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

	if data.Expression == "" {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             "Expression is required",
			OutputData:      map[string]interface{}{"error": "Expression is required"},
		}, nil
	}

	runner, err := docker.NewRunner()
	if err != nil {
		sanitizedErr := utils.SanitizeError(err)
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to initialize Docker runner: %v", sanitizedErr),
			OutputData:      map[string]interface{}{"error": sanitizedErr},
		}, nil
	}

	inputJSON, _ := json.Marshal(data.Input)
	codeJSON, _ := json.Marshal(data.Expression)

	wrapperScript := fmt.Sprintf(`
const input = %s;
let _result = false;
let _error = null;
try {
    _result = Boolean(eval(%s));
} catch (err) {
    _error = err.message || String(err);
}
process.stdout.write(JSON.stringify({ result: _result, error: _error }));
`, string(inputJSON), string(codeJSON))

	outputStr, err := runner.RunContainer(ctx, docker.RunRequest{
		Image:   "node:alpine",
		Command: []string{"node", "-e", wrapperScript},
	})

	if err != nil {
		sanitizedErr := utils.SanitizeError(err)
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Condition JS Execution Error: %v", sanitizedErr),
			OutputData:      map[string]interface{}{"error": sanitizedErr},
		}, nil
	}

	var parsedOutput struct {
		Result bool    `json:"result"`
		Error  *string `json:"error"`
	}

	if err := json.Unmarshal([]byte(outputStr), &parsedOutput); err != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to parse condition output: %v", err),
			OutputData:      map[string]interface{}{"error": "invalid output from condition engine"},
		}, nil
	}

	if parsedOutput.Error != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("JS Expression Error: %s", *parsedOutput.Error),
			OutputData:      map[string]interface{}{"error": *parsedOutput.Error},
		}, nil
	}

	result := parsedOutput.Result

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
